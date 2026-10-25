package engine

import (
	"errors"
	"math/big"
	"slices"
	"sort"

	"github.com/ethereum/go-ethereum/common"
)

// OrderBook 一个市场的订单簿，价格-时间优先：
//   - 买单：价格高的优先；卖单：价格低的优先
//   - 同一价格：Seq 小的（先到的）优先
//
// 不加锁：并发控制由持有它的 Engine 负责。
//
// 内部结构（D7 第 6 步讨论后可调整）：每侧一个按优先级排好序的价位切片，
// 每个价位一条 FIFO 队列，外加 ID → 订单的索引，撤单不用遍历。
type OrderBook struct {
	perp    common.Address
	bids    []*priceLevel // 价格从高到低
	asks    []*priceLevel // 价格从低到高
	index   map[string]*BookOrder
	nextSeq uint64
}

// priceLevel 同一价格的所有挂单，按到达顺序排队。
type priceLevel struct {
	price  *big.Int
	orders []*BookOrder
}

// Level 对外展示的一档深度（GET /orderbook 用）。
type Level struct {
	Price  *big.Int // USDC 6 位 / 1 张
	Amount *big.Int // 该价位剩余 paper 之和（1e18 精度）
	Count  int      // 该价位挂单数
}

func NewOrderBook(perp common.Address) *OrderBook {
	return &OrderBook{perp: perp, index: make(map[string]*BookOrder)}
}

// Add 挂单：分配 Seq，放进对应一侧的对应价位队尾。ID 重复、Perp 不属于本簿、Remaining <= 0 时返回错误。
func (b *OrderBook) Add(o *BookOrder) error {
	if b.index[o.ID] != nil {
		return errors.New("order already exists")
	}
	if b.perp != o.Order.Perp {
		return errors.New("invalid perp")
	}

	if o.Remaining == nil || o.Remaining.Sign() <= 0 {
		return errors.New("invalid remaining amount")
	}

	o.Seq = b.nextSeq
	b.nextSeq++

	if o.Side == Buy {
		i := sort.Search(len(b.bids), func(i int) bool {
			return b.bids[i].price.Cmp(o.Price) <= 0
		})
		if i < len(b.bids) && b.bids[i].price.Cmp(o.Price) == 0 {
			b.bids[i].orders = append(b.bids[i].orders, o)
		} else {
			lvl := &priceLevel{price: o.Price, orders: []*BookOrder{o}}
			b.bids = slices.Insert(b.bids, i, lvl)
		}
	} else {
		i := sort.Search(len(b.asks), func(i int) bool {
			return b.asks[i].price.Cmp(o.Price) >= 0
		})
		if i < len(b.asks) && b.asks[i].price.Cmp(o.Price) == 0 {
			b.asks[i].orders = append(b.asks[i].orders, o)
		} else {
			lvl := &priceLevel{price: o.Price, orders: []*BookOrder{o}}
			b.asks = slices.Insert(b.asks, i, lvl)
		}
	}
	b.index[o.ID] = o
	return nil
}

// Cancel 撤单：从价位队列和索引里删掉，价位空了要连价位一起删。不存在时返回 (nil, false)。
func (b *OrderBook) Cancel(id string) (*BookOrder, bool) {
	bookOrder := b.index[id]
	if bookOrder == nil {
		return nil, false
	}
	b.remove(bookOrder)
	return bookOrder, true
}

// Get 按 ID 查订单。
func (b *OrderBook) Get(id string) (*BookOrder, bool) {
	o, ok := b.index[id]
	return o, ok
}

// Best 返回某一侧当前最优先的那张单（最优价位的队首），空簿返回 nil。D8 撮合时 taker 就从这里取对手单。
func (b *OrderBook) Best(side Side) *BookOrder {
	var bookOrder *BookOrder
	if side == Buy {
		if len(b.bids) > 0 {
			bookOrder = b.bids[0].orders[0]
		} else {
			bookOrder = nil
		}
	} else {
		if len(b.asks) > 0 {
			bookOrder = b.asks[0].orders[0]
		} else {
			bookOrder = nil
		}
	}

	return bookOrder
}

// Fill 记一笔成交：Remaining 减 amount（paper 绝对值，1e18），减到 0 就移出订单簿。
// 订单不存在、amount <= 0 或超过 Remaining 时返回错误（对应合约的 ORDER_FILLED_OVERFLOW）。
func (b *OrderBook) Fill(id string, amount *big.Int) error {
	if amount == nil || amount.Sign() <= 0 {
		return errors.New("invalid amount")
	}
	bookOrder := b.index[id]
	if bookOrder == nil {
		return errors.New("order not found")
	}

	if bookOrder.Remaining.Cmp(amount) < 0 {
		return errors.New("fill amount exceeds remaining")
	}
	bookOrder.Remaining.Sub(bookOrder.Remaining, amount)
	if bookOrder.Remaining.Sign() == 0 {
		b.remove(bookOrder)
	}
	return nil
}

// Depth 返回某一侧前 n 档（按优先级排），n <= 0 表示全部。
func (b *OrderBook) Depth(side Side, n int) []Level {
	levels := b.bids
	if side == Sell {
		levels = b.asks
	}
	if n <= 0 || n > len(levels) {
		n = len(levels)
	}
	cutPriceLevelSlice := levels[:n]

	result := make([]Level, len(cutPriceLevelSlice))
	for i, lvl := range cutPriceLevelSlice {
		totalRemaining := big.NewInt(0)
		for _, bookOrder := range lvl.orders {
			totalRemaining.Add(totalRemaining, bookOrder.Remaining)
		}
		result[i] = Level{
			Price:  new(big.Int).Set(lvl.price),
			Amount: totalRemaining,
			Count:  len(lvl.orders),
		}
	}
	return result
}

// Len 当前挂单总数。
func (b *OrderBook) Len() int {
	return len(b.index)
}

func (b *OrderBook) remove(o *BookOrder) {
	if o.Side == Buy {
		i := sort.Search(len(b.bids), func(i int) bool {
			return b.bids[i].price.Cmp(o.Price) <= 0
		})
		if i < len(b.bids) && b.bids[i].price.Cmp(o.Price) == 0 {
			for j, o1 := range b.bids[i].orders {
				if o1.ID == o.ID {
					b.bids[i].orders = slices.Delete(b.bids[i].orders, j, j+1)
					break
				}
			}
			if len(b.bids[i].orders) == 0 {
				b.bids = slices.Delete(b.bids, i, i+1)
			}
		}
	} else {
		i := sort.Search(len(b.asks), func(i int) bool {
			return b.asks[i].price.Cmp(o.Price) >= 0
		})
		if i < len(b.asks) && b.asks[i].price.Cmp(o.Price) == 0 {
			for j, o1 := range b.asks[i].orders {
				if o1.ID == o.ID {
					b.asks[i].orders = slices.Delete(b.asks[i].orders, j, j+1)
					break
				}
			}
			if len(b.asks[i].orders) == 0 {
				b.asks = slices.Delete(b.asks, i, i+1)
			}
		}
	}
	delete(b.index, o.ID)

}
