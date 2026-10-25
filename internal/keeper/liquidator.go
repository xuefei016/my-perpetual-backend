package keeper

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
)

// TraderSource 提供要盯的交易者名单。engine.Engine 实现它（下过单的人）。
type TraderSource interface {
	Traders() []common.Address
}

// Liquidator 清算机器人：轮询名单，发现 Dealer.IsSafe = false 就用自己的账户接走仓位。
type Liquidator struct {
	c        *chain.Chain
	auth     *bind.TransactOpts // 清算人账户（自己要有足够保证金，清算后必须 IM safe）
	perp     common.Address
	traders  TraderSource
	interval time.Duration
}

func NewLiquidator(c *chain.Chain, auth *bind.TransactOpts, perp common.Address, traders TraderSource, interval time.Duration) *Liquidator {
	return &Liquidator{c: c, auth: auth, perp: perp, traders: traders, interval: interval}
}

// Run 每 interval 调一次 checkOnce，直到 ctx 取消（返回 nil）。
func (l *Liquidator) Run(ctx context.Context) error {
	panic("TODO(D9): Liquidator.Run")
}

// checkOnce 扫一遍名单，对不安全的人调 liquidate；单个人失败不影响其他人。
func (l *Liquidator) checkOnce(ctx context.Context) error {
	panic("TODO(D9): Liquidator.checkOnce")
}

// liquidate 先读 trader 在 perp 的 paper，再用 Dealer.GetLiquidationCost 估出 expectCredit（滑点保护），
// 最后调 Perpetual.Liquidate(清算人, trader, requestPaper, expectCredit) 并等回执。
func (l *Liquidator) liquidate(ctx context.Context, trader common.Address) error {
	panic("TODO(D9): Liquidator.liquidate")
}
