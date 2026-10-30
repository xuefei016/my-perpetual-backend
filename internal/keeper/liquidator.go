package keeper

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
	"github.com/xuefei016/my-perpetual-backend/internal/chain/bindings/perpetual"
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
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := l.checkOnce(ctx)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

// checkOnce 扫一遍名单，对不安全的人调 liquidate；单个人失败不影响其他人。
func (l *Liquidator) checkOnce(ctx context.Context) error {
	var errs []error
	for _, trader := range l.traders.Traders() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if trader == l.auth.From {
			continue
		}

		safeContext, cancel := context.WithTimeout(ctx, 10*time.Second)
		opts := &bind.CallOpts{Context: safeContext}
		isSafe, err := l.c.Dealer.IsSafe(opts, trader)
		cancel()
		if err != nil {
			errs = append(errs, fmt.Errorf("isSafe for %s: %w", trader.Hex(), err))
			continue
		}
		if !isSafe {
			err = l.liquidate(ctx, trader)
			if err != nil {
				errs = append(errs, fmt.Errorf("liquidate: %w", err))
				continue
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// liquidate 先读 trader 在 perp 的 paper，再用 Dealer.GetLiquidationCost 估出 expectCredit（滑点保护），
// 最后调 Perpetual.Liquidate(清算人, trader, requestPaper, expectCredit) 并等回执。
func (l *Liquidator) liquidate(ctx context.Context, trader common.Address) error {
	perp, err := perpetual.NewPerpetual(l.perp, l.c.Backend)
	if err != nil {
		return fmt.Errorf("bind liquidation perp %s: %w", l.perp.Hex(), err)
	}

	sendContext, cancelSend := context.WithTimeout(ctx, 10*time.Second)
	defer cancelSend()
	opts := &bind.CallOpts{Context: sendContext}

	balance, err := perp.BalanceOf(opts, trader)
	if err != nil {
		return fmt.Errorf("get balance for %s on %s: %w", trader.Hex(), l.perp.Hex(), err)
	}

	if balance.Paper.Sign() == 0 {
		return nil
	}

	cost, err := l.c.Dealer.GetLiquidationCost(opts, l.perp, trader, balance.Paper)
	if err != nil {
		return fmt.Errorf("get liquidation cost for %s on %s: %w", trader.Hex(), l.perp.Hex(), err)
	}

	lcc := cost.LiqtorCreditChange
	lpc := cost.LiqtorPaperChange

	txOpts := *l.auth
	txOpts.Context = sendContext
	tx, err := perp.Liquidate(&txOpts, l.auth.From, trader, lpc, lcc)
	cancelSend()
	if err != nil {
		return fmt.Errorf("send liquidation for %s on %s: %w", trader.Hex(), l.perp.Hex(), err)
	}
	waitContext, cancelWait := context.WithTimeout(ctx, 60*time.Second)
	txReceipt, err := chain.WaitSuccess(waitContext, l.c, tx)
	defer cancelWait()
	if err != nil {
		return fmt.Errorf("wait liquidation %s for %s: %w", tx.Hash().Hex(), trader.Hex(), err)
	}
	log.Printf("liquidated trader=%s perp=%s paper=%s credit=%s tx=%s",
		trader.Hex(), l.perp.Hex(), lpc, lcc, txReceipt.TxHash.Hex())
	return nil
}
