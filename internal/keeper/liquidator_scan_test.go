package keeper

import (
	"context"
	"errors"
	"math/big"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/xuefei016/my-perpetual-backend/internal/chain/bindings/dealer"
	"github.com/xuefei016/my-perpetual-backend/internal/chaintest"
)

type traderList []common.Address

func (s traderList) Traders() []common.Address { return s }

// 只对 isSafe 注入按账户区分的失败，其他调用交给已有假链。
type liquidationScanBackend struct {
	bind.ContractBackend
	abi     *abi.ABI
	checked []common.Address
	fail    map[common.Address]error
	onQuery func(common.Address)
}

func (b *liquidationScanBackend) CallContract(ctx context.Context, call ethereum.CallMsg, block *big.Int) ([]byte, error) {
	if call.To != nil && *call.To == testDealer {
		m, err := b.abi.MethodById(call.Data[:4])
		if err != nil {
			return nil, err
		}
		if m.Name == "isSafe" {
			args, err := m.Inputs.Unpack(call.Data[4:])
			if err != nil {
				return nil, err
			}
			trader := args[0].(common.Address)
			b.checked = append(b.checked, trader)
			if b.onQuery != nil {
				b.onQuery(trader)
			}
			if err := b.fail[trader]; err != nil {
				return nil, err
			}
		}
	}
	return b.ContractBackend.CallContract(ctx, call, block)
}

func newLiquidationScanEnv(t *testing.T, traders ...common.Address) (*Liquidator, *chaintest.Backend, *liquidationBackend, *liquidationScanBackend) {
	t.Helper()
	l, fb, b := newLiquidationEnv(t)
	a, err := dealer.DealerMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	scan := &liquidationScanBackend{ContractBackend: b, abi: a, fail: make(map[common.Address]error)}
	l.c.Dealer, err = dealer.NewDealer(testDealer, scan)
	if err != nil {
		t.Fatal(err)
	}
	l.traders = traderList(traders)
	return l, fb, b, scan
}

func TestLiquidatorCheckOnceEmptyAndSafe(t *testing.T) {
	for _, traders := range []traderList{nil, {liqTrader}} {
		l, fb, _, scan := newLiquidationScanEnv(t, traders...)
		if err := l.checkOnce(context.Background()); err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(scan.checked, []common.Address(traders)) {
			t.Errorf("checked %v, want %v", scan.checked, traders)
		}
		if fb.Called("balanceOf") || fb.Called("getLiquidationCost") {
			t.Errorf("safe accounts must not enter liquidation: %v", fb.Calls)
		}
	}
}

func TestLiquidatorCheckOnceSkipsSelf(t *testing.T) {
	l, _, _, scan := newLiquidationScanEnv(t, liqOwner, liqTrader)
	scan.fail[liqOwner] = errors.New("self must not be queried")
	if err := l.checkOnce(context.Background()); err != nil {
		t.Errorf("self should be skipped, got %v", err)
	}
	if !slices.Equal(scan.checked, []common.Address{liqTrader}) {
		t.Errorf("checked %v, want only trader", scan.checked)
	}
}

func TestLiquidatorCheckOnceContinuesAndJoinsErrors(t *testing.T) {
	queryTrader := common.HexToAddress("0x1001")
	safeTrader := common.HexToAddress("0x1002")
	lastTrader := common.HexToAddress("0x1003")
	traders := []common.Address{queryTrader, safeTrader, liqTrader, lastTrader}
	l, fb, b, scan := newLiquidationScanEnv(t, traders...)
	queryErr, quoteErr := errors.New("query unavailable"), errors.New("quote unavailable")
	scan.fail[queryTrader] = queryErr
	acc := chaintest.Rich()
	acc.Safe, acc.Paper = false, big.NewInt(5e17)
	fb.Accounts[liqTrader] = acc
	var quoted bool
	fb.LiqCost = func(perp, trader common.Address, paper *big.Int) (*big.Int, *big.Int, error) {
		quoted = true
		if perp != l.perp || trader != liqTrader || paper.Cmp(acc.Paper) != 0 {
			t.Errorf("wrong liquidation quote request: %s %s %s", perp, trader, paper)
		}
		return nil, nil, quoteErr
	}
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), liquidationTraceKey{}, t.Name()))
	defer cancel()
	err := l.checkOnce(ctx)
	if !errors.Is(err, queryErr) || !errors.Is(err, quoteErr) {
		t.Fatalf("want both causes preserved, got %v", err)
	}
	if !quoted || !slices.Equal(scan.checked, traders) {
		t.Errorf("scan stopped early: quoted=%v checked=%v", quoted, scan.checked)
	}
	if !strings.Contains(err.Error(), queryTrader.Hex()) || !strings.Contains(strings.ToLower(err.Error()), "safe") {
		t.Errorf("isSafe error missing trader/stage: %v", err)
	}
	for _, readCtx := range b.readContexts {
		assertLiquidationChildContext(t, ctx, readCtx, 10*time.Second)
		if !errors.Is(readCtx.Err(), context.Canceled) {
			t.Errorf("completed query context not released: %v", readCtx.Err())
		}
	}
}

func TestLiquidatorCheckOnceAlreadyCanceled(t *testing.T) {
	l, _, _, scan := newLiquidationScanEnv(t, liqTrader, liqOwner)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := l.checkOnce(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("want cancellation, got %v", err)
	}
	if len(scan.checked) != 0 {
		t.Errorf("queried accounts after cancellation: %v", scan.checked)
	}
}

func TestLiquidatorCheckOnceStopsAfterCancellation(t *testing.T) {
	nextTrader := common.HexToAddress("0x2001")
	l, _, _, scan := newLiquidationScanEnv(t, liqTrader, nextTrader)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scan.onQuery = func(trader common.Address) {
		if trader == liqTrader {
			cancel()
		}
	}
	scan.fail[liqTrader] = context.Canceled
	if err := l.checkOnce(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("want cancellation cause, got %v", err)
	}
	if !slices.Equal(scan.checked, []common.Address{liqTrader}) {
		t.Errorf("continued scan after cancellation: %v", scan.checked)
	}
}
