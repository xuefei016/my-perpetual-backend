package engine

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/xuefei016/my-perpetual-backend/internal/chain"
)

// Side 买卖方向，取值和 paper 的符号一致，方便直接相乘。
type Side int8

const (
	Buy  Side = 1  // paper > 0，开多 / 平空
	Sell Side = -1 // paper < 0，开空 / 平多
)

// BookOrder 订单簿里的一张单：链上订单原文 + 签名 + 引擎自己维护的撮合状态。
type BookOrder struct {
	ID        string      // chain.Digest 的 hex：链上 orderFilledPaperAmount 也以它为 key
	Order     chain.Order // 签名原文，上链时原样带上，一个字节都不能改
	Signature []byte      // 65 字节 r ‖ s ‖ v
	Side      Side
	Price     *big.Int // 1 张值多少 USDC（6 位小数）= |credit| * 1e18 / |paper|，例 30000e6
	Remaining *big.Int // 还没成交的 paper 绝对值（1e18 精度），挂单时 = |PaperAmount|
	Seq       uint64   // 进簿顺序号，由 OrderBook.Add 分配；同价按它先来先成交
}

// NewBookOrder 校验订单（paper、credit 非零且异号，价格能整除）并算出 Side / Price / Remaining。
// id 由调用方传入（一般是 chain.Digest(...).Hex()），Seq 留给 OrderBook.Add 填。
func NewBookOrder(id string, o chain.Order, sig []byte) (*BookOrder, error) {
	if o.PaperAmount == nil || o.CreditAmount == nil {
		return nil, errors.New("paper and credit are required")
	}
	if new(big.Int).Mul(o.PaperAmount, o.CreditAmount).Sign() >= 0 {
		return nil, errors.New("invalid paper amount or credit amount")
	}

	num := new(big.Int).Mul(new(big.Int).Abs(o.CreditAmount), big.NewInt(1e18))
	paperAbs := new(big.Int).Abs(o.PaperAmount)
	price, rem := new(big.Int).QuoRem(num, paperAbs, new(big.Int))
	if rem.Sign() != 0 {
		return nil, errors.New("price is invalid")
	}

	side := Buy
	if o.PaperAmount.Sign() < 0 {
		side = Sell
	}

	return &BookOrder{
		ID:        id,
		Order:     o,
		Signature: bytes.Clone(sig),
		Side:      side,
		Price:     price,
		Remaining: paperAbs,
	}, nil
}
