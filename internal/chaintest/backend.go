// Package chaintest 给单测用的假链：实现 bind.ContractBackend 的只读部分（CodeAt / CallContract），
// 把 abigen 绑定发来的 eth_call 按"目标地址 + 4 字节选择器"分发到内存里的假状态，返回值再按 ABI 打包回去。
// 生产代码里的 c.Dealer.Xxx / NewPerpetual(…, c.Backend) 原样调用，不用为测试抽接口。
// 只覆盖 eth_call；发交易的路径（tick、liquidate）在 anvil 上端到端验证。
package chaintest

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
	"github.com/xuefei016/my-perpetual-backend/internal/chain/bindings/dealer"
	"github.com/xuefei016/my-perpetual-backend/internal/chain/bindings/perpetual"
	"github.com/xuefei016/my-perpetual-backend/internal/chain/bindings/pricesource"
)

// ErrNodeDown 模拟节点挂了。
var ErrNodeDown = errors.New("node down")

// Account 一个交易者在假链上的状态（单位同合约：paper 1e18，其余 USDC 6 位）。
// Paper / Credit 只对 Backend.BTC 这个市场生效。
type Account struct {
	Paper, Credit     *big.Int // Perpetual.balanceOf
	NetValue          *big.Int // Dealer.getTraderRisk
	InitialMargin     *big.Int
	MaintenanceMargin *big.Int
	PrimaryCredit     *big.Int
	Safe              bool
}

// Rich 没仓位、钱很多：IM 预检一定通过，也一定 safe。
func Rich() *Account {
	return &Account{
		Paper: new(big.Int), Credit: new(big.Int),
		NetValue: big.NewInt(1e15), InitialMargin: new(big.Int), MaintenanceMargin: new(big.Int),
		PrimaryCredit: big.NewInt(1e15), Safe: true,
	}
}

// LiqCostFunc 模拟 Dealer.getLiquidationCost(perp, trader, requestPaper)。
type LiqCostFunc func(perp, trader common.Address, requestPaper *big.Int) (liqtorPaper, liqtorCredit *big.Int, err error)

type Backend struct {
	bind.ContractBackend // 只为满足接口；用到没实现的方法会 nil panic，提醒测试漏了

	mu                           sync.Mutex
	dealerABI, perpABI, priceABI *abi.ABI

	Dealer, BTC, PriceSrc common.Address
	Perps                 map[common.Address]bool // 已注册的市场（默认只有 BTC），都用 PriceSrc 当价格源

	Operators    map[[2]common.Address]bool // (client, operator) → 是否授权
	Filled       map[common.Hash]*big.Int   // orderHash → 链上已成交 paper
	Accounts     map[common.Address]*Account
	FundingRates map[common.Address]*big.Int // perp → 累计资金费率
	Mark         *big.Int                    // 标记价（USDC 6 位 / 1 张）
	IMR          *big.Int                    // initialMarginRatio（1e18）
	LiqCost      LiqCostFunc

	CallErr  error           // 非 nil 时所有 eth_call 都失败
	FailOnly map[string]bool // 只让这些方法失败（返回 ErrNodeDown）
	Calls    []string        // 记录被调用的方法名
}

// New 默认状态：标记价 26000、IM 10%、BTC 已注册、所有人都是 Rich。
func New(dealerAddr, btc, priceSrc common.Address) (*Backend, error) {
	d, err := dealer.DealerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	p, err := perpetual.PerpetualMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	ps, err := pricesource.PriceSourceMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return &Backend{
		dealerABI: d, perpABI: p, priceABI: ps,
		Dealer: dealerAddr, BTC: btc, PriceSrc: priceSrc,
		Perps:        map[common.Address]bool{btc: true},
		Operators:    map[[2]common.Address]bool{},
		Filled:       map[common.Hash]*big.Int{},
		Accounts:     map[common.Address]*Account{},
		FundingRates: map[common.Address]*big.Int{},
		Mark:         big.NewInt(26000e6),
		IMR:          big.NewInt(1e17),
		FailOnly:     map[string]bool{},
	}, nil
}

// Chain 用这个假后端拼一个 *chain.Chain（Client 为 nil，只能做 eth_call）。
func (f *Backend) Chain(chainID int64) (*chain.Chain, error) {
	d, err := dealer.NewDealer(f.Dealer, f)
	if err != nil {
		return nil, err
	}
	p, err := perpetual.NewPerpetual(f.BTC, f)
	if err != nil {
		return nil, err
	}
	return &chain.Chain{Backend: f, ChainId: big.NewInt(chainID), Dealer: d, BTC: p}, nil
}

func (f *Backend) account(a common.Address) *Account {
	if acc, ok := f.Accounts[a]; ok {
		return acc
	}
	return Rich()
}

func (f *Backend) CodeAt(context.Context, common.Address, *big.Int) ([]byte, error) {
	return []byte{1}, nil
}

func (f *Backend) CallContract(_ context.Context, call ethereum.CallMsg, _ *big.Int) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.CallErr != nil {
		return nil, f.CallErr
	}
	to := *call.To
	var a *abi.ABI
	switch {
	case to == f.Dealer:
		a = f.dealerABI
	case f.Perps[to]:
		a = f.perpABI
	case to == f.PriceSrc:
		a = f.priceABI
	default:
		return nil, fmt.Errorf("fake: no contract at %s", to)
	}
	m, err := a.MethodById(call.Data[:4])
	if err != nil {
		return nil, fmt.Errorf("fake: unknown selector %x", call.Data[:4])
	}
	args, err := m.Inputs.Unpack(call.Data[4:])
	if err != nil {
		return nil, err
	}
	f.Calls = append(f.Calls, m.Name)
	if f.FailOnly[m.Name] {
		return nil, ErrNodeDown
	}
	switch m.Name {
	case "isOperatorValid":
		return m.Outputs.Pack(f.Operators[[2]common.Address{args[0].(common.Address), args[1].(common.Address)}])
	case "getOrderFilledAmount":
		v := f.Filled[common.Hash(args[0].([32]byte))]
		if v == nil {
			v = new(big.Int)
		}
		return m.Outputs.Pack(v)
	case "getTraderRisk":
		acc := f.account(args[0].(common.Address))
		exposure := new(big.Int) // 实现不该用到，给 0
		return m.Outputs.Pack(acc.NetValue, exposure, acc.InitialMargin, acc.MaintenanceMargin)
	case "getRiskParams":
		if !f.Perps[args[0].(common.Address)] {
			return nil, fmt.Errorf("fake: risk params of unknown perp %s", args[0])
		}
		return m.Outputs.Pack(dealer.TypesRiskParams{
			InitialMarginRatio: f.IMR, LiquidationThreshold: big.NewInt(5e16),
			LiquidationPriceOff: big.NewInt(1e16), InsuranceFeeRate: big.NewInt(1e16),
			MarkPriceSource: f.PriceSrc, Name: "BTC-USDC", IsRegistered: true,
		})
	case "getMarkPrice":
		return m.Outputs.Pack(f.Mark)
	case "getFundingRate":
		v := f.FundingRates[to]
		if v == nil {
			v = new(big.Int)
		}
		return m.Outputs.Pack(new(big.Int).Set(v))
	case "balanceOf":
		acc := f.account(args[0].(common.Address))
		if to != f.BTC {
			return m.Outputs.Pack(new(big.Int), new(big.Int))
		}
		return m.Outputs.Pack(acc.Paper, acc.Credit)
	case "getPrimaryCredit":
		return m.Outputs.Pack(f.account(args[0].(common.Address)).PrimaryCredit)
	case "isSafe":
		return m.Outputs.Pack(f.account(args[0].(common.Address)).Safe)
	case "getLiquidationCost":
		if f.LiqCost == nil {
			return nil, errors.New("fake: LiqCost not set")
		}
		p, c, err := f.LiqCost(args[0].(common.Address), args[1].(common.Address), args[2].(*big.Int))
		if err != nil {
			return nil, err
		}
		return m.Outputs.Pack(p, c)
	}
	return nil, fmt.Errorf("fake: method %s not faked", m.Name)
}

// Called 是否调用过某个方法。
func (f *Backend) Called(name string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Contains(f.Calls, name)
}
