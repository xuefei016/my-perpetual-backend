package engine

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
)

var testPerp = common.HexToAddress("0xB7f8BC63BbcaD18155201308C8f3540b07f84F5e")

// e17 返回 n × 1e17（n 张的十分之一），让用例里的数量好读：e17(3) = 0.3 张
func e17(n int64) *big.Int { return new(big.Int).Mul(big.NewInt(n), big.NewInt(1e17)) }

// usdc 返回 n × 1e6（n USDC）
func usdc(n int64) *big.Int { return new(big.Int).Mul(big.NewInt(n), big.NewInt(1e6)) }

// mustOrder 用 paper / credit 造一张进簿前的订单，ID 直接用名字（真实场景是 digest hex）
func mustOrder(t *testing.T, id string, paper, credit *big.Int) *BookOrder {
	t.Helper()
	o, err := NewBookOrder(id, chain.Order{Perp: testPerp, PaperAmount: paper, CreditAmount: credit}, []byte{1, 2, 3})
	if err != nil {
		t.Fatalf("NewBookOrder(%s): %v", id, err)
	}
	return o
}

func TestNewBookOrder(t *testing.T) {
	sig := []byte{0xaa, 0xbb}
	paper, credit := e17(-1), usdc(2600)
	o, err := NewBookOrder("bob", chain.Order{Perp: testPerp, PaperAmount: paper, CreditAmount: credit}, sig)
	if err != nil {
		t.Fatal(err)
	}
	if o.ID != "bob" {
		t.Errorf("ID = %q, want %q", o.ID, "bob")
	}
	if o.Side != Sell {
		t.Errorf("Side = %d, want Sell", o.Side)
	}
	if o.Price.Cmp(usdc(26000)) != 0 {
		t.Errorf("Price = %s, want 26000e6", o.Price)
	}
	if o.Remaining.Cmp(e17(1)) != 0 {
		t.Errorf("Remaining = %s, want 1e17", o.Remaining)
	}
	// 签名原文不能被改：paper 仍是负数；Remaining 不能和 PaperAmount 共用一个 big.Int
	if paper.Cmp(e17(-1)) != 0 {
		t.Errorf("NewBookOrder 改了订单原文的 paper：%s", paper)
	}
	o.Remaining.Sub(o.Remaining, big.NewInt(1))
	if o.Order.PaperAmount.Cmp(e17(-1)) != 0 {
		t.Error("Remaining 和 Order.PaperAmount 是同一个 big.Int，Fill 会把签名原文改掉")
	}
	// 签名要存副本：调用方之后复用这块内存，不能影响簿里的单
	sig[0] = 0
	if !bytes.Equal(o.Signature, []byte{0xaa, 0xbb}) {
		t.Error("Signature 没有拷贝，调用方改 sig 影响了簿里的单")
	}

	buy := mustOrder(t, "alice", e17(1), usdc(-2600))
	if buy.Side != Buy || buy.Price.Cmp(usdc(26000)) != 0 {
		t.Errorf("买单 Side=%d Price=%s", buy.Side, buy.Price)
	}
}

func TestNewBookOrderRejects(t *testing.T) {
	cases := map[string]chain.Order{
		"paper nil":          {Perp: testPerp, CreditAmount: usdc(2600)},
		"credit nil":         {Perp: testPerp, PaperAmount: e17(1)},
		"paper 0":            {Perp: testPerp, PaperAmount: big.NewInt(0), CreditAmount: usdc(-2600)},
		"credit 0":           {Perp: testPerp, PaperAmount: e17(1), CreditAmount: big.NewInt(0)},
		"同号（都正）":             {Perp: testPerp, PaperAmount: e17(1), CreditAmount: usdc(2600)},
		"同号（都负）":             {Perp: testPerp, PaperAmount: e17(-1), CreditAmount: usdc(-2600)},
		"价格不能整除 0.3 张 10000": {Perp: testPerp, PaperAmount: e17(3), CreditAmount: usdc(-10000)},
	}
	for name, o := range cases {
		if _, err := NewBookOrder(name, o, nil); err == nil {
			t.Errorf("%s：应该报错", name)
		}
	}
}

// 布置里的 5 张单：A 卖 0.2@26100、B 卖 0.1@26000、C 卖 0.3@26000、D 买 0.5@25900、E 买 0.1@25950
func fiveOrderBook(t *testing.T) *OrderBook {
	t.Helper()
	b := NewOrderBook(testPerp)
	for _, o := range []*BookOrder{
		mustOrder(t, "A", e17(-2), usdc(5220)),
		mustOrder(t, "B", e17(-1), usdc(2600)),
		mustOrder(t, "C", e17(-3), usdc(7800)),
		mustOrder(t, "D", e17(5), usdc(-12950)),
		mustOrder(t, "E", e17(1), usdc(-2595)),
	} {
		if err := b.Add(o); err != nil {
			t.Fatalf("Add(%s): %v", o.ID, err)
		}
	}
	return b
}

// levelIDs 把一侧的价位结构拍平成 [[价格, id, id…]…]，方便和期望整体比较
func levelIDs(levels []*priceLevel) [][]string {
	var out [][]string
	for _, l := range levels {
		row := []string{new(big.Int).Div(l.price, big.NewInt(1e6)).String()}
		for _, o := range l.orders {
			row = append(row, o.ID)
		}
		out = append(out, row)
	}
	return out
}

func equalRows(a, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}

func TestAddPriceTimePriority(t *testing.T) {
	b := fiveOrderBook(t)

	// 卖：价低优先；B、C 同价按到达顺序。A 最早到但价差，排最后 —— 价格优先于时间
	if got, want := levelIDs(b.asks), [][]string{{"26000", "B", "C"}, {"26100", "A"}}; !equalRows(got, want) {
		t.Errorf("asks = %v, want %v", got, want)
	}
	// 买：价高优先
	if got, want := levelIDs(b.bids), [][]string{{"25950", "E"}, {"25900", "D"}}; !equalRows(got, want) {
		t.Errorf("bids = %v, want %v", got, want)
	}
	if got := b.Best(Sell); got == nil || got.ID != "B" {
		t.Errorf("Best(Sell) = %v, want B", got)
	}
	if got := b.Best(Buy); got == nil || got.ID != "E" {
		t.Errorf("Best(Buy) = %v, want E", got)
	}
	if b.Len() != 5 {
		t.Errorf("Len = %d, want 5", b.Len())
	}

	// Seq 按到达顺序严格递增
	ids := []string{"A", "B", "C", "D", "E"}
	for i := 1; i < len(ids); i++ {
		prev, _ := b.Get(ids[i-1])
		cur, _ := b.Get(ids[i])
		if prev == nil || cur == nil || cur.Seq <= prev.Seq {
			t.Fatalf("Seq 没有按到达顺序递增：%s, %s", ids[i-1], ids[i])
		}
	}

	// 中间插入一档：卖 0.1@26050 应该落在 26000 和 26100 之间；买 0.1@26000 应该成为新的最优买价
	if err := b.Add(mustOrder(t, "F", e17(-1), usdc(2605))); err != nil {
		t.Fatal(err)
	}
	if err := b.Add(mustOrder(t, "G", e17(1), usdc(-2600))); err != nil {
		t.Fatal(err)
	}
	if got, want := levelIDs(b.asks), [][]string{{"26000", "B", "C"}, {"26050", "F"}, {"26100", "A"}}; !equalRows(got, want) {
		t.Errorf("插入中间档后 asks = %v, want %v", got, want)
	}
	if got := b.Best(Buy); got == nil || got.ID != "G" {
		t.Errorf("新的最优买价应该是 G，got %v", got)
	}
	// 买侧同价也要先到先得：H 和 G 同价但晚到，排在 G 后面，最优仍是 G
	if err := b.Add(mustOrder(t, "H", e17(2), usdc(-5200))); err != nil {
		t.Fatal(err)
	}
	if got, want := levelIDs(b.bids), [][]string{{"26000", "G", "H"}, {"25950", "E"}, {"25900", "D"}}; !equalRows(got, want) {
		t.Errorf("买侧同价 bids = %v, want %v", got, want)
	}
	if got := b.Best(Buy); got == nil || got.ID != "G" {
		t.Errorf("同价晚到的 H 插队了：Best(Buy) = %v", got)
	}
}

func TestGetAndBestOnEmpty(t *testing.T) {
	b := NewOrderBook(testPerp)
	if b.Best(Buy) != nil || b.Best(Sell) != nil {
		t.Error("空簿 Best 应该返回 nil")
	}
	if _, ok := b.Get("nope"); ok {
		t.Error("空簿 Get 应该返回 false")
	}
	b2 := fiveOrderBook(t)
	if o, ok := b2.Get("C"); !ok || o.Remaining.Cmp(e17(3)) != 0 {
		t.Errorf("Get(C) = %v, %v", o, ok)
	}
}

func TestAddRejects(t *testing.T) {
	b := fiveOrderBook(t)

	dup := mustOrder(t, "B", e17(-1), usdc(2600))
	if err := b.Add(dup); err == nil {
		t.Error("重复 ID 应该报错")
	}

	other := mustOrder(t, "X", e17(-1), usdc(2600))
	other.Order.Perp = common.HexToAddress("0x5FC8d32690cc91D4c39d9d3abcBD16989F875707")
	if err := b.Add(other); err == nil {
		t.Error("别的市场的单应该报错")
	}

	zero := mustOrder(t, "Y", e17(-1), usdc(2600))
	zero.Remaining = big.NewInt(0)
	if err := b.Add(zero); err == nil {
		t.Error("Remaining = 0 应该报错")
	}

	nilRem := mustOrder(t, "Z", e17(-1), usdc(2600))
	nilRem.Remaining = nil
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Error("Remaining = nil 时 Add panic 了，应该返回错误")
			}
		}()
		if err := b.Add(nilRem); err == nil {
			t.Error("Remaining = nil 应该报错")
		}
	}()

	// 被拒绝的 Add 不能改动订单簿
	if b.Len() != 5 {
		t.Errorf("被拒绝的 Add 改动了订单簿：Len = %d", b.Len())
	}
	if got, want := levelIDs(b.asks), [][]string{{"26000", "B", "C"}, {"26100", "A"}}; !equalRows(got, want) {
		t.Errorf("被拒绝的 Add 改动了 asks：%v", got)
	}
}

// levelsEqual 比较 Depth 结果和期望：价格用 USDC 整数、数量用 0.1 张为单位
func levelsEqual(t *testing.T, name string, got []Level, want [][3]int64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s：%d 档，want %d 档（got %v）", name, len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].Price.Cmp(usdc(w[0])) != 0 || got[i].Amount.Cmp(e17(w[1])) != 0 || got[i].Count != int(w[2]) {
			t.Errorf("%s 第 %d 档 = {%s, %s, %d}, want {%de6, %de17, %d}", name, i, got[i].Price, got[i].Amount, got[i].Count, w[0], w[1], w[2])
		}
	}
}

func TestDepth(t *testing.T) {
	b := fiveOrderBook(t)
	levelsEqual(t, "Depth(Sell, 0)", b.Depth(Sell, 0), [][3]int64{{26000, 4, 2}, {26100, 2, 1}})
	levelsEqual(t, "Depth(Buy, 0)", b.Depth(Buy, 0), [][3]int64{{25950, 1, 1}, {25900, 5, 1}})
	levelsEqual(t, "Depth(Buy, 1)", b.Depth(Buy, 1), [][3]int64{{25950, 1, 1}})
	levelsEqual(t, "Depth(Sell, 99)", b.Depth(Sell, 99), [][3]int64{{26000, 4, 2}, {26100, 2, 1}})
	levelsEqual(t, "空簿 Depth", NewOrderBook(testPerp).Depth(Buy, 0), nil)

	// 返回值是副本：外部改它不能改到簿里
	d := b.Depth(Sell, 1)
	d[0].Price.SetInt64(1)
	d[0].Amount.SetInt64(1)
	levelsEqual(t, "改返回值之后", b.Depth(Sell, 1), [][3]int64{{26000, 4, 2}})
	if o, _ := b.Get("B"); o.Remaining.Cmp(e17(1)) != 0 {
		t.Errorf("改 Depth 返回值改到了订单 B：%s", o.Remaining)
	}
}

func TestCancel(t *testing.T) {
	b := fiveOrderBook(t)

	o, ok := b.Cancel("B")
	if !ok || o == nil || o.ID != "B" {
		t.Fatalf("Cancel(B) = %v, %v", o, ok)
	}
	if got, want := levelIDs(b.asks), [][]string{{"26000", "C"}, {"26100", "A"}}; !equalRows(got, want) {
		t.Errorf("撤 B 后 asks = %v, want %v", got, want)
	}
	if got := b.Best(Sell); got == nil || got.ID != "C" {
		t.Errorf("撤 B 后 Best(Sell) 应该是 C，got %v", got)
	}
	if _, ok := b.Get("B"); ok {
		t.Error("撤掉的 B 还能 Get 到")
	}

	// 撤掉一档里唯一的单 → 整档删除
	if _, ok := b.Cancel("A"); !ok {
		t.Fatal("Cancel(A) 失败")
	}
	if got, want := levelIDs(b.asks), [][]string{{"26000", "C"}}; !equalRows(got, want) {
		t.Errorf("撤 A 后 asks = %v, want %v（空档要删掉）", got, want)
	}
	if o, ok := b.Cancel("A"); ok || o != nil {
		t.Errorf("重复撤单应该返回 (nil, false)，got %v, %v", o, ok)
	}

	// 买侧：撤最优档的单，下一档顶上来
	if _, ok := b.Cancel("E"); !ok {
		t.Fatal("Cancel(E) 失败")
	}
	if got := b.Best(Buy); got == nil || got.ID != "D" {
		t.Errorf("撤 E 后 Best(Buy) 应该是 D，got %v", got)
	}
	if b.Len() != 2 {
		t.Errorf("Len = %d, want 2", b.Len())
	}

	// 撤掉之后同 ID 可以重新挂（索引确实删干净了）
	if err := b.Add(mustOrder(t, "B", e17(-1), usdc(2600))); err != nil {
		t.Errorf("撤单后重新挂 B 失败：%v", err)
	}
}

// 同一档中间的单被撤：前后顺序保持不变
func TestCancelMiddleKeepsFIFO(t *testing.T) {
	b := NewOrderBook(testPerp)
	for _, id := range []string{"P", "Q", "R"} {
		if err := b.Add(mustOrder(t, id, e17(1), usdc(-2600))); err != nil {
			t.Fatal(err)
		}
	}
	b.Cancel("Q")
	if got, want := levelIDs(b.bids), [][]string{{"26000", "P", "R"}}; !equalRows(got, want) {
		t.Errorf("撤中间的 Q 后 bids = %v, want %v", got, want)
	}
}

func TestFill(t *testing.T) {
	b := fiveOrderBook(t)
	b.Cancel("B")
	b.Cancel("A")

	// 部分成交：C 0.3 → 0.2，仍在簿上
	if err := b.Fill("C", e17(1)); err != nil {
		t.Fatal(err)
	}
	if o, ok := b.Get("C"); !ok || o.Remaining.Cmp(e17(2)) != 0 {
		t.Errorf("部分成交后 C = %v, %v", o, ok)
	}
	levelsEqual(t, "部分成交后 Depth(Sell)", b.Depth(Sell, 0), [][3]int64{{26000, 2, 1}})
	// 签名原文不能被 Fill 改到（Remaining 必须是独立的 big.Int）
	if o, _ := b.Get("C"); o.Order.PaperAmount.Cmp(e17(-3)) != 0 {
		t.Errorf("Fill 改了签名原文 PaperAmount：%s", o.Order.PaperAmount)
	}

	// 超额：报错且不改簿
	if err := b.Fill("C", e17(3)); err == nil {
		t.Error("超过剩余量应该报错")
	}
	if o, _ := b.Get("C"); o.Remaining.Cmp(e17(2)) != 0 {
		t.Errorf("超额 Fill 被拒后 Remaining 变了：%s", o.Remaining)
	}

	// 刚好成交完：移出订单簿，空档删掉
	if err := b.Fill("C", e17(2)); err != nil {
		t.Fatal(err)
	}
	if _, ok := b.Get("C"); ok {
		t.Error("成交完的 C 还在簿上")
	}
	if b.Best(Sell) != nil || len(b.asks) != 0 {
		t.Errorf("卖侧应该空了：asks = %v", levelIDs(b.asks))
	}
	if b.Len() != 2 {
		t.Errorf("Len = %d, want 2（D、E）", b.Len())
	}
}

func TestFillRejects(t *testing.T) {
	b := fiveOrderBook(t)
	cases := map[string]struct {
		id     string
		amount *big.Int
	}{
		"不存在的单":      {"nope", e17(1)},
		"amount 0":   {"B", big.NewInt(0)},
		"amount 负":   {"B", e17(-1)},
		"amount nil": {"B", nil},
	}
	for name, c := range cases {
		if err := b.Fill(c.id, c.amount); err == nil {
			t.Errorf("%s：应该报错", name)
		}
	}
	if o, _ := b.Get("B"); o.Remaining.Cmp(e17(1)) != 0 {
		t.Errorf("被拒绝的 Fill 改了 B：%s", o.Remaining)
	}
}
