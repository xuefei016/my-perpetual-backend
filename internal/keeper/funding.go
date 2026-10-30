package keeper

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
	"github.com/xuefei016/my-perpetual-backend/internal/chain/bindings/perpetual"
	"github.com/xuefei016/my-perpetual-backend/internal/engine"
)

// BookSource 提供盘口，用来算中间价。engine.Engine 实现它。
type BookSource interface {
	Snapshot(perp common.Address, depth int) (bids, asks []engine.Level, err error)
}

// fundingPeriod 资金费的结算周期：溢价按 interval / fundingPeriod 的比例折算到一次 tick。
const fundingPeriod = 8 * time.Hour

// FundingKeeper 定时更新各市场的资金费率：Dealer.UpdateFundingRate(perpList, rateList)，
// 只有 fundingRateKeeper（anvil #3）能调。
type FundingKeeper struct {
	c        *chain.Chain
	auth     *bind.TransactOpts // keeper 账户
	perps    []common.Address
	book     BookSource
	interval time.Duration
}

func NewFundingKeeper(c *chain.Chain, auth *bind.TransactOpts, perps []common.Address, book BookSource, interval time.Duration) *FundingKeeper {
	return &FundingKeeper{c: c, auth: auth, perps: perps, book: book, interval: interval}
}

// Run 每 interval 调一次 tick，直到 ctx 取消（返回 nil）。单次 tick 失败只记日志、不退出。
func (k *FundingKeeper) Run(ctx context.Context) error {
	ticker := time.NewTicker(k.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			tctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			err := k.tick(tctx)
			cancel()
			if err != nil {
				log.Println("funding ticker error:", err)
			}
		}
	}
}

// tick 调 nextRates；有要更新的市场就发 UpdateFundingRate → WaitSuccess。
func (k *FundingKeeper) tick(ctx context.Context) error {
	perps, rates, err := k.nextRates(ctx)
	if err != nil {
		return fmt.Errorf("nextRates: %w", err)
	}

	if len(perps) == 0 {
		log.Println("No perpetual rates found")
		return nil
	}

	opts := *k.auth
	opts.Context = ctx

	tx, err := k.c.Dealer.UpdateFundingRate(&opts, perps, rates)
	if err != nil {
		return fmt.Errorf("update funding rates: %w", err)
	}

	_, err = chain.WaitSuccess(ctx, k.c, tx)
	if err != nil {
		return fmt.Errorf("wait updateFundingRate %s: %w", tx.Hash().Hex(), err)
	}

	for i, p := range perps {
		log.Printf("funding updated %s → %s (tx %s)", p.Hex(), rates[i], tx.Hash().Hex())
	}
	return nil

}

// nextRates 对每个市场：盘口中间价 mid、标记价 mark、旧累计费率 old → new = old + fundingDelta(...)。
// 盘口缺一侧（没有 mid）或 delta = 0 的市场跳过，所以返回的 perps 是 k.perps 的子集，与 rates 一一对应。
func (k *FundingKeeper) nextRates(ctx context.Context) (perps []common.Address, rates []*big.Int, err error) {
	for _, perp := range k.perps {
		bids, asks, err := k.book.Snapshot(perp, 1)
		if err != nil {
			return nil, nil, fmt.Errorf("snapshot for %s: %w", perp.Hex(), err)
		}
		if len(bids) == 0 && len(asks) == 0 {
			log.Println("order book is empty", "perp", perp)
			continue
		}
		if (len(bids) != 0 && len(asks) == 0) || (len(bids) == 0 && len(asks) != 0) {
			log.Println("one of bids or ask is empty", perp)
			continue
		}

		bidPrice := bids[0].Price
		askPrice := asks[0].Price
		mid := new(big.Int).Add(bidPrice, askPrice)
		mid = mid.Quo(mid, big.NewInt(2))

		opts := &bind.CallOpts{Context: ctx}
		mp, err := k.c.MarkPrice(opts, perp)
		if err != nil {
			return nil, nil, fmt.Errorf("mark price for %s: %w", perp.Hex(), err)
		}

		delta := fundingDelta(mid, mp, k.interval)
		if delta.Sign() == 0 {
			continue
		}

		newPerp, err := perpetual.NewPerpetual(perp, k.c.Backend)
		if err != nil {
			return nil, nil, fmt.Errorf("bind funding perp %s: %w", perp.Hex(), err)
		}
		fundingRate, err := newPerp.GetFundingRate(opts)
		if err != nil {
			return nil, nil, fmt.Errorf("get funding rate for %s: %w", perp.Hex(), err)
		}
		rates = append(rates, fundingRate.Add(fundingRate, delta))
		perps = append(perps, perp)
	}
	return perps, rates, nil
}

// fundingDelta 一次 tick 的累计费率增量（USDC 6 位 / 1 张）：
// delta = −(mid − mark) × interval / fundingPeriod，再夹到 [−mark/1000, +mark/1000]。
// mid > mark（合约比指数贵）→ delta < 0 → 多头 credit 减少，多头付钱给空头。
func fundingDelta(mid, mark *big.Int, interval time.Duration) *big.Int {
	d := new(big.Int).Sub(mark, mid)      // −(mid − mark)
	d.Mul(d, big.NewInt(int64(interval))) // × 纳秒数
	d.Quo(d, big.NewInt(int64(fundingPeriod)))

	limit := new(big.Int).Quo(mark, big.NewInt(1000))
	if d.Cmp(limit) > 0 {
		return limit
	}
	if neg := new(big.Int).Neg(limit); d.Cmp(neg) < 0 {
		return neg
	}
	return d
}
