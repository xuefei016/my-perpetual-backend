package keeper

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xuefei016/my-perpetual-backend/internal/chaintest"
	"github.com/xuefei016/my-perpetual-backend/internal/engine"
)

var (
	testDealer   = common.HexToAddress("0x5FC8d32690cc91D4c39d9d3abcBD16989F875707")
	testBTC      = common.HexToAddress("0xB7f8BC63BbcaD18155201308C8f3540b07f84F5e")
	testETH      = common.HexToAddress("0x00000000000000000000000000000000000E7400")
	testPriceSrc = common.HexToAddress("0x610178dA211FEF7D417bC0e6FeD39F05609AD788")
)

func usd(v int64) *big.Int { return big.NewInt(v * 1e6) }

// fakeBook 每个市场的最优买卖价（nil 表示那一侧空）。
type fakeBook struct {
	best map[common.Address][2]*big.Int // [bid, ask]
	err  error
}

func (b *fakeBook) Snapshot(perp common.Address, depth int) (bids, asks []engine.Level, err error) {
	if b.err != nil {
		return nil, nil, b.err
	}
	ba, ok := b.best[perp]
	if !ok {
		return nil, nil, fmt.Errorf("unknown perp %s", perp)
	}
	one := big.NewInt(1e18)
	if ba[0] != nil {
		bids = append(bids, engine.Level{Price: ba[0], Amount: one, Count: 1})
		bids = append(bids, engine.Level{Price: new(big.Int).Sub(ba[0], usd(100)), Amount: one, Count: 1})
	}
	if ba[1] != nil {
		asks = append(asks, engine.Level{Price: ba[1], Amount: one, Count: 1})
		asks = append(asks, engine.Level{Price: new(big.Int).Add(ba[1], usd(100)), Amount: one, Count: 1})
	}
	if depth < len(bids) {
		bids = bids[:depth]
	}
	if depth < len(asks) {
		asks = asks[:depth]
	}
	return bids, asks, nil
}

func TestFundingDelta(t *testing.T) {
	tests := []struct {
		name      string
		mid, mark *big.Int
		interval  time.Duration
		want      *big.Int
	}{
		{"mid > mark: longs pay", usd(26100), usd(26000), time.Hour, big.NewInt(-12_500000)},
		{"mid < mark: shorts pay", usd(25900), usd(26000), time.Hour, big.NewInt(12_500000)},
		{"clamped negative", usd(28000), usd(26000), time.Hour, big.NewInt(-26_000000)},
		{"clamped positive", usd(24000), usd(26000), time.Hour, big.NewInt(26_000000)},
		{"exactly at cap stays", usd(26208), usd(26000), time.Hour, big.NewInt(-26_000000)},
		{"zero premium", usd(26000), usd(26000), time.Hour, big.NewInt(0)},
		// −100e6 × 15s / 8h = −52083.33 → 向零截断 −52083（Div 会得 −52084）
		{"15s truncates toward zero", usd(26100), usd(26000), 15 * time.Second, big.NewInt(-52083)},
		{"15s positive", usd(25900), usd(26000), 15 * time.Second, big.NewInt(52083)},
		// 整个周期 8h：delta = −premium（不超过夹限时）
		{"full period", big.NewInt(26010_000000), usd(26000), 8 * time.Hour, big.NewInt(-10_000000)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			midCopy, markCopy := new(big.Int).Set(tt.mid), new(big.Int).Set(tt.mark)
			got := fundingDelta(tt.mid, tt.mark, tt.interval)
			if got.Cmp(tt.want) != 0 {
				t.Errorf("delta = %s, want %s", got, tt.want)
			}
			if tt.mid.Cmp(midCopy) != 0 || tt.mark.Cmp(markCopy) != 0 {
				t.Errorf("inputs modified: mid %s mark %s", tt.mid, tt.mark)
			}
			// 返回值不能和输入共用对象，否则调用方改结果会改到标记价
			if got == tt.mid || got == tt.mark {
				t.Errorf("result aliases an input")
			}
			got.Add(got, big.NewInt(1))
			if tt.mark.Cmp(markCopy) != 0 {
				t.Errorf("modifying result changed mark")
			}
		})
	}
}

type fundingEnv struct {
	fb   *chaintest.Backend
	book *fakeBook
	k    *FundingKeeper
}

func newFundingEnv(t *testing.T, interval time.Duration, perps ...common.Address) *fundingEnv {
	t.Helper()
	fb, err := chaintest.New(testDealer, testBTC, testPriceSrc)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range perps {
		fb.Perps[p] = true
	}
	c, err := fb.Chain(31337)
	if err != nil {
		t.Fatal(err)
	}
	book := &fakeBook{best: map[common.Address][2]*big.Int{}}
	return &fundingEnv{fb: fb, book: book, k: NewFundingKeeper(c, nil, perps, book, interval)}
}

func TestNextRates(t *testing.T) {
	env := newFundingEnv(t, time.Hour, testBTC, testETH)
	env.book.best[testBTC] = [2]*big.Int{usd(26000), usd(26200)} // mid 26100 → −12.5
	env.book.best[testETH] = [2]*big.Int{usd(25800), usd(26000)} // mid 25900 → +12.5（共用价格源 26000）
	env.fb.FundingRates[testBTC] = big.NewInt(-3_000000)
	env.fb.FundingRates[testETH] = big.NewInt(1_000000)

	perps, rates, err := env.k.nextRates(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(perps) != 2 || len(rates) != 2 || perps[0] != testBTC || perps[1] != testETH {
		t.Fatalf("perps = %v rates = %v", perps, rates)
	}
	if rates[0].Cmp(big.NewInt(-15_500000)) != 0 || rates[1].Cmp(big.NewInt(13_500000)) != 0 {
		t.Errorf("rates = %v, want [-15500000 13500000]", rates)
	}
}

// 盘口缺一侧、两侧都空、delta = 0 的市场跳过；返回的是子集且与 rates 对齐。
func TestNextRatesSkips(t *testing.T) {
	cases := []struct {
		name string
		btc  [2]*big.Int
	}{
		{"only asks", [2]*big.Int{nil, usd(26200)}},
		{"only bids", [2]*big.Int{usd(26000), nil}},
		{"empty book", [2]*big.Int{nil, nil}},
		{"zero premium", [2]*big.Int{usd(26000), big.NewInt(26000_000001)}}, // mid 26000.0000005 → Quo 26000
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := newFundingEnv(t, time.Hour, testBTC, testETH)
			env.book.best[testBTC] = tc.btc
			env.book.best[testETH] = [2]*big.Int{usd(25800), usd(26000)}
			env.fb.FundingRates[testETH] = big.NewInt(7)
			perps, rates, err := env.k.nextRates(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if len(perps) != 1 || len(rates) != 1 || perps[0] != testETH || rates[0].Cmp(big.NewInt(12_500007)) != 0 {
				t.Errorf("want only ETH 12500007, got perps %v rates %v", perps, rates)
			}
		})
	}
}

// mid 用的是最优一档，不是第二档。
func TestNextRatesUsesBestLevel(t *testing.T) {
	env := newFundingEnv(t, 8*time.Hour, testBTC)
	env.book.best[testBTC] = [2]*big.Int{usd(26000), big.NewInt(26020_000000)} // mid 26010 → −10
	_, rates, err := env.k.nextRates(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rates) != 1 || rates[0].Cmp(big.NewInt(-10_000000)) != 0 {
		t.Errorf("rates = %v, want [-10000000]", rates)
	}
}

// 任一读取失败：整次返回错误，不返回部分结果。
func TestNextRatesErrors(t *testing.T) {
	for _, method := range []string{"getRiskParams", "getMarkPrice", "getFundingRate"} {
		t.Run(method, func(t *testing.T) {
			env := newFundingEnv(t, time.Hour, testBTC)
			env.book.best[testBTC] = [2]*big.Int{usd(26000), usd(26200)}
			env.fb.FailOnly[method] = true
			perps, rates, err := env.k.nextRates(context.Background())
			if !errors.Is(err, chaintest.ErrNodeDown) {
				t.Errorf("err = %v, want wrapping ErrNodeDown", err)
			}
			if err != nil && !strings.Contains(err.Error(), testBTC.Hex()) {
				t.Errorf("error missing market address: %v", err)
			}
			if len(perps) != 0 || len(rates) != 0 {
				t.Errorf("partial result returned: %v %v", perps, rates)
			}
		})
	}
	t.Run("book error", func(t *testing.T) {
		env := newFundingEnv(t, time.Hour, testBTC)
		bookErr := errors.New("book down")
		env.book.err = bookErr
		if _, _, err := env.k.nextRates(context.Background()); !errors.Is(err, bookErr) || !strings.Contains(err.Error(), testBTC.Hex()) || !strings.Contains(err.Error(), "snapshot") {
			t.Errorf("want snapshot stage, market and cause, got %v", err)
		}
	})
}
