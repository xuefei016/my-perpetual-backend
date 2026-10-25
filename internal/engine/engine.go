package engine

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
)

// TradeSubmitter 把一次撮合结果提交上链。chain.TradeSender 实现它；单测里换成假的，不用连 anvil。
type TradeSubmitter interface {
	SubmitTrade(ctx context.Context, perp common.Address, orders []chain.Order, sigs [][]byte, amounts []*big.Int) (common.Hash, error)
}

// Fill 一次撮合里 taker 吃掉的一张 maker 单。
type Fill struct {
	Maker  *BookOrder
	Amount *big.Int // 成交 paper 绝对值（1e18）
}

// Match 一次 PlaceOrder 的结果。Fills 为空表示没吃到单、整张挂进簿里。
type Match struct {
	Taker  *BookOrder
	Fills  []Fill
	TxHash common.Hash // 上链交易哈希；没成交时为零值
}

// Engine 管理所有市场的订单簿，串行处理下单/撤单（一把锁），撮合后经 submitter 上链。
type Engine struct {
	mu        sync.Mutex
	books     map[common.Address]*OrderBook
	submitter TradeSubmitter
	traders   map[common.Address]struct{} // 下过单的人，清算机器人要盯的名单
}

func New(submitter TradeSubmitter, perps ...common.Address) *Engine {
	e := &Engine{
		books:     make(map[common.Address]*OrderBook),
		submitter: submitter,
		traders:   make(map[common.Address]struct{}),
	}
	for _, p := range perps {
		e.books[p] = NewOrderBook(p)
	}
	return e
}

// PlaceOrder 价格-时间优先撮合：taker 依次吃对手侧最优单直到价格不再交叉或吃完；
// 有成交就一次性 SubmitTrade 上链，成功后才 Fill 订单簿；剩余部分挂单。
func (e *Engine) PlaceOrder(ctx context.Context, o *BookOrder) (*Match, error) {
	panic("TODO(D8): Engine.PlaceOrder")
}

// CancelOrder 撤掉 perp 市场里 ID 为 id 的挂单。
func (e *Engine) CancelOrder(perp common.Address, id string) (*BookOrder, error) {
	panic("TODO(D8): Engine.CancelOrder")
}

// Snapshot 返回 perp 市场买卖两侧前 depth 档。
func (e *Engine) Snapshot(perp common.Address, depth int) (bids, asks []Level, err error) {
	panic("TODO(D8): Engine.Snapshot")
}

// Traders 返回下过单的地址，实现 keeper.TraderSource。
func (e *Engine) Traders() []common.Address {
	panic("TODO(D8): Engine.Traders")
}
