package keeper

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/xuefei016/my-perpetual-backend/internal/chain/bindings/dealer"
	"github.com/xuefei016/my-perpetual-backend/internal/chain/bindings/perpetual"
	"github.com/xuefei016/my-perpetual-backend/internal/chaintest"
)

var (
	liqTrader = common.HexToAddress("0x00000000000000000000000000000000000a11ce")
	liqOwner  = common.HexToAddress("0x0000000000000000000000000000000000000b07")
	errLiqGas = errors.New("gas estimation rejected")
)

type liquidationTraceKey struct{}

func assertLiquidationChildContext(t *testing.T, parent, child context.Context, limit time.Duration) {
	t.Helper()
	if child.Value(liquidationTraceKey{}) != parent.Value(liquidationTraceKey{}) {
		t.Error("child context lost parent values")
	}
	deadline, ok := child.Deadline()
	if !ok || time.Until(deadline) > limit {
		t.Errorf("child deadline = %v (present %v), want at most %s", deadline, ok, limit)
	}
}

// 拦截真实绑定的 gas 估算，检查编码参数后返回错误，不广播交易。
type liquidationBackend struct {
	bind.ContractBackend
	readContexts []context.Context
	estimate     func(context.Context, ethereum.CallMsg) (uint64, error)
}

func (b *liquidationBackend) CallContract(ctx context.Context, call ethereum.CallMsg, block *big.Int) ([]byte, error) {
	b.readContexts = append(b.readContexts, ctx)
	return b.ContractBackend.CallContract(ctx, call, block)
}

func (b *liquidationBackend) PendingCodeAt(context.Context, common.Address) ([]byte, error) {
	return []byte{1}, nil
}

func (b *liquidationBackend) EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error) {
	return b.estimate(ctx, call)
}

func newLiquidationEnv(t *testing.T) (*Liquidator, *chaintest.Backend, *liquidationBackend) {
	t.Helper()
	fb, err := chaintest.New(testDealer, testETH, testPriceSrc)
	if err != nil {
		t.Fatal(err)
	}
	c, err := fb.Chain(31337)
	if err != nil {
		t.Fatal(err)
	}
	// c.BTC 故意绑定另一个地址，防止实现忽略 l.perp。
	c.BTC, err = perpetual.NewPerpetual(testBTC, fb)
	if err != nil {
		t.Fatal(err)
	}
	b := &liquidationBackend{ContractBackend: fb, estimate: func(context.Context, ethereum.CallMsg) (uint64, error) {
		t.Fatal("unexpected transaction attempt")
		return 0, errLiqGas
	}}
	c.Backend = b
	c.Dealer, err = dealer.NewDealer(testDealer, b)
	if err != nil {
		t.Fatal(err)
	}
	auth := &bind.TransactOpts{From: liqOwner, GasPrice: big.NewInt(1), Nonce: new(big.Int), Context: context.Background()}
	return NewLiquidator(c, auth, testETH, nil, time.Second), fb, b
}

func TestLiquidateZeroPositionSkipsQuoteAndTransaction(t *testing.T) {
	l, fb, _ := newLiquidationEnv(t)
	l.auth = nil // 无仓位时连交易配置也不应访问。
	if err := l.liquidate(context.Background(), liqTrader); err != nil {
		t.Fatal(err)
	}
	if !fb.Called("balanceOf") || fb.Called("getLiquidationCost") {
		t.Fatalf("want only balance lookup, got %v", fb.Calls)
	}
}

func TestLiquidateReadErrorsHaveStageAndCause(t *testing.T) {
	for _, tc := range []struct{ method, stage string }{{"balanceOf", "balance"}, {"getLiquidationCost", "cost"}} {
		t.Run(tc.method, func(t *testing.T) {
			l, fb, _ := newLiquidationEnv(t)
			fb.Accounts[liqTrader] = chaintest.Rich()
			fb.Accounts[liqTrader].Paper = big.NewInt(5e17)
			fb.FailOnly[tc.method] = true
			err := l.liquidate(context.Background(), liqTrader)
			if !errors.Is(err, chaintest.ErrNodeDown) || !strings.Contains(strings.ToLower(err.Error()), tc.stage) {
				t.Errorf("want %s stage wrapping node error, got %v", tc.stage, err)
			}
		})
	}
}

func TestLiquidateQuotePairAndContexts(t *testing.T) {
	for _, tc := range []struct {
		name                        string
		paper, quotePaper, quoteUSD int64
	}{
		{"long pays", 5e17, 5e17, -11880},
		{"short receives", -5e17, -5e17, 12120},
		{"position shrank before quote", 5e17, 25e16, -5940},
	} {
		t.Run(tc.name, func(t *testing.T) {
			l, fb, b := newLiquidationEnv(t)
			fb.Accounts[liqTrader] = chaintest.Rich()
			fb.Accounts[liqTrader].Paper = big.NewInt(tc.paper)
			fb.LiqCost = func(perp, trader common.Address, paper *big.Int) (*big.Int, *big.Int, error) {
				if perp != l.perp || trader != liqTrader || paper.Cmp(big.NewInt(tc.paper)) != 0 {
					t.Errorf("wrong quote request: %s %s %s", perp, trader, paper)
				}
				return big.NewInt(tc.quotePaper), usd(tc.quoteUSD), nil
			}
			ctx, cancel := context.WithCancel(context.WithValue(context.Background(), liquidationTraceKey{}, t.Name()))
			defer cancel()
			originalCtx := l.auth.Context
			var estimated bool
			b.estimate = func(gotCtx context.Context, call ethereum.CallMsg) (uint64, error) {
				estimated = true
				assertLiquidationChildContext(t, ctx, gotCtx, 10*time.Second)
				if call.To == nil || *call.To != l.perp || call.From != liqOwner {
					t.Errorf("wrong send target/sender: %+v", call)
				}
				a, err := perpetual.PerpetualMetaData.GetAbi()
				if err != nil {
					t.Fatal(err)
				}
				m, err := a.MethodById(call.Data[:4])
				if err != nil || m.Name != "liquidate" {
					t.Fatalf("wrong method: %v %v", m, err)
				}
				args, err := m.Inputs.Unpack(call.Data[4:])
				if err != nil {
					t.Fatal(err)
				}
				if args[0].(common.Address) != liqOwner || args[1].(common.Address) != liqTrader ||
					args[2].(*big.Int).Cmp(big.NewInt(tc.quotePaper)) != 0 || args[3].(*big.Int).Cmp(usd(tc.quoteUSD)) != 0 {
					t.Errorf("wrong liquidation args: %v", args)
				}
				return 0, errLiqGas
			}
			err := l.liquidate(ctx, liqTrader)
			if !estimated || !errors.Is(err, errLiqGas) {
				t.Fatalf("want estimation rejection, estimated=%v err=%v", estimated, err)
			}
			if !strings.Contains(strings.ToLower(err.Error()), "liquidat") {
				t.Errorf("send error missing liquidation stage: %v", err)
			}
			if l.auth.Context != originalCtx {
				t.Error("modified shared auth context")
			}
			for _, readCtx := range b.readContexts {
				assertLiquidationChildContext(t, ctx, readCtx, 10*time.Second)
				if !errors.Is(readCtx.Err(), context.Canceled) {
					t.Errorf("completed read/send context not released: %v", readCtx.Err())
				}
			}
			if ctx.Err() != nil {
				t.Error("releasing child canceled its parent")
			}
		})
	}
}
