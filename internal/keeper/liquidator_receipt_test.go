package keeper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/xuefei016/my-perpetual-backend/internal/chaintest"
)

type receiptTransport func(*http.Request) (*http.Response, error)

func (f receiptTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// NoSend 阻止广播；假 HTTP transport 只返回回执，真实 WaitSuccess 原样运行。
func TestLiquidateWaitsForReceipt(t *testing.T) {
	for _, outcome := range []string{"success", "reverted", "parent canceled"} {
		t.Run(outcome, func(t *testing.T) {
			l, fb, _ := newLiquidationEnv(t)
			fb.Accounts[liqTrader] = chaintest.Rich()
			fb.Accounts[liqTrader].Paper = big.NewInt(5e17)
			fb.LiqCost = func(_ common.Address, _ common.Address, paper *big.Int) (*big.Int, *big.Int, error) {
				return new(big.Int).Set(paper), usd(-11880), nil
			}
			l.auth.NoSend, l.auth.GasLimit = true, 500000
			var expectedHash common.Hash
			l.auth.Signer = func(_ common.Address, tx *types.Transaction) (*types.Transaction, error) {
				expectedHash = tx.Hash()
				return tx, nil
			}
			ctx, cancel := context.WithCancel(context.WithValue(context.Background(), liquidationTraceKey{}, t.Name()))
			defer cancel()
			var waitCtx context.Context
			transport := receiptTransport(func(req *http.Request) (*http.Response, error) {
				waitCtx = req.Context()
				assertLiquidationChildContext(t, ctx, waitCtx, 60*time.Second)
				if deadline, ok := waitCtx.Deadline(); ok && time.Until(deadline) < 55*time.Second {
					t.Error("receipt timeout was shortened below the configured 60s")
				}
				if waitCtx.Err() != nil {
					t.Errorf("receipt reused canceled send context: %v", waitCtx.Err())
				}
				if outcome == "parent canceled" {
					cancel()
					return nil, waitCtx.Err()
				}
				var request struct {
					ID     json.RawMessage
					Method string
					Params []common.Hash
				}
				if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
					t.Fatal(err)
				}
				if request.Method != "eth_getTransactionReceipt" || len(request.Params) != 1 || request.Params[0] != expectedHash {
					t.Fatalf("unexpected receipt request: %+v", request)
				}
				status := uint64(types.ReceiptStatusSuccessful)
				if outcome == "reverted" {
					status = types.ReceiptStatusFailed
				}
				receipt := &types.Receipt{TxHash: expectedHash, Status: status, BlockNumber: big.NewInt(1), Logs: []*types.Log{}}
				body, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": receipt})
				if err != nil {
					t.Fatal(err)
				}
				return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(bytes.NewReader(body))}, nil
			})
			client, err := rpc.DialOptions(ctx, "http://receipt.invalid", rpc.WithHTTPClient(&http.Client{Transport: transport}))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			l.c.Client = ethclient.NewClient(client)
			err = l.liquidate(ctx, liqTrader)
			if waitCtx == nil {
				t.Fatal("liquidation never checked receipt")
			}
			if outcome == "success" && err != nil {
				t.Fatalf("successful receipt: %v", err)
			}
			if outcome != "success" && (err == nil || !strings.Contains(err.Error(), expectedHash.Hex()) || !strings.Contains(err.Error(), "wait liquidation")) {
				t.Fatalf("want wait error with hash, got %v", err)
			}
			if outcome == "parent canceled" && !errors.Is(err, context.Canceled) {
				t.Errorf("cancellation cause lost: %v", err)
			}
			if !errors.Is(waitCtx.Err(), context.Canceled) {
				t.Errorf("receipt context not released: %v", waitCtx.Err())
			}
		})
	}
}
