package api

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
	"github.com/xuefei016/my-perpetual-backend/internal/chaintest"
	"github.com/xuefei016/my-perpetual-backend/internal/engine"
)

var (
	testDealer        = common.HexToAddress("0x5FC8d32690cc91D4c39d9d3abcBD16989F875707")
	testPerp          = common.HexToAddress("0xB7f8BC63BbcaD18155201308C8f3540b07f84F5e")
	fakeTxHash        = common.HexToHash("0xabcd")
	testChainID int64 = 4242
)

// fakeSubmitter 记录提交参数；err 非 nil 时模拟上链 revert。ctxErr 记录提交时 ctx 是否已被取消。
type fakeSubmitter struct {
	mu      sync.Mutex
	err     error
	calls   int
	amounts []*big.Int
	ctxErr  error
}

func (f *fakeSubmitter) SubmitTrade(ctx context.Context, _ common.Address, _ []chain.Order, _ [][]byte, amounts []*big.Int) (common.Hash, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.amounts = amounts
	f.ctxErr = ctx.Err()
	if f.err != nil {
		return common.Hash{}, f.err
	}
	return fakeTxHash, nil
}

type testEnv struct {
	srv *Server
	h   http.Handler
	fb  *chaintest.Backend
	sub *fakeSubmitter
	eng *engine.Engine
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	fb := newFakeBackend(t)
	sub := &fakeSubmitter{}
	eng := engine.New(sub, testPerp)
	srv := New(eng, newFakeChain(t, fb), testDealer)
	return &testEnv{srv: srv, h: srv.Handler(), fb: fb, sub: sub, eng: eng}
}

type trader struct {
	key  *ecdsa.PrivateKey
	addr common.Address
}

func newTrader(t *testing.T) trader {
	t.Helper()
	k, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	return trader{key: k, addr: crypto.PubkeyToAddress(k.PublicKey)}
}

var e18 = big.NewInt(1e18)

func amt(paperMilli int64) *big.Int { // 0.001 张为单位，避免手写 1e18
	return new(big.Int).Mul(big.NewInt(paperMilli), big.NewInt(1e15))
}

type orderSpec struct {
	signer     common.Address
	paper      *big.Int
	credit     *big.Int
	expiration uint64
	nonce      uint64
}

func (o orderSpec) chainOrder() chain.Order {
	exp := o.expiration
	if exp == 0 {
		exp = uint64(time.Now().Add(time.Hour).Unix())
	}
	return chain.Order{
		Perp: testPerp, Signer: o.signer, PaperAmount: o.paper, CreditAmount: o.credit,
		Info: chain.PackInfo(chain.OrderInfo{Expiration: exp, Nonce: o.nonce}),
	}
}

// sign 返回请求体和期望的订单 ID（digest）。签名人可以和 signer 不同（operator 场景）。
func sign(t *testing.T, by trader, o orderSpec, chainID int64) (PlaceOrderRequest, common.Hash) {
	t.Helper()
	co := o.chainOrder()
	digest := chain.Digest(chain.DomainSeparator(big.NewInt(chainID), testDealer), co)
	sig, err := chain.SignOrder(by.key, digest)
	if err != nil {
		t.Fatal(err)
	}
	return PlaceOrderRequest{
		Perp:         co.Perp.Hex(),
		Signer:       co.Signer.Hex(),
		PaperAmount:  co.PaperAmount.String(),
		CreditAmount: co.CreditAmount.String(),
		Info:         hexutil.Encode(co.Info[:]),
		Signature:    hexutil.Encode(sig),
	}, digest
}

// price 26000：credit = |paper| × 26000e6 / 1e18
func at26000(paper *big.Int) *big.Int {
	c := new(big.Int).Mul(new(big.Int).Abs(paper), big.NewInt(26000e6))
	c.Quo(c, e18)
	if paper.Sign() > 0 {
		c.Neg(c)
	}
	return c
}

func do(t *testing.T, h http.Handler, ctx context.Context, method, path string, body any) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	switch b := body.(type) {
	case nil:
	case string:
		buf.WriteString(b)
	default:
		if err := json.NewEncoder(&buf).Encode(b); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequestWithContext(ctx, method, path, &buf)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var out map[string]any
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("response is not JSON: %q", rec.Body.String())
		}
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q", ct)
	}
	return rec, out
}

func post(t *testing.T, h http.Handler, body any) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	return do(t, h, context.Background(), http.MethodPost, "/order", body)
}

func wantStatus(t *testing.T, rec *httptest.ResponseRecorder, out map[string]any, code int) {
	t.Helper()
	if rec.Code != code {
		t.Fatalf("status = %d, want %d, body %v", rec.Code, code, out)
	}
	if code >= 400 {
		if s, _ := out["error"].(string); s == "" {
			t.Errorf("error response without message: %v", out)
		}
	}
}

// 贯穿例子：carol 卖 0.1 @26000，链上已成交 0.08 → 只挂 0.02；alice 买 0.05 → 吃 0.02，剩 0.03 挂单。
func TestPlaceOrderWalkthrough(t *testing.T) {
	env := newTestEnv(t)
	carol, alice := newTrader(t), newTrader(t)

	carolReq, carolID := sign(t, carol, orderSpec{signer: carol.addr, paper: amt(-100), credit: at26000(amt(-100))}, testChainID)
	env.fb.Filled[carolID] = amt(80)

	rec, out := post(t, env.h, carolReq)
	wantStatus(t, rec, out, http.StatusOK)
	if out["orderId"] != carolID.Hex() {
		t.Errorf("orderId = %v, want %s", out["orderId"], carolID.Hex())
	}
	if out["filled"] != "0" || out["remaining"] != amt(20).String() {
		t.Errorf("filled/remaining = %v/%v, want 0/%s", out["filled"], out["remaining"], amt(20))
	}
	if _, has := out["txHash"]; has {
		t.Errorf("no fill should omit txHash, got %v", out["txHash"])
	}
	if env.sub.calls != 0 {
		t.Errorf("no cross should not submit, calls=%d", env.sub.calls)
	}
	_, asks, _ := env.eng.Snapshot(testPerp, 5)
	if len(asks) != 1 || asks[0].Amount.Cmp(amt(20)) != 0 {
		t.Fatalf("asks = %+v, want 1 level of 0.02", asks)
	}

	aliceReq, aliceID := sign(t, alice, orderSpec{signer: alice.addr, paper: amt(50), credit: at26000(amt(50))}, testChainID)
	rec, out = post(t, env.h, aliceReq)
	wantStatus(t, rec, out, http.StatusOK)
	if out["orderId"] != aliceID.Hex() || out["filled"] != amt(20).String() || out["remaining"] != amt(30).String() {
		t.Errorf("alice resp = %v, want filled 0.02 remaining 0.03", out)
	}
	if out["txHash"] != fakeTxHash.Hex() {
		t.Errorf("txHash = %v, want %s", out["txHash"], fakeTxHash.Hex())
	}
	if env.sub.calls != 1 || len(env.sub.amounts) != 2 ||
		env.sub.amounts[0].Cmp(amt(20)) != 0 || env.sub.amounts[1].Cmp(amt(20)) != 0 {
		t.Errorf("submit amounts = %v", env.sub.amounts)
	}
	bids, asks, _ := env.eng.Snapshot(testPerp, 5)
	if len(asks) != 0 || len(bids) != 1 || bids[0].Amount.Cmp(amt(30)) != 0 {
		t.Errorf("after fill bids=%+v asks=%+v", bids, asks)
	}

	// 同一张单再 POST：引擎拒绝重复 → 422，不能再提交上链
	rec, out = post(t, env.h, aliceReq)
	wantStatus(t, rec, out, http.StatusUnprocessableEntity)
	if env.sub.calls != 1 {
		t.Errorf("duplicate should not submit, calls=%d", env.sub.calls)
	}
}

func TestPlaceOrderFullyFilledOnChain(t *testing.T) {
	env := newTestEnv(t)
	carol := newTrader(t)
	req, id := sign(t, carol, orderSpec{signer: carol.addr, paper: amt(-100), credit: at26000(amt(-100))}, testChainID)
	env.fb.Filled[id] = amt(100)
	rec, out := post(t, env.h, req)
	wantStatus(t, rec, out, http.StatusConflict)
	if _, asks, _ := env.eng.Snapshot(testPerp, 5); len(asks) != 0 {
		t.Errorf("fully filled order must not enter book: %+v", asks)
	}
}

func TestPlaceOrderOperator(t *testing.T) {
	env := newTestEnv(t)
	carol, bot := newTrader(t), newTrader(t)
	req, _ := sign(t, bot, orderSpec{signer: carol.addr, paper: amt(-100), credit: at26000(amt(-100))}, testChainID)

	rec, out := post(t, env.h, req)
	wantStatus(t, rec, out, http.StatusUnauthorized)
	if !env.fb.Called("isOperatorValid") {
		t.Error("recovered != signer should ask chain isOperatorValid")
	}

	env.fb.Operators[[2]common.Address{carol.addr, bot.addr}] = true
	rec, out = post(t, env.h, req)
	wantStatus(t, rec, out, http.StatusOK)

	// 参数顺序反了（operator 授权 client）不算
	env2 := newTestEnv(t)
	env2.fb.Operators[[2]common.Address{bot.addr, carol.addr}] = true
	rec, out = post(t, env2.h, req)
	wantStatus(t, rec, out, http.StatusUnauthorized)
}

func TestPlaceOrderSignerItselfSkipsOperatorCheck(t *testing.T) {
	env := newTestEnv(t)
	carol := newTrader(t)
	req, _ := sign(t, carol, orderSpec{signer: carol.addr, paper: amt(-100), credit: at26000(amt(-100))}, testChainID)
	rec, out := post(t, env.h, req)
	wantStatus(t, rec, out, http.StatusOK)
	if env.fb.Called("isOperatorValid") {
		t.Error("signer signs own order: no need to query operator")
	}
}

func TestPlaceOrderRejects(t *testing.T) {
	carol := newTrader(t)
	good, _ := sign(t, carol, orderSpec{signer: carol.addr, paper: amt(-100), credit: at26000(amt(-100))}, testChainID)
	expired, _ := sign(t, carol, orderSpec{signer: carol.addr, paper: amt(-100), credit: at26000(amt(-100)),
		expiration: uint64(time.Now().Add(-time.Minute).Unix())}, testChainID)
	wrongChain, _ := sign(t, carol, orderSpec{signer: carol.addr, paper: amt(-100), credit: at26000(amt(-100))}, 1)
	sameSign, _ := sign(t, carol, orderSpec{signer: carol.addr, paper: amt(-100), credit: big.NewInt(-2600e6)}, testChainID)

	mod := func(f func(*PlaceOrderRequest)) PlaceOrderRequest { r := good; f(&r); return r }
	sig := hexutil.MustDecode(good.Signature)

	tests := []struct {
		name string
		body any
		code int
	}{
		{"bad json", "{", http.StatusBadRequest},
		{"bad perp", mod(func(r *PlaceOrderRequest) { r.Perp = "0x1234" }), http.StatusBadRequest},
		{"unknown perp", mod(func(r *PlaceOrderRequest) { r.Perp = common.HexToAddress("0x99").Hex() }), http.StatusBadRequest},
		{"bad signer", mod(func(r *PlaceOrderRequest) { r.Signer = "carol" }), http.StatusBadRequest},
		// 多出一个字节：HexToAddress 会悄悄截成正确地址，只有 IsHexAddress 能挡住
		{"perp 21 bytes", mod(func(r *PlaceOrderRequest) { r.Perp = "0x00" + r.Perp[2:] }), http.StatusBadRequest},
		{"signer 21 bytes", mod(func(r *PlaceOrderRequest) { r.Signer = "0x00" + r.Signer[2:] }), http.StatusBadRequest},
		{"bad paper", mod(func(r *PlaceOrderRequest) { r.PaperAmount = "1.5" }), http.StatusBadRequest},
		{"bad credit", mod(func(r *PlaceOrderRequest) { r.CreditAmount = "" }), http.StatusBadRequest},
		{"info 31 bytes", mod(func(r *PlaceOrderRequest) { r.Info = hexutil.Encode(make([]byte, 31)) }), http.StatusBadRequest},
		{"info not hex", mod(func(r *PlaceOrderRequest) { r.Info = "zz" }), http.StatusBadRequest},
		{"sig not hex", mod(func(r *PlaceOrderRequest) { r.Signature = "0xzz" }), http.StatusBadRequest},
		{"sig 64 bytes", mod(func(r *PlaceOrderRequest) { r.Signature = hexutil.Encode(sig[:64]) }), http.StatusUnauthorized},
		{"tampered paper", mod(func(r *PlaceOrderRequest) { r.PaperAmount = amt(-200).String() }), http.StatusUnauthorized},
		{"wrong chain id", wrongChain, http.StatusUnauthorized},
		{"expired", expired, http.StatusBadRequest},
		{"paper and credit same sign", sameSign, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			rec, out := post(t, env.h, tt.body)
			wantStatus(t, rec, out, tt.code)
			if _, asks, _ := env.eng.Snapshot(testPerp, 5); len(asks) != 0 {
				t.Errorf("rejected order entered book")
			}
		})
	}
}

func TestPlaceOrderChainReadError(t *testing.T) {
	env := newTestEnv(t)
	carol := newTrader(t)
	req, _ := sign(t, carol, orderSpec{signer: carol.addr, paper: amt(-100), credit: at26000(amt(-100))}, testChainID)
	env.fb.CallErr = errNodeDown
	rec, out := post(t, env.h, req)
	wantStatus(t, rec, out, http.StatusBadGateway)
	if _, asks, _ := env.eng.Snapshot(testPerp, 5); len(asks) != 0 {
		t.Errorf("order entered book without knowing on-chain filled amount")
	}
}

func TestPlaceOrderSubmitFails(t *testing.T) {
	env := newTestEnv(t)
	carol, alice := newTrader(t), newTrader(t)
	carolReq, _ := sign(t, carol, orderSpec{signer: carol.addr, paper: amt(-100), credit: at26000(amt(-100))}, testChainID)
	rec, out := post(t, env.h, carolReq)
	wantStatus(t, rec, out, http.StatusOK)

	env.sub.err = errors.New("execution reverted: TRADER_NOT_SAFE")
	aliceReq, _ := sign(t, alice, orderSpec{signer: alice.addr, paper: amt(50), credit: at26000(amt(50))}, testChainID)
	rec, out = post(t, env.h, aliceReq)
	wantStatus(t, rec, out, http.StatusUnprocessableEntity)
	if s, _ := out["error"].(string); !strings.Contains(s, "TRADER_NOT_SAFE") {
		t.Errorf("error should carry revert reason, got %q", s)
	}
}

// 客户端断开（请求 ctx 取消）不能把上链 ctx 一起取消，否则可能"链上成交了、引擎没 Fill"。
func TestPlaceOrderSubmitCtxSurvivesClientCancel(t *testing.T) {
	env := newTestEnv(t)
	carol, alice := newTrader(t), newTrader(t)
	carolReq, _ := sign(t, carol, orderSpec{signer: carol.addr, paper: amt(-100), credit: at26000(amt(-100))}, testChainID)
	post(t, env.h, carolReq)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	aliceReq, _ := sign(t, alice, orderSpec{signer: alice.addr, paper: amt(50), credit: at26000(amt(50))}, testChainID)
	do(t, env.h, ctx, http.MethodPost, "/order", aliceReq)
	if env.sub.calls != 1 {
		t.Fatalf("submit calls = %d", env.sub.calls)
	}
	if env.sub.ctxErr != nil {
		t.Errorf("submit ctx already canceled: %v", env.sub.ctxErr)
	}
}
