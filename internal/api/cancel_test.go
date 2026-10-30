package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// signCancel 模拟钱包 personal_sign："cancel <小写 orderId> <deadline>"，v = 27/28。
func signCancel(t *testing.T, by trader, id string, deadline int64) string {
	t.Helper()
	hash := accounts.TextHash([]byte(fmt.Sprintf("cancel %s %d", id, deadline)))
	sig, err := crypto.Sign(hash, by.key)
	if err != nil {
		t.Fatal(err)
	}
	sig[64] += 27
	return hexutil.Encode(sig)
}

func cancelReq(t *testing.T, h http.Handler, id string, body any) (int, map[string]any) {
	t.Helper()
	rec, out := do(t, h, context.Background(), http.MethodDelete, "/order/"+id, body)
	return rec.Code, out
}

// restCarol 挂一张 carol 卖 0.1 @26000（链上已成交 0.08 → 簿里剩 0.02），返回 orderId。
func restCarol(t *testing.T, env *testEnv, carol trader) string {
	t.Helper()
	req, id := sign(t, carol, orderSpec{signer: carol.addr, paper: amt(-100), credit: at26000(amt(-100))}, testChainID)
	env.fb.Filled[id] = amt(80)
	rec, out := post(t, env.h, req)
	wantStatus(t, rec, out, http.StatusOK)
	return id.Hex()
}

func TestCancelOrderWalkthrough(t *testing.T) {
	env := newTestEnv(t)
	carol := newTrader(t)
	id := restCarol(t, env, carol)
	dl := time.Now().Add(2 * time.Minute).Unix()
	body := CancelOrderRequest{Perp: testPerp.Hex(), Deadline: dl, Signature: signCancel(t, carol, id, dl)}

	code, out := cancelReq(t, env.h, id, body)
	if code != http.StatusOK {
		t.Fatalf("status %d %v", code, out)
	}
	if out["orderId"] != id || out["remaining"] != amt(20).String() {
		t.Errorf("resp = %v, want orderId %s remaining %s", out, id, amt(20))
	}
	if _, asks, _ := env.eng.Snapshot(testPerp, 5); len(asks) != 0 {
		t.Errorf("order still in book: %+v", asks)
	}
	if env.fb.Called("isOperatorValid") {
		t.Error("signer cancels own order: no need to query operator")
	}

	// 原样重发：单已不在簿里 → 404
	code, out = cancelReq(t, env.h, id, body)
	if code != http.StatusNotFound {
		t.Errorf("repeat cancel status %d %v, want 404", code, out)
	}
}

// 路径里传大写 ID 也要能撤（规范化成小写再查、再验签）。
func TestCancelOrderNormalizesID(t *testing.T) {
	env := newTestEnv(t)
	carol := newTrader(t)
	id := restCarol(t, env, carol)
	dl := time.Now().Add(time.Minute).Unix()
	body := CancelOrderRequest{Perp: testPerp.Hex(), Deadline: dl, Signature: signCancel(t, carol, id, dl)}
	upper := "0x" + strings.ToUpper(id[2:])
	code, out := cancelReq(t, env.h, upper, body)
	if code != http.StatusOK {
		t.Fatalf("status %d %v", code, out)
	}
	if out["orderId"] != id {
		t.Errorf("orderId = %v, want normalized %s", out["orderId"], id)
	}
}

func TestCancelOrderAuth(t *testing.T) {
	carol, bob, bot := newTrader(t), newTrader(t), newTrader(t)
	now := time.Now()
	dl := now.Add(2 * time.Minute).Unix()

	tests := []struct {
		name  string
		setup func(env *testEnv, id string) (pathID string, body any)
		code  int
	}{
		{"bob cancels carol's order", func(env *testEnv, id string) (string, any) {
			return id, CancelOrderRequest{Perp: testPerp.Hex(), Deadline: dl, Signature: signCancel(t, bob, id, dl)}
		}, http.StatusUnauthorized},
		{"authorized operator", func(env *testEnv, id string) (string, any) {
			env.fb.Operators[[2]common.Address{carol.addr, bot.addr}] = true
			return id, CancelOrderRequest{Perp: testPerp.Hex(), Deadline: dl, Signature: signCancel(t, bot, id, dl)}
		}, http.StatusOK},
		{"operator authorized the other way round", func(env *testEnv, id string) (string, any) {
			env.fb.Operators[[2]common.Address{bot.addr, carol.addr}] = true
			return id, CancelOrderRequest{Perp: testPerp.Hex(), Deadline: dl, Signature: signCancel(t, bot, id, dl)}
		}, http.StatusUnauthorized},
		{"deadline in body differs from signed", func(env *testEnv, id string) (string, any) {
			return id, CancelOrderRequest{Perp: testPerp.Hex(), Deadline: dl + 1, Signature: signCancel(t, carol, id, dl)}
		}, http.StatusUnauthorized},
		{"signature for another order", func(env *testEnv, id string) (string, any) {
			other := common.HexToHash("0x1234").Hex()
			return id, CancelOrderRequest{Perp: testPerp.Hex(), Deadline: dl, Signature: signCancel(t, carol, other, dl)}
		}, http.StatusUnauthorized},
		{"deadline too far (1h)", func(env *testEnv, id string) (string, any) {
			far := now.Add(time.Hour).Unix()
			return id, CancelOrderRequest{Perp: testPerp.Hex(), Deadline: far, Signature: signCancel(t, carol, id, far)}
		}, http.StatusBadRequest},
		{"deadline just over 5min", func(env *testEnv, id string) (string, any) {
			over := now.Add(5*time.Minute + 30*time.Second).Unix()
			return id, CancelOrderRequest{Perp: testPerp.Hex(), Deadline: over, Signature: signCancel(t, carol, id, over)}
		}, http.StatusBadRequest},
		{"replay after deadline", func(env *testEnv, id string) (string, any) {
			past := now.Add(-10 * time.Minute).Unix()
			return id, CancelOrderRequest{Perp: testPerp.Hex(), Deadline: past, Signature: signCancel(t, carol, id, past)}
		}, http.StatusBadRequest},
		// 过期请求不该浪费 RPC：节点挂了也应该先报 400
		{"expired checked before chain", func(env *testEnv, id string) (string, any) {
			env.fb.CallErr = errNodeDown
			past := now.Add(-time.Minute).Unix()
			return id, CancelOrderRequest{Perp: testPerp.Hex(), Deadline: past, Signature: signCancel(t, bob, id, past)}
		}, http.StatusBadRequest},
		{"operator check chain error", func(env *testEnv, id string) (string, any) {
			env.fb.CallErr = errNodeDown
			return id, CancelOrderRequest{Perp: testPerp.Hex(), Deadline: dl, Signature: signCancel(t, bot, id, dl)}
		}, http.StatusBadGateway},
		{"bad json", func(env *testEnv, id string) (string, any) { return id, "{" }, http.StatusBadRequest},
		{"id not hex", func(env *testEnv, id string) (string, any) {
			return "carol-order", CancelOrderRequest{Perp: testPerp.Hex(), Deadline: dl, Signature: signCancel(t, carol, id, dl)}
		}, http.StatusBadRequest},
		{"id 31 bytes", func(env *testEnv, id string) (string, any) {
			return id[:64], CancelOrderRequest{Perp: testPerp.Hex(), Deadline: dl, Signature: signCancel(t, carol, id, dl)}
		}, http.StatusBadRequest},
		{"bad perp", func(env *testEnv, id string) (string, any) {
			return id, CancelOrderRequest{Perp: "0x00" + testPerp.Hex()[2:], Deadline: dl, Signature: signCancel(t, carol, id, dl)}
		}, http.StatusBadRequest},
		{"unknown perp", func(env *testEnv, id string) (string, any) {
			return id, CancelOrderRequest{Perp: common.HexToAddress("0x99").Hex(), Deadline: dl, Signature: signCancel(t, carol, id, dl)}
		}, http.StatusNotFound},
		{"unknown order", func(env *testEnv, id string) (string, any) {
			other := common.HexToHash("0x1234").Hex()
			return other, CancelOrderRequest{Perp: testPerp.Hex(), Deadline: dl, Signature: signCancel(t, carol, other, dl)}
		}, http.StatusNotFound},
		{"signature not hex", func(env *testEnv, id string) (string, any) {
			return id, CancelOrderRequest{Perp: testPerp.Hex(), Deadline: dl, Signature: "0xzz"}
		}, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			id := restCarol(t, env, carol)
			pathID, body := tt.setup(env, id)
			code, out := cancelReq(t, env.h, pathID, body)
			if code != tt.code {
				t.Fatalf("status %d, want %d, body %v", code, tt.code, out)
			}
			_, asks, _ := env.eng.Snapshot(testPerp, 5)
			if stillThere := len(asks) == 1; stillThere == (tt.code == http.StatusOK) {
				t.Errorf("book after cancel: %+v (status %d)", asks, code)
			}
		})
	}
}
