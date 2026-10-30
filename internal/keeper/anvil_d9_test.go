package keeper

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/xuefei016/my-perpetual-backend/internal/api"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
	"github.com/xuefei016/my-perpetual-backend/internal/config"
	"github.com/xuefei016/my-perpetual-backend/internal/engine"
)

// 显式开启；在 TestD9AnvilHTTPTrade 已成交且尚未更新资金费的本地链上执行一次真实 tick。
func TestD9AnvilFunding(t *testing.T) {
	if os.Getenv("D9_ANVIL_FUNDING") != "1" {
		t.Skip("requires explicit D9 anvil funding fixture")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	cfg := config.Default()
	perp := cfg.Markets[0].Perp
	c, err := chain.New(ctx, cfg.RPCURL, cfg.Dealer, perp, cfg.USDC)
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
	opts := &bind.CallOpts{Context: ctx}
	oldRate, err := c.BTC.GetFundingRate(opts)
	if err != nil || oldRate.Sign() != 0 {
		t.Fatalf("requires initial fundingRate=0: %v %v", oldRate, err)
	}
	mark, err := c.MarkPrice(opts, perp)
	if err != nil || mark.Cmp(usd(26000)) != 0 {
		t.Fatalf("requires mark 26000: %v %v", mark, err)
	}
	orderAuth, err := c.Transactor(os.Getenv("ORDER_SENDER_PK"))
	if err != nil {
		t.Fatal(err)
	}
	keeperAuth, err := c.Transactor(os.Getenv("KEEPER_PK"))
	if err != nil {
		t.Fatal(err)
	}
	eng := engine.New(chain.NewTradeSender(c, orderAuth), perp)
	srv := httptest.NewServer(api.New(eng, c, cfg.Dealer).Handler())
	defer srv.Close()
	client := &http.Client{Timeout: 30 * time.Second}
	var traders [2]common.Address
	var beforePaper, beforeCredit [2]*big.Int
	for i, actor := range []struct {
		name, keyEnv  string
		paper, credit int64
	}{
		{"alice", "ALICE_PK", 1e16, -260e6}, // bid 26000
		{"bob", "BOB_PK", -1e16, 262e6},     // ask 26200
	} {
		key, err := crypto.HexToECDSA(strings.TrimPrefix(os.Getenv(actor.keyEnv), "0x"))
		if err != nil {
			t.Fatal(err)
		}
		traders[i] = crypto.PubkeyToAddress(key.PublicKey)
		before, err := c.BTC.BalanceOf(opts, traders[i])
		if err != nil || before.Paper.Sign() == 0 {
			t.Fatalf("requires existing HTTP trade position: %v %v", before, err)
		}
		beforePaper[i], beforeCredit[i] = before.Paper, before.Credit
		order := chain.Order{Perp: perp, Signer: traders[i], PaperAmount: big.NewInt(actor.paper), CreditAmount: big.NewInt(actor.credit),
			Info: chain.PackInfo(chain.OrderInfo{Expiration: uint64(time.Now().Add(time.Hour).Unix()), Nonce: uint64(time.Now().UnixNano())})}
		sig, err := chain.SignOrder(key, chain.Digest(chain.DomainSeparator(c.ChainId, cfg.Dealer), order))
		if err != nil {
			t.Fatal(err)
		}
		body, err := json.Marshal(api.PlaceOrderRequest{Perp: perp.Hex(), Signer: traders[i].Hex(), PaperAmount: order.PaperAmount.String(),
			CreditAmount: order.CreditAmount.String(), Info: hexutil.Encode(order.Info[:]), Signature: hexutil.Encode(sig)})
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequestWithContext(ctx, "POST", srv.URL+"/order", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		var result api.PlaceOrderResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if err != nil || resp.StatusCode != 200 || result.Filled != "0" || result.TxHash != "" {
			t.Fatalf("book order %s: status=%d result=%+v err=%v", actor.name, resp.StatusCode, result, err)
		}
		if dir := os.Getenv("D9_REQUESTS_DIR"); dir != "" {
			if err := os.WriteFile(filepath.Join(dir, actor.name+"-quote.json"), body, 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
	bids, asks, err := eng.Snapshot(perp, 1)
	if err != nil || len(bids) != 1 || len(asks) != 1 || bids[0].Price.Cmp(usd(26000)) != 0 || asks[0].Price.Cmp(usd(26200)) != 0 {
		t.Fatalf("unexpected book: %v %v %v", bids, asks, err)
	}
	k := NewFundingKeeper(c, keeperAuth, []common.Address{perp}, eng, 15*time.Second)
	if err := k.tick(ctx); err != nil {
		t.Fatal(err)
	}
	rate, err := c.BTC.GetFundingRate(opts)
	// 独立期望值：−100e6 × 15 / (8×3600)，向零截断，单位 USDC 6 位/张。
	delta := big.NewInt(-52083)
	if err != nil || rate.Cmp(new(big.Int).Add(oldRate, delta)) != 0 {
		t.Fatalf("fundingRate=%v err=%v, want -52083", rate, err)
	}
	for i, trader := range traders {
		resp, err := client.Get(srv.URL + "/position?perp=" + perp.Hex() + "&trader=" + trader.Hex())
		if err != nil {
			t.Fatal(err)
		}
		var position api.PositionResponse
		err = json.NewDecoder(resp.Body).Decode(&position)
		resp.Body.Close()
		change := new(big.Int).Quo(new(big.Int).Mul(beforePaper[i], delta), big.NewInt(1e18))
		want := new(big.Int).Add(beforeCredit[i], change)
		if err != nil || resp.StatusCode != 200 || position.Paper != beforePaper[i].String() || position.Credit != want.String() {
			t.Fatalf("funding position=%+v want credit=%s err=%v", position, want, err)
		}
		t.Logf("funding signer=%s paper=%s credit=%s deltaCredit=%s", trader.Hex(), position.Paper, position.Credit, change)
	}
	block, err := c.Client.BlockByNumber(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, tx := range block.Transactions() {
		from, err := types.Sender(types.LatestSignerForChainID(c.ChainId), tx)
		if err != nil || from != keeperAuth.From {
			continue
		}
		receipt, err := c.Client.TransactionReceipt(ctx, tx.Hash())
		if err != nil || receipt.Status != types.ReceiptStatusSuccessful {
			t.Fatalf("funding receipt: %v %v", receipt, err)
		}
		found = true
		t.Logf("D9 funding tx=%s block=%s status=%d rate=%s", tx.Hash(), receipt.BlockNumber, receipt.Status, rate)
	}
	if !found {
		t.Fatal("keeper transaction not found in latest block")
	}
}
