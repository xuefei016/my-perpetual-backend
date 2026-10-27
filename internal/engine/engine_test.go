package engine

import (
	"context"
	"errors"
	"math/big"
	"slices"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
)

// TestCrosses 期望值用"精确价格比较"独立推出（买：taker 价 >= maker 价；卖：taker 价 <= maker 价），
// 数量故意取得各不相同，确保 crosses 的交叉相乘和合约 _priceMatchCheck 结论一致、不受单子大小影响。
func TestCrosses(t *testing.T) {
	cases := []struct {
		name  string
		taker *BookOrder
		maker *BookOrder
		want  bool
	}{
		{"买 26000 吃 卖 25000", mustOrder(t, "t", e17(1), usdc(-2600)), mustOrder(t, "m", e17(-10), usdc(25000)), true},
		{"买 26000 吃 卖 26000（相等可成交）", mustOrder(t, "t", e17(1), usdc(-2600)), mustOrder(t, "m", e17(-3), usdc(7800)), true},
		{"买 26000 碰 卖 26100", mustOrder(t, "t", e17(5), usdc(-13000)), mustOrder(t, "m", e17(-1), usdc(2610)), false},
		{"卖 25000 吃 买 26000", mustOrder(t, "t", e17(-2), usdc(5000)), mustOrder(t, "m", e17(7), usdc(-18200)), true},
		{"卖 26000 吃 买 26000（相等可成交）", mustOrder(t, "t", e17(-10), usdc(26000)), mustOrder(t, "m", e17(1), usdc(-2600)), true},
		{"卖 26100 碰 买 26000", mustOrder(t, "t", e17(-1), usdc(2610)), mustOrder(t, "m", e17(10), usdc(-26000)), false},
		{"买 碰 买（同侧，单看交叉相乘会误判成交）", mustOrder(t, "t", e17(1), usdc(-2000)), mustOrder(t, "m", e17(1), usdc(-3000)), false},
		{"卖 碰 卖（同侧，单看交叉相乘会误判成交）", mustOrder(t, "t", e17(-1), usdc(3000)), mustOrder(t, "m", e17(-1), usdc(2000)), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := crosses(tc.taker, tc.maker); got != tc.want {
				t.Fatalf("crosses = %v, want %v (taker 价 %s, maker 价 %s)", got, tc.want, tc.taker.Price, tc.maker.Price)
			}
		})
	}
}

// 真实 anvil 地址，按数值 alice(0x15d3) < carol(0x976E) < bob(0x9965) < dave(0xa0Ee)
var (
	alice = common.HexToAddress("0x15d34AAf54267DB7D7c367839AAf71A00a2C6A65")
	carol = common.HexToAddress("0x976EA74026E726554dB657fA54763abd0C3a0aa9")
	bob   = common.HexToAddress("0x9965507D1a55bcC2695C58ba16FB37d819B0A4dc")
	dave  = common.HexToAddress("0xa0Ee7A142d267C1f36714E4a8F75612F20a79720")
)

// fakeSubmitter 记录每次 SubmitTrade 收到的参数；err 非 nil 时模拟链上 revert。
type fakeSubmitter struct {
	err   error
	calls []submitCall
}

type submitCall struct {
	perp    common.Address
	orders  []chain.Order
	sigs    [][]byte
	amounts []*big.Int
}

var fakeHash = common.HexToHash("0xfeed")

func (f *fakeSubmitter) SubmitTrade(_ context.Context, perp common.Address, orders []chain.Order, sigs [][]byte, amounts []*big.Int) (common.Hash, error) {
	f.calls = append(f.calls, submitCall{perp, orders, sigs, amounts})
	if f.err != nil {
		return common.Hash{}, f.err
	}
	return fakeHash, nil
}

// signedOrder 带 signer 的订单；签名用 id 的字节，方便检查 sigs 和 orders 是否对得上
func signedOrder(t *testing.T, id string, signer common.Address, paper, credit *big.Int) *BookOrder {
	t.Helper()
	o, err := NewBookOrder(id, chain.Order{Perp: testPerp, Signer: signer, PaperAmount: paper, CreditAmount: credit}, []byte(id))
	if err != nil {
		t.Fatalf("NewBookOrder(%s): %v", id, err)
	}
	return o
}

func mustPlace(t *testing.T, e *Engine, o *BookOrder) *Match {
	t.Helper()
	m, err := e.PlaceOrder(context.Background(), o)
	if err != nil {
		t.Fatalf("PlaceOrder(%s): %v", o.ID, err)
	}
	return m
}

// walkthroughEngine 是第 4 步布置里的贯穿例子：
//
//	asks: [25900: carol1 卖 0.3] [26000: bob 卖 0.2 → aliceS 卖 0.5 → carol2 卖 0.4] [26100: dave 卖 1]
func walkthroughEngine(t *testing.T) (*Engine, *fakeSubmitter) {
	t.Helper()
	f := &fakeSubmitter{}
	e := New(f, testPerp)
	for _, o := range []*BookOrder{
		signedOrder(t, "carol1", carol, e17(-3), usdc(7770)),
		signedOrder(t, "bob", bob, e17(-2), usdc(5200)),
		signedOrder(t, "aliceS", alice, e17(-5), usdc(13000)),
		signedOrder(t, "carol2", carol, e17(-4), usdc(10400)),
		signedOrder(t, "dave", dave, e17(-10), usdc(26100)),
	} {
		if m := mustPlace(t, e, o); len(m.Fills) != 0 {
			t.Fatalf("挂 %s 时不该成交", o.ID)
		}
	}
	if len(f.calls) != 0 {
		t.Fatalf("只挂单不该上链，SubmitTrade 被调了 %d 次", len(f.calls))
	}
	return e, f
}

func fillSummary(fills []Fill) []string {
	var out []string
	for _, f := range fills {
		out = append(out, f.Maker.ID+":"+new(big.Int).Div(f.Amount, big.NewInt(1e17)).String())
	}
	return out
}

func amountSummary(amounts []*big.Int) []int64 {
	var out []int64
	for _, a := range amounts {
		out = append(out, new(big.Int).Div(a, big.NewInt(1e17)).Int64())
	}
	return out
}

func TestPlaceOrderRestsWhenNoCross(t *testing.T) {
	e, f := walkthroughEngine(t)
	// 买 25800 低于最优卖价 25900：不成交，挂进 bids
	o := signedOrder(t, "aliceB", alice, e17(2), usdc(-5160))
	m := mustPlace(t, e, o)
	if len(m.Fills) != 0 || m.TxHash != (common.Hash{}) || len(f.calls) != 0 {
		t.Fatalf("不交叉却成交了：fills=%v tx=%s calls=%d", fillSummary(m.Fills), m.TxHash, len(f.calls))
	}
	book := e.books[testPerp]
	if got, want := levelIDs(book.bids), [][]string{{"25800", "aliceB"}}; !equalRows(got, want) {
		t.Fatalf("bids = %v, want %v", got, want)
	}
	if o.Remaining.Cmp(e17(2)) != 0 {
		t.Fatalf("Remaining = %s, want 2e17", o.Remaining)
	}
}

func TestPlaceOrderWalkthrough(t *testing.T) {
	e, f := walkthroughEngine(t)
	taker := signedOrder(t, "aliceB", alice, e17(8), usdc(-20800)) // 买 0.8 @26000
	m := mustPlace(t, e, taker)

	// Match.Fills 保持价格-时间顺序，跳过 alice 自己的 aliceS；Amount 是独立对象，Fill 后不会变 0
	if got, want := fillSummary(m.Fills), []string{"carol1:3", "bob:2", "carol2:3"}; !slices.Equal(got, want) {
		t.Fatalf("Fills = %v, want %v", got, want)
	}
	if m.Taker != taker || m.TxHash != fakeHash {
		t.Fatalf("Match.Taker/TxHash 不对：%v %s", m.Taker.ID, m.TxHash)
	}

	// 上链参数：taker 在前，maker 按 signer 升序（carol 0x97 < bob 0x99），同一个人保持原顺序
	if len(f.calls) != 1 {
		t.Fatalf("SubmitTrade 调了 %d 次，want 1", len(f.calls))
	}
	c := f.calls[0]
	if c.perp != testPerp {
		t.Errorf("perp = %s", c.perp.Hex())
	}
	var ids []string
	for i, sig := range c.sigs {
		ids = append(ids, string(sig))
		if c.orders[i].Signer != map[string]common.Address{"aliceB": alice, "carol1": carol, "carol2": carol, "bob": bob}[string(sig)] {
			t.Errorf("orders[%d] 和 sigs[%d] 对不上", i, i)
		}
	}
	if want := []string{"aliceB", "carol1", "carol2", "bob"}; !slices.Equal(ids, want) {
		t.Fatalf("提交顺序 = %v, want %v", ids, want)
	}
	if got, want := amountSummary(c.amounts), []int64{8, 3, 3, 2}; !slices.Equal(got, want) {
		t.Fatalf("amounts(0.1 张) = %v, want %v", got, want)
	}

	// 上链成功后才改簿：carol1、bob 吃完移出；carol2 剩 0.1；aliceS 不动；taker 吃满不挂
	book := e.books[testPerp]
	if got, want := levelIDs(book.asks), [][]string{{"26000", "aliceS", "carol2"}, {"26100", "dave"}}; !equalRows(got, want) {
		t.Fatalf("asks = %v, want %v", got, want)
	}
	if c2, _ := book.Get("carol2"); c2.Remaining.Cmp(e17(1)) != 0 {
		t.Fatalf("carol2 Remaining = %s, want 1e17", c2.Remaining)
	}
	if as, _ := book.Get("aliceS"); as.Remaining.Cmp(e17(5)) != 0 {
		t.Fatalf("aliceS Remaining = %s, 自己的单不该被动", as.Remaining)
	}
	if _, ok := book.Get("aliceB"); ok || len(book.bids) != 0 {
		t.Fatal("taker 已经吃满，不该挂进簿")
	}
	if taker.Remaining.Sign() != 0 {
		t.Fatalf("taker Remaining = %s, want 0", taker.Remaining)
	}
}

func TestPlaceOrderPartialThenRest(t *testing.T) {
	cases := []struct {
		name      string
		taker     func(t *testing.T) *BookOrder
		wantFills []string
		wantAmts  []int64
		wantBids  [][]string
		wantAsks  [][]string
		wantLeft  int64
	}{
		{
			// 买 1.0 @26000：吃 carol1 0.3、bob 0.2、carol2 0.4，剩 0.1 挂在 bids 26000
			name:      "买单吃穿一档半后挂剩余",
			taker:     func(t *testing.T) *BookOrder { return signedOrder(t, "aliceB", alice, e17(10), usdc(-26000)) },
			wantFills: []string{"carol1:3", "bob:2", "carol2:4"},
			wantAmts:  []int64{9, 3, 4, 2},
			wantBids:  [][]string{{"26000", "aliceB"}},
			wantAsks:  [][]string{{"26000", "aliceS"}, {"26100", "dave"}},
			wantLeft:  1,
		},
		{
			// dave 买 0.25 @25900 不是 0.1 的整数倍也能算；只吃得到 25900 那一档
			name:      "限价只够第一档",
			taker:     func(t *testing.T) *BookOrder { return signedOrder(t, "daveB", dave, e17(5), usdc(-12950)) },
			wantFills: []string{"carol1:3"},
			wantAmts:  []int64{3, 3},
			wantBids:  [][]string{{"25900", "daveB"}},
			wantAsks:  [][]string{{"26000", "bob", "aliceS", "carol2"}, {"26100", "dave"}},
			wantLeft:  2,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, f := walkthroughEngine(t)
			o := tc.taker(t)
			m := mustPlace(t, e, o)
			if got := fillSummary(m.Fills); !slices.Equal(got, tc.wantFills) {
				t.Fatalf("Fills = %v, want %v", got, tc.wantFills)
			}
			if len(f.calls) != 1 {
				t.Fatalf("SubmitTrade 调了 %d 次", len(f.calls))
			}
			if got := amountSummary(f.calls[0].amounts); !slices.Equal(got, tc.wantAmts) {
				t.Fatalf("amounts = %v, want %v", got, tc.wantAmts)
			}
			book := e.books[testPerp]
			if got := levelIDs(book.bids); !equalRows(got, tc.wantBids) {
				t.Fatalf("bids = %v, want %v", got, tc.wantBids)
			}
			if got := levelIDs(book.asks); !equalRows(got, tc.wantAsks) {
				t.Fatalf("asks = %v, want %v", got, tc.wantAsks)
			}
			if o.Remaining.Cmp(e17(tc.wantLeft)) != 0 {
				t.Fatalf("taker Remaining = %s, want %de17", o.Remaining, tc.wantLeft)
			}
			// 挂上去的剩余部分之后能被别人当 maker 吃掉
			if got, _ := book.Get(o.ID); got != o {
				t.Fatal("剩余部分没有挂进簿")
			}
		})
	}
}

func TestPlaceOrderSellTaker(t *testing.T) {
	f := &fakeSubmitter{}
	e := New(f, testPerp)
	mustPlace(t, e, signedOrder(t, "bobB", bob, e17(2), usdc(-5220)))     // 买 0.2 @26100
	mustPlace(t, e, signedOrder(t, "carolB", carol, e17(3), usdc(-7800))) // 买 0.3 @26000
	mustPlace(t, e, signedOrder(t, "daveB", dave, e17(1), usdc(-2590)))   // 买 0.1 @25900
	// alice 卖 0.4 @26000：先吃价格高的 bobB 0.2，再吃 carolB 0.2；daveB 价格不够
	m := mustPlace(t, e, signedOrder(t, "aliceS", alice, e17(-4), usdc(10400)))
	if got, want := fillSummary(m.Fills), []string{"bobB:2", "carolB:2"}; !slices.Equal(got, want) {
		t.Fatalf("Fills = %v, want %v", got, want)
	}
	var ids []string
	for _, s := range f.calls[0].sigs {
		ids = append(ids, string(s))
	}
	if want := []string{"aliceS", "carolB", "bobB"}; !slices.Equal(ids, want) {
		t.Fatalf("提交顺序 = %v, want %v（carol 地址小于 bob）", ids, want)
	}
	if got, want := levelIDs(e.books[testPerp].bids), [][]string{{"26000", "carolB"}, {"25900", "daveB"}}; !equalRows(got, want) {
		t.Fatalf("bids = %v, want %v", got, want)
	}
}

func TestPlaceOrderSubmitFailLeavesBookUntouched(t *testing.T) {
	e, f := walkthroughEngine(t)
	before := levelIDs(e.books[testPerp].asks)
	f.err = errors.New("execution reverted: ORDER_EXPIRED")
	taker := signedOrder(t, "aliceB", alice, e17(10), usdc(-26000))

	m, err := e.PlaceOrder(context.Background(), taker)
	if !errors.Is(err, f.err) {
		t.Fatalf("err = %v, 要用 %%w 把链上错误带出来", err)
	}
	if m != nil {
		t.Fatal("失败时 Match 应为 nil")
	}
	book := e.books[testPerp]
	if got := levelIDs(book.asks); !equalRows(got, before) {
		t.Fatalf("上链失败却改了 asks：%v, want %v", got, before)
	}
	for _, id := range []string{"carol1", "bob", "carol2"} {
		o, _ := book.Get(id)
		if o.Remaining.Cmp(new(big.Int).Abs(o.Order.PaperAmount)) != 0 {
			t.Fatalf("上链失败却改了 %s 的 Remaining：%s", id, o.Remaining)
		}
	}
	if _, ok := book.Get("aliceB"); ok || len(book.bids) != 0 {
		t.Fatal("上链失败不该把 taker 挂进簿")
	}
	if taker.Remaining.Cmp(e17(10)) != 0 {
		t.Fatalf("上链失败却改了 taker Remaining：%s", taker.Remaining)
	}
}

func TestPlaceOrderRejects(t *testing.T) {
	other := common.HexToAddress("0x1234")
	cases := []struct {
		name  string
		order func(t *testing.T) *BookOrder
	}{
		{"未知市场", func(t *testing.T) *BookOrder {
			o, err := NewBookOrder("x", chain.Order{Perp: other, Signer: alice, PaperAmount: e17(1), CreditAmount: usdc(-26000)}, nil)
			if err != nil {
				t.Fatal(err)
			}
			return o
		}},
		{"ID 重复（同一张单下两次）", func(t *testing.T) *BookOrder { return signedOrder(t, "bob", bob, e17(-2), usdc(5200)) }},
		{"ID 重复且会成交", func(t *testing.T) *BookOrder { return signedOrder(t, "dave", alice, e17(5), usdc(-13000)) }},
		{"Remaining 为 nil", func(t *testing.T) *BookOrder {
			o := signedOrder(t, "y", alice, e17(1), usdc(-26000))
			o.Remaining = nil
			return o
		}},
		{"Remaining 为 0", func(t *testing.T) *BookOrder {
			o := signedOrder(t, "z", alice, e17(1), usdc(-26000))
			o.Remaining = big.NewInt(0)
			return o
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, f := walkthroughEngine(t)
			before := levelIDs(e.books[testPerp].asks)
			if _, err := e.PlaceOrder(context.Background(), tc.order(t)); err == nil {
				t.Fatal("want error, got nil")
			}
			if len(f.calls) != 0 {
				t.Fatal("被拒的单不该上链")
			}
			if got := levelIDs(e.books[testPerp].asks); !equalRows(got, before) || len(e.books[testPerp].bids) != 0 {
				t.Fatalf("被拒的单改了簿：%v", got)
			}
		})
	}
}

func TestPlaceOrderRecordsTraders(t *testing.T) {
	e, _ := walkthroughEngine(t)
	mustPlace(t, e, signedOrder(t, "aliceB", alice, e17(1), usdc(-2590)))
	got := map[common.Address]bool{}
	for _, a := range e.Traders() {
		got[a] = true
	}
	for _, a := range []common.Address{alice, bob, carol, dave} {
		if !got[a] {
			t.Errorf("Traders 里缺 %s", a.Hex())
		}
	}
	if len(got) != 4 {
		t.Errorf("Traders 有 %d 个，want 4（不能重复）", len(e.Traders()))
	}
}

func TestCancelOrderAndSnapshot(t *testing.T) {
	e, _ := walkthroughEngine(t)
	mustPlace(t, e, signedOrder(t, "aliceB", alice, e17(8), usdc(-20800))) // 贯穿例子成交后

	bids, asks, err := e.Snapshot(testPerp, 1)
	if err != nil {
		t.Fatal(err)
	}
	levelsEqual(t, "bids", bids, nil)
	levelsEqual(t, "asks 前 1 档", asks, [][3]int64{{26000, 6, 2}})
	if _, asks, _ = e.Snapshot(testPerp, 0); len(asks) != 2 {
		t.Fatalf("depth 0 应返回全部 2 档，got %d", len(asks))
	}

	o, err := e.CancelOrder(testPerp, "carol2")
	if err != nil || o.ID != "carol2" || o.Remaining.Cmp(e17(1)) != 0 {
		t.Fatalf("CancelOrder = %v, %v", o, err)
	}
	if _, err := e.CancelOrder(testPerp, "carol2"); err == nil {
		t.Fatal("重复撤单应报错")
	}
	_, asks, _ = e.Snapshot(testPerp, 0)
	levelsEqual(t, "撤单后 asks", asks, [][3]int64{{26000, 5, 1}, {26100, 10, 1}})

	// 撤单不改 traders：carol 可能已经有仓位，清算机器人还要盯
	found := false
	for _, a := range e.Traders() {
		found = found || a == carol
	}
	if !found {
		t.Fatal("撤单后 carol 不应从 Traders 里消失")
	}

	other := common.HexToAddress("0x1234")
	if _, err := e.CancelOrder(other, "x"); err == nil {
		t.Fatal("未知市场撤单应报错")
	}
	if _, _, err := e.Snapshot(other, 1); err == nil {
		t.Fatal("未知市场 Snapshot 应报错")
	}
}
