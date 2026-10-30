package api

import (
	"math/big"
	"net/http"
	"strings"
	"testing"
)

func usdc(v int64) *big.Int { return big.NewInt(v * 1e6) }

// D8 结束时 alice 的状态：多 0.54 张，净值 1482.6，IM = 0.54×26000×10% = 1404，标记价 26000。
func aliceD8() *fakeAccount {
	return &fakeAccount{
		Paper: amt(540), Credit: usdc(-14040),
		NetValue: big.NewInt(1482_600000), InitialMargin: usdc(1404), MaintenanceMargin: usdc(702),
		PrimaryCredit: big.NewInt(1482_600000), Safe: true,
	}
}

func priced(paper *big.Int, price int64) *big.Int { // credit = −paper × price（USDC 6 位）
	c := new(big.Int).Mul(paper, big.NewInt(price*1e6))
	c.Quo(c, e18)
	return c.Neg(c)
}

func TestCheckIM(t *testing.T) {
	tests := []struct {
		name   string
		acc    func() *fakeAccount
		paper  *big.Int // 带符号
		price  int64
		filled *big.Int // 链上已成交（预检要用扣过之后的 remaining）
		ok     bool
	}{
		// 布置里的 5 行贯穿例子
		{"buy 0.1 @26000: im 1664 > 1482.6", aliceD8, amt(100), 26000, nil, false},
		{"buy 0.03 @26000: im 1482 <= 1482.6", aliceD8, amt(30), 26000, nil, true},
		{"buy 0.03 @27000: net 1452.6 < 1482", aliceD8, amt(30), 27000, nil, false},
		{"sell 0.2 @25000: reduce passes", aliceD8, amt(-200), 25000, nil, true},
		{"sell 1.0 @26000: flip, im 1196", aliceD8, amt(-1000), 26000, nil, true},

		// 只按剩余量预估：buy 0.1 但链上已成交 0.07 → 只剩 0.03
		{"remaining not paper", aliceD8, amt(100), 26000, amt(70), true},
		// 空头侧同样只按剩余量：sell 0.1 已成交 0.07；若误用整张 0.1 当 fillPaper 会算出 −1820 的假浮亏
		{"remaining not paper (short)", func() *fakeAccount {
			a := aliceD8()
			a.Paper, a.Credit = amt(-540), usdc(14040)
			return a
		}, amt(-100), 26000, amt(70), true},
		// 已经低于 IM 的人反手：|new| 0.46 < |old| 0.54，但开出了新空仓，不能当减仓放行
		{"flip smaller but under IM", func() *fakeAccount {
			a := aliceD8()
			a.NetValue = usdc(100)
			return a
		}, amt(-1000), 26000, nil, false},
		// 已经低于 IM 的人减仓：必须放行，否则只能等清算
		{"reduce while under IM", func() *fakeAccount {
			a := aliceD8()
			a.NetValue = usdc(100)
			return a
		}, amt(-200), 26000, nil, true},
		// 恰好平完（newPaper = 0）算减仓，即使价格很差
		{"close exactly at bad price", aliceD8, amt(-540), 1000, nil, true},
		// 反手且新仓 IM 不够：多 0.54 卖 3 → 空 2.46，im 6396 > 1482.6
		{"flip too big", aliceD8, amt(-3000), 26000, nil, false},
		// 新开仓（oldPaper = 0）不能当成减仓放行
		{"new position, poor", func() *fakeAccount {
			return &fakeAccount{Paper: new(big.Int), Credit: new(big.Int), NetValue: usdc(100),
				InitialMargin: new(big.Int), MaintenanceMargin: new(big.Int), PrimaryCredit: usdc(100), Safe: true}
		}, amt(100), 26000, nil, false},
		// 空头加仓、卖得比标记价便宜：浮亏 −0.03×26000 + 750 = −30 → net 1452.6 < 0.57×2600 = 1482
		{"short add below mark", func() *fakeAccount {
			a := aliceD8()
			a.Paper, a.Credit = amt(-540), usdc(14040)
			return a
		}, amt(-30), 25000, nil, false},
		// 同一笔，卖在标记价：浮亏 0 → 1482.6 >= 1482
		{"short add at mark", func() *fakeAccount {
			a := aliceD8()
			a.Paper, a.Credit = amt(-540), usdc(14040)
			return a
		}, amt(-30), 26000, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			alice := newTrader(t)
			env.fb.Accounts[alice.addr] = tt.acc()
			req, id := sign(t, alice, orderSpec{signer: alice.addr, paper: tt.paper, credit: priced(tt.paper, tt.price)}, testChainID)
			if tt.filled != nil {
				env.fb.Filled[id] = tt.filled
			}
			rec, out := post(t, env.h, req)
			bids, asks, _ := env.eng.Snapshot(testPerp, 5)
			if tt.ok {
				wantStatus(t, rec, out, http.StatusOK)
				if len(bids)+len(asks) != 1 {
					t.Errorf("passed order should rest in book")
				}
				return
			}
			wantStatus(t, rec, out, http.StatusUnprocessableEntity)
			if len(bids)+len(asks) != 0 {
				t.Errorf("rejected order entered book")
			}
		})
	}
}

// 报错要带数字，方便用户知道差多少。
func TestCheckIMErrorShowsNumbers(t *testing.T) {
	env := newTestEnv(t)
	alice := newTrader(t)
	env.fb.Accounts[alice.addr] = aliceD8()
	req, _ := sign(t, alice, orderSpec{signer: alice.addr, paper: amt(100), credit: priced(amt(100), 26000)}, testChainID)
	rec, out := post(t, env.h, req)
	wantStatus(t, rec, out, http.StatusUnprocessableEntity)
	msg, _ := out["error"].(string)
	if !strings.Contains(msg, "1482600000") || !strings.Contains(msg, "1664000000") {
		t.Errorf("error should show net 1482600000 and im 1664000000, got %q", msg)
	}
}

// 合约 / 是向零截断：fillCredit = −26000000001 × 0.5 = −13000000000.5 → Quo 得 −13000000000，
// Div（欧几里得）得 −13000000001。把净值卡在边界上，差 1 就会误拒。
func TestCheckIMTruncatesTowardZero(t *testing.T) {
	env := newTestEnv(t)
	alice := newTrader(t)
	env.fb.Accounts[alice.addr] = &fakeAccount{Paper: new(big.Int), Credit: new(big.Int),
		NetValue: usdc(1300), InitialMargin: new(big.Int), MaintenanceMargin: new(big.Int),
		PrimaryCredit: usdc(1300), Safe: true}
	req, id := sign(t, alice, orderSpec{signer: alice.addr, paper: amt(1000), credit: big.NewInt(-26000_000001)}, testChainID)
	env.fb.Filled[id] = amt(500)
	rec, out := post(t, env.h, req)
	wantStatus(t, rec, out, http.StatusOK)
}

// IM 预检读链失败（任意一个调用）→ 502，不能当成"通过"放进簿。
func TestCheckIMChainErrors(t *testing.T) {
	for _, method := range []string{"getTraderRisk", "getRiskParams", "getMarkPrice", "balanceOf"} {
		t.Run(method, func(t *testing.T) {
			env := newTestEnv(t)
			alice := newTrader(t)
			env.fb.Accounts[alice.addr] = aliceD8()
			env.fb.FailOnly[method] = true
			req, _ := sign(t, alice, orderSpec{signer: alice.addr, paper: amt(30), credit: priced(amt(30), 26000)}, testChainID)
			rec, out := post(t, env.h, req)
			wantStatus(t, rec, out, http.StatusBadGateway)
			if bids, _, _ := env.eng.Snapshot(testPerp, 5); len(bids) != 0 {
				t.Errorf("order entered book although IM check failed")
			}
		})
	}
}
