package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"slices"
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
	e.mu.Lock()
	defer e.mu.Unlock()

	book := e.books[o.Order.Perp]
	if book == nil {
		return nil, fmt.Errorf("unknown perp %s", o.Order.Perp)
	}
	if _, ok := book.Get(o.ID); ok {
		return nil, fmt.Errorf("order %s already in book", o.ID)
	}
	if o.Remaining == nil || o.Remaining.Sign() <= 0 {
		return nil, errors.New("invalid remaining amount")
	}
	left := new(big.Int).Set(o.Remaining)
	var fills []Fill
	for maker := range book.Iter(-o.Side) {
		if left.Sign() == 0 {
			break
		}
		if !crosses(o, maker) {
			break
		}
		if o.Order.Signer == maker.Order.Signer {
			continue
		}

		makerRemaining := new(big.Int).Set(maker.Remaining)
		if left.Cmp(makerRemaining) <= 0 {
			makerRemaining.Set(left)
		}
		fills = append(fills, Fill{Maker: maker, Amount: makerRemaining})
		left.Sub(left, makerRemaining)
	}

	m := &Match{
		Taker: o,
		Fills: fills,
	}
	if len(fills) > 0 {
		orders, sigs, amounts := buildTrade(o, fills)
		txHash, err := e.submitter.SubmitTrade(ctx, o.Order.Perp, orders, sigs, amounts)
		if err != nil {
			return nil, fmt.Errorf("submit trade: %w", err)
		}
		m.TxHash = txHash
		for _, fill := range fills {
			makerOrder := fill.Maker
			err = book.Fill(makerOrder.ID, fill.Amount)
			if err != nil {
				return nil, fmt.Errorf("trade %s on chain but local fill %s failed: %w",
					txHash, fill.Maker.ID, err)
			}
		}
		o.Remaining = left
	}
	if left.Sign() > 0 {
		err := book.Add(o)
		if err != nil {
			return nil, fmt.Errorf("add order %s: %w", o.ID, err)
		}
	}
	e.traders[o.Order.Signer] = struct{}{}
	return m, nil
}

// CancelOrder 撤掉 perp 市场里 ID 为 id 的挂单。
func (e *Engine) CancelOrder(perp common.Address, id string) (*BookOrder, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	book := e.books[perp]
	if book == nil {
		return nil, fmt.Errorf("unknown perp %s", perp)
	}
	o, ok := book.Cancel(id)
	if !ok {
		return nil, fmt.Errorf("order %s not found", id)
	}
	return o, nil
}

// Snapshot 返回 perp 市场买卖两侧前 depth 档。
func (e *Engine) Snapshot(perp common.Address, depth int) (bids, asks []Level, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	book := e.books[perp]
	if book == nil {
		return nil, nil, fmt.Errorf("unknown perp %s", perp)
	}
	bids = book.Depth(Buy, depth)
	asks = book.Depth(Sell, depth)
	return bids, asks, nil
}

// Traders 返回下过单的地址，实现 keeper.TraderSource。
func (e *Engine) Traders() []common.Address {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]common.Address, 0, len(e.traders))
	for addr := range e.traders {
		out = append(out, addr)
	}
	return out
}

func crosses(taker, maker *BookOrder) bool {
	t, m := taker.Order, maker.Order
	temp1 := new(big.Int).Mul(t.CreditAmount, m.PaperAmount)
	temp2 := new(big.Int).Mul(m.CreditAmount, t.PaperAmount)
	if t.PaperAmount.Sign() > 0 {
		return m.PaperAmount.Sign() < 0 && temp1.Cmp(temp2) >= 0
	}
	return m.PaperAmount.Sign() > 0 && temp1.Cmp(temp2) <= 0
}

func buildTrade(taker *BookOrder, fills []Fill) (orders []chain.Order, sigs [][]byte, amounts []*big.Int) {
	// 这里是只对副本排序，不影响fill之前的List
	sorted := slices.Clone(fills)
	slices.SortStableFunc(sorted, func(a, b Fill) int {
		return bytes.Compare(a.Maker.Order.Signer[:], b.Maker.Order.Signer[:])
	})

	orders = make([]chain.Order, 0, len(sorted)+1)
	sigs = make([][]byte, 0, len(sorted)+1)
	amounts = make([]*big.Int, 0, len(sorted)+1)

	total := big.NewInt(0)
	orders = append(orders, taker.Order)
	sigs = append(sigs, taker.Signature)
	amounts = append(amounts, total)
	for _, fill := range sorted {
		orders = append(orders, fill.Maker.Order)
		sigs = append(sigs, fill.Maker.Signature)
		amounts = append(amounts, fill.Amount)
		total.Add(total, fill.Amount)
	}
	return orders, sigs, amounts
}
