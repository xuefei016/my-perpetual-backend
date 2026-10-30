package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
	"github.com/xuefei016/my-perpetual-backend/internal/config"
	"github.com/xuefei016/my-perpetual-backend/internal/engine"
)

// 显式启用才会提交本地交易；需要新部署、已存款但尚未开仓的 D9 anvil，
// D9_REQUESTS_DIR 指向本次 signorder 生成的 alice-order.json / bob-order.json。
func TestD9AnvilHTTPTrade(t *testing.T) {
	if os.Getenv("D9_ANVIL_HTTP") != "1" {
		t.Skip("requires explicit D9 anvil fixture")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	cfg := config.Default()
	c, err := chain.New(ctx, cfg.RPCURL, cfg.Dealer, cfg.Markets[0].Perp, cfg.USDC)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Client.Close()
	var version string
	if err := c.Client.Client().CallContext(ctx, &version, "web3_clientVersion"); err != nil {
		t.Fatal(err)
	}
	if c.ChainId.Cmp(big.NewInt(31337)) != 0 || !strings.Contains(strings.ToLower(version), "anvil") {
		t.Fatal("requires anvil chain 31337")
	}
	auth, err := c.Transactor(os.Getenv("ORDER_SENDER_PK"))
	if err != nil {
		t.Fatal(err)
	}
	perp := cfg.Markets[0].Perp
	eng := engine.New(chain.NewTradeSender(c, auth), perp)
	srv := httptest.NewServer(New(eng, c, cfg.Dealer).Handler())
	defer srv.Close()
	client := &http.Client{Timeout: 35 * time.Second}
	exchange := func(method, path string, body []byte, out any) int {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, method, srv.URL+path, bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if out != nil {
			if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
				t.Fatal(err)
			}
		} else if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			t.Fatal(err)
		}
		return resp.StatusCode
	}
	type signedRequest struct {
		body                      []byte
		request                   PlaceOrderRequest
		paper, credit             *big.Int
		digest                    common.Hash
		beforePaper, beforeCredit *big.Int
	}
	var orders [2]signedRequest // alice 买 / bob 卖
	for i, name := range []string{"alice", "bob"} {
		o := &orders[i]
		o.body, err = os.ReadFile(filepath.Join(os.Getenv("D9_REQUESTS_DIR"), name+"-order.json"))
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(o.body, &o.request); err != nil {
			t.Fatal(err)
		}
		var ok bool
		o.paper, ok = new(big.Int).SetString(o.request.PaperAmount, 10)
		if !ok {
			t.Fatal("invalid paper")
		}
		o.credit, ok = new(big.Int).SetString(o.request.CreditAmount, 10)
		if !ok {
			t.Fatal("invalid credit")
		}
		info, err := hexutil.Decode(o.request.Info)
		if err != nil || len(info) != 32 {
			t.Fatal("invalid info")
		}
		var packed [32]byte
		copy(packed[:], info)
		signer := common.HexToAddress(o.request.Signer)
		o.digest = chain.Digest(chain.DomainSeparator(c.ChainId, cfg.Dealer), chain.Order{
			Perp: perp, Signer: signer, PaperAmount: o.paper, CreditAmount: o.credit, Info: packed,
		})
		before, err := c.BTC.BalanceOf(&bind.CallOpts{Context: ctx}, signer)
		if err != nil {
			t.Fatal(err)
		}
		if before.Paper.Sign() != 0 {
			t.Fatal("requires fresh accounts without existing positions")
		}
		o.beforePaper, o.beforeCredit = before.Paper, before.Credit
	}
	if orders[0].paper.Sign() <= 0 || orders[1].paper.Sign() >= 0 ||
		new(big.Int).Add(orders[0].paper, orders[1].paper).Sign() != 0 ||
		new(big.Int).Add(orders[0].credit, orders[1].credit).Sign() != 0 {
		t.Fatal("requires equal-size opposite orders at the same price")
	}
	amount := new(big.Int).Abs(orders[0].paper).String()
	var maker, taker PlaceOrderResponse
	if status := exchange("POST", "/order", orders[1].body, &maker); status != 200 || maker.Filled != "0" || maker.Remaining != amount || maker.TxHash != "" || maker.OrderID != orders[1].digest.Hex() {
		t.Fatalf("maker status=%d response=%+v", status, maker)
	}
	var book OrderBookResponse
	if status := exchange("GET", "/orderbook?perp="+perp.Hex(), nil, &book); status != 200 || len(book.Asks) != 1 || book.Asks[0].Amount != amount || len(book.Bids) != 0 {
		t.Fatalf("maker book status=%d response=%+v", status, book)
	}
	if status := exchange("POST", "/order", orders[0].body, &taker); status != 200 || taker.Filled != amount || taker.Remaining != "0" || taker.TxHash == "" || taker.OrderID != orders[0].digest.Hex() {
		t.Fatalf("taker status=%d response=%+v", status, taker)
	}
	receipt, err := c.Client.TransactionReceipt(ctx, common.HexToHash(taker.TxHash))
	if err != nil || receipt.Status != types.ReceiptStatusSuccessful {
		t.Fatalf("trade receipt: %v %v", receipt, err)
	}
	for _, o := range orders {
		var pos PositionResponse
		if status := exchange("GET", "/position?perp="+perp.Hex()+"&trader="+o.request.Signer, nil, &pos); status != 200 {
			t.Fatalf("position status=%d", status)
		}
		wantPaper := new(big.Int).Add(o.beforePaper, o.paper)
		wantCredit := new(big.Int).Add(o.beforeCredit, o.credit)
		if pos.Paper != wantPaper.String() || pos.Credit != wantCredit.String() || !pos.Safe {
			t.Fatalf("position mismatch: got %+v want paper=%s credit=%s", pos, wantPaper, wantCredit)
		}
		filled, err := c.Dealer.GetOrderFilledAmount(&bind.CallOpts{Context: ctx}, o.digest)
		if err != nil || filled.String() != amount {
			t.Fatalf("chain filled=%v err=%v, want %s", filled, err, amount)
		}
		t.Logf("position signer=%s paper=%s credit=%s primary=%s MM-safe=%v filled=%s", pos.Trader, pos.Paper, pos.Credit, pos.PrimaryCredit, pos.Safe, filled)
	}
	if status := exchange("GET", "/orderbook?perp="+perp.Hex(), nil, &book); status != 200 || len(book.Asks) != 0 || len(book.Bids) != 0 {
		t.Fatalf("book not empty after trade: status=%d %+v", status, book)
	}
	if status := exchange("POST", "/order", orders[0].body, nil); status != http.StatusConflict {
		t.Fatalf("fully filled order replay status=%d, want 409", status)
	}
	t.Logf("D9 HTTP trade tx=%s block=%s status=%d", taker.TxHash, receipt.BlockNumber, receipt.Status)
}
