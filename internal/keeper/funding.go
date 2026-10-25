package keeper

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
)

// FundingKeeper 定时更新各市场的资金费率：Dealer.UpdateFundingRate(perpList, rateList)，
// 只有 fundingRateKeeper（anvil #3）能调。
type FundingKeeper struct {
	c        *chain.Chain
	auth     *bind.TransactOpts // keeper 账户
	perps    []common.Address
	interval time.Duration
}

func NewFundingKeeper(c *chain.Chain, auth *bind.TransactOpts, perps []common.Address, interval time.Duration) *FundingKeeper {
	return &FundingKeeper{c: c, auth: auth, perps: perps, interval: interval}
}

// Run 每 interval 调一次 tick，直到 ctx 取消（返回 nil）。单次 tick 失败只记日志、不退出。
func (k *FundingKeeper) Run(ctx context.Context) error {
	panic("TODO(D9): FundingKeeper.Run")
}

// tick 算出新费率 → 发 UpdateFundingRate → WaitSuccess。
func (k *FundingKeeper) tick(ctx context.Context) error {
	panic("TODO(D9): FundingKeeper.tick")
}

// nextRates 返回每个市场新的累计资金费率（与 perps 一一对应，单位同 Perpetual.fundingRate）。
// 计算公式 D9 再定（例如按标记价与盘口中间价的溢价）。
func (k *FundingKeeper) nextRates(ctx context.Context) ([]*big.Int, error) {
	panic("TODO(D9): FundingKeeper.nextRates")
}
