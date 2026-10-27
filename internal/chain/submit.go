package chain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/xuefei016/my-perpetual-backend/internal/chain/bindings/perpetual"
)

// TradeSender 用 orderSender 账户把撮合结果提交上链，实现 engine.TradeSubmitter 接口。
type TradeSender struct {
	c    *Chain
	auth *bind.TransactOpts // orderSender（anvil #1）的签名器
}

func NewTradeSender(c *Chain, auth *bind.TransactOpts) *TradeSender {
	return &TradeSender{c: c, auth: auth}
}

// SubmitTrade 编码 tradeData → 调 perp 市场的 Perpetual.Trade → 等回执成功，返回交易哈希。
// 参数含义同 EncodeTradeData。链上 revert 时返回错误，引擎据此不改订单簿。
func (s *TradeSender) SubmitTrade(ctx context.Context, perp common.Address, orders []Order, sigs [][]byte, amounts []*big.Int) (common.Hash, error) {
	tradeData, err := EncodeTradeData(orders, sigs, amounts)
	if err != nil {
		return common.Hash{}, fmt.Errorf("encode trade data: %w", err)
	}

	perpCaller, err := perpetual.NewPerpetual(perp, s.c.Client)
	if err != nil {
		return common.Hash{}, fmt.Errorf("bind perp %s: %w", perp, err)
	}

	opts := *s.auth
	opts.Context = ctx
	tx, err := perpCaller.Trade(&opts, tradeData)
	if err != nil {
		return common.Hash{}, fmt.Errorf("send trade: %w", err)
	}
	_, err = WaitSuccess(ctx, s.c, tx)
	if err != nil {
		return tx.Hash(), fmt.Errorf("wait trade %s: %w", tx.Hash(), err)
	}
	return tx.Hash(), nil
}
