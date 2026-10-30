package api

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func get(t *testing.T, h http.Handler, path string) (int, map[string]any) {
	t.Helper()
	rec, out := do(t, h, context.Background(), http.MethodGet, path, nil)
	return rec.Code, out
}

// rest 挂一张不会成交的单（价格 = price，USDC 整数）。
func rest(t *testing.T, env *testEnv, paper *big.Int, price int64) {
	t.Helper()
	tr := newTrader(t)
	req, _ := sign(t, tr, orderSpec{signer: tr.addr, paper: paper, credit: priced(paper, price)}, testChainID)
	rec, out := post(t, env.h, req)
	wantStatus(t, rec, out, http.StatusOK)
}

func levels(t *testing.T, v any) []map[string]any {
	t.Helper()
	arr, ok := v.([]any)
	if !ok {
		t.Fatalf("levels should be a JSON array, got %#v", v)
	}
	out := make([]map[string]any, len(arr))
	for i, x := range arr {
		out[i] = x.(map[string]any)
	}
	return out
}

func wantLevel(t *testing.T, got map[string]any, price int64, amount *big.Int, count int) {
	t.Helper()
	if got["price"] != big.NewInt(price*1e6).String() || got["amount"] != amount.String() || got["count"] != float64(count) {
		t.Errorf("level = %v, want price %d amount %s count %d", got, price*1e6, amount, count)
	}
}

func TestOrderBook(t *testing.T) {
	env := newTestEnv(t)
	rest(t, env, amt(-20), 26000)
	rest(t, env, amt(-10), 26000)
	rest(t, env, amt(-50), 26100)
	rest(t, env, amt(30), 25900)
	rest(t, env, amt(40), 25800)
	rest(t, env, amt(70), 25700)

	// 小写地址也能查；返回 EIP-55 规范写法
	code, out := get(t, env.h, "/orderbook?perp="+strings.ToLower(testPerp.Hex())+"&depth=2")
	if code != http.StatusOK {
		t.Fatalf("status %d %v", code, out)
	}
	if out["perp"] != testPerp.Hex() {
		t.Errorf("perp = %v, want %s", out["perp"], testPerp.Hex())
	}
	bids, asks := levels(t, out["bids"]), levels(t, out["asks"])
	if len(bids) != 2 || len(asks) != 2 {
		t.Fatalf("depth=2: got %d bids %d asks: %v", len(bids), len(asks), out)
	}
	wantLevel(t, bids[0], 25900, amt(30), 1)
	wantLevel(t, bids[1], 25800, amt(40), 1)
	wantLevel(t, asks[0], 26000, amt(30), 2)
	wantLevel(t, asks[1], 26100, amt(50), 1)

	// 不传 depth 用默认值（≥ 3 档都在）
	code, out = get(t, env.h, "/orderbook?perp="+testPerp.Hex())
	if code != http.StatusOK || len(levels(t, out["bids"])) != 3 || len(levels(t, out["asks"])) != 2 {
		t.Errorf("default depth: %d %v", code, out)
	}
}

func TestOrderBookEmptyIsArrayNotNull(t *testing.T) {
	env := newTestEnv(t)
	rec, _ := do(t, env.h, context.Background(), http.MethodGet, "/orderbook?perp="+testPerp.Hex(), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw["bids"]) != "[]" || string(raw["asks"]) != "[]" {
		t.Errorf("empty book should be [] not null: bids=%s asks=%s", raw["bids"], raw["asks"])
	}
}

func TestOrderBookRejects(t *testing.T) {
	tests := []struct {
		query string
		code  int
	}{
		{"", http.StatusBadRequest},
		{"perp=0x1234", http.StatusBadRequest},
		{"perp=" + testPerp.Hex() + "&depth=0", http.StatusBadRequest},
		{"perp=" + testPerp.Hex() + "&depth=-1", http.StatusBadRequest},
		{"perp=" + testPerp.Hex() + "&depth=abc", http.StatusBadRequest},
		{"perp=" + testPerp.Hex() + "&depth=101", http.StatusBadRequest},
		{"perp=" + testPerp.Hex() + "&depth=100000000", http.StatusBadRequest},
		{"perp=" + common.HexToAddress("0x99").Hex(), http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			env := newTestEnv(t)
			code, out := get(t, env.h, "/orderbook?"+tt.query)
			if code != tt.code {
				t.Errorf("status %d, want %d, %v", code, tt.code, out)
			}
		})
	}
	env := newTestEnv(t)
	if code, out := get(t, env.h, "/orderbook?perp="+testPerp.Hex()+"&depth=100"); code != http.StatusOK {
		t.Errorf("depth=100 should be allowed: %d %v", code, out)
	}
}

func TestPosition(t *testing.T) {
	env := newTestEnv(t)
	alice := newTrader(t)
	acc := aliceD8()
	acc.Safe = false // 用 false 区分"读到了"和"零值"
	env.fb.Accounts[alice.addr] = acc

	code, out := get(t, env.h, "/position?trader="+strings.ToLower(alice.addr.Hex())+"&perp="+strings.ToLower(testPerp.Hex()))
	if code != http.StatusOK {
		t.Fatalf("status %d %v", code, out)
	}
	want := map[string]any{
		"trader": alice.addr.Hex(), "perp": testPerp.Hex(),
		"paper": "540000000000000000", "credit": "-14040000000", "primaryCredit": "1482600000", "safe": false,
	}
	for k, v := range want {
		if out[k] != v {
			t.Errorf("%s = %v, want %v", k, out[k], v)
		}
	}
	// credit 和 primaryCredit 要来自不同的调用（防止拿错字段）
	acc.Credit, acc.PrimaryCredit, acc.Safe = usdc(-1), usdc(2), true
	_, out = get(t, env.h, "/position?trader="+alice.addr.Hex()+"&perp="+testPerp.Hex())
	if out["credit"] != "-1000000" || out["primaryCredit"] != "2000000" || out["safe"] != true {
		t.Errorf("after change: %v", out)
	}
}

func TestPositionRejects(t *testing.T) {
	alice := newTrader(t)
	tests := []struct {
		name  string
		query string
		fail  string // 让哪个链上调用失败
		code  int
	}{
		{"no trader", "perp=" + testPerp.Hex(), "", http.StatusBadRequest},
		{"no perp", "trader=" + alice.addr.Hex(), "", http.StatusBadRequest},
		{"bad trader", "trader=0x00" + alice.addr.Hex()[2:] + "&perp=" + testPerp.Hex(), "", http.StatusBadRequest},
		{"unknown perp", "trader=" + alice.addr.Hex() + "&perp=" + common.HexToAddress("0x99").Hex(), "", http.StatusNotFound},
		{"balanceOf fails", "trader=" + alice.addr.Hex() + "&perp=" + testPerp.Hex(), "balanceOf", http.StatusBadGateway},
		{"primary fails", "trader=" + alice.addr.Hex() + "&perp=" + testPerp.Hex(), "getPrimaryCredit", http.StatusBadGateway},
		{"isSafe fails", "trader=" + alice.addr.Hex() + "&perp=" + testPerp.Hex(), "isSafe", http.StatusBadGateway},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			if tt.fail != "" {
				env.fb.FailOnly[tt.fail] = true
			}
			code, out := get(t, env.h, "/position?"+tt.query)
			if code != tt.code {
				t.Errorf("status %d, want %d, %v", code, tt.code, out)
			}
			if tt.code == http.StatusNotFound && len(env.fb.Calls) != 0 {
				t.Errorf("unknown perp should be rejected before chain calls: %v", env.fb.Calls)
			}
		})
	}
}
