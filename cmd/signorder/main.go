// Command signorder 是 D7 的链上验收工具：用 Go 给 alice(#4) 和 bob(#5) 各签一张单，
// 打印 cast 编码 tradeData 需要的参数，再由 orderSender(#1) 用 cast send 调 btc.trade。
// 成交成功 = Go 算的 digest 与合约 approveTrade 重算的逐字节一致。
// 用法：ALICE_PK=… BOB_PK=… go run ./cmd/signorder
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/xuefei016/my-perpetual-backend/internal/api"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
	"github.com/xuefei016/my-perpetual-backend/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	c, err := chain.New(ctx, cfg.RPCURL, cfg.Dealer, cfg.Markets[0].Perp, cfg.USDC)
	if err != nil {
		panic(err)
	}
	defer c.Client.Close()

	dsOnchain, err := c.Dealer.DomainSeparator(&bind.CallOpts{Context: ctx})
	if err != nil {
		panic(err)
	}
	dsOnlocal := chain.DomainSeparator(c.ChainId, cfg.Dealer)
	if dsOnchain != dsOnlocal {
		panic("ds_onchain != ds_onlocal")
	}

	head, err := c.Client.HeaderByNumber(ctx, nil)
	if err != nil {
		panic(err)
	}
	base := uint64(time.Now().UnixNano())
	maxTime := max(head.Time, uint64(time.Now().Unix()))
	oi := chain.OrderInfo{
		MakerFeeRate: 0,
		TakerFeeRate: 0,
		Expiration:   maxTime + 3600,
	}

	oi.Nonce = base + 1
	aliceInfo := chain.PackInfo(oi)
	aliceKey, err := crypto.HexToECDSA(strings.TrimPrefix(os.Getenv("ALICE_PK"), "0x"))
	if err != nil {
		panic(err)
	}
	aliceOrder := chain.Order{
		Perp:         cfg.Markets[0].Perp,
		Signer:       crypto.PubkeyToAddress(aliceKey.PublicKey),
		PaperAmount:  big.NewInt(1e17),
		CreditAmount: big.NewInt(-2600e6),
		Info:         aliceInfo,
	}

	oi.Nonce = base + 2
	bobInfo := chain.PackInfo(oi)
	bobKey, err := crypto.HexToECDSA(strings.TrimPrefix(os.Getenv("BOB_PK"), "0x"))
	if err != nil {
		panic(err)
	}
	bobOrder := chain.Order{
		Perp:         cfg.Markets[0].Perp,
		Signer:       crypto.PubkeyToAddress(bobKey.PublicKey),
		PaperAmount:  big.NewInt(-1e17),
		CreditAmount: big.NewInt(2600e6),
		Info:         bobInfo,
	}

	aliceOrderHash := chain.Digest(dsOnlocal, aliceOrder)
	bobOrderHash := chain.Digest(dsOnlocal, bobOrder)

	aliceSig, err := chain.SignOrder(aliceKey, aliceOrderHash)
	if err != nil {
		panic(err)
	}

	aliceRecoverAddress, err := chain.RecoverSigner(aliceOrderHash, aliceSig)
	if err != nil {
		panic(err)
	}
	if aliceRecoverAddress != aliceOrder.Signer {
		panic("alice invalid recover address")
	}

	bobSig, err := chain.SignOrder(bobKey, bobOrderHash)
	if err != nil {
		panic(err)
	}
	bobRecoverAddress, err := chain.RecoverSigner(bobOrderHash, bobSig)
	if err != nil {
		panic(err)
	}
	if bobRecoverAddress != bobOrder.Signer {
		panic("bob invalid recover address")
	}

	fmt.Printf("ORDERS='[%s,%s]'\n", tuple(aliceOrder), tuple(bobOrder))
	fmt.Printf("SIGS='[%s,%s]'\n", hexutil.Encode(aliceSig), hexutil.Encode(bobSig))
	fmt.Printf("AMOUNTS='[%s,%s]'\n", new(big.Int).Abs(aliceOrder.PaperAmount), new(big.Int).Abs(bobOrder.PaperAmount))

	aliceJSON, err := json.MarshalIndent(orderRequest(aliceOrder, aliceSig), "", "  ")
	if err != nil {
		panic(err)
	}
	bobJSON, err := json.MarshalIndent(orderRequest(bobOrder, bobSig), "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("ALICE_ORDER_JSON=%s\n", aliceJSON)
	fmt.Printf("BOB_ORDER_JSON=%s\n", bobJSON)
}

func tuple(o chain.Order) string {
	return fmt.Sprintf("(%s,%s,%s,%s,%s)",
		o.Perp.Hex(),
		o.Signer.Hex(),
		o.PaperAmount.String(),
		o.CreditAmount.String(),
		hexutil.Encode(o.Info[:]))
}

func orderRequest(o chain.Order, sig []byte) api.PlaceOrderRequest {
	return api.PlaceOrderRequest{
		Perp:         o.Perp.Hex(),
		Signer:       o.Signer.Hex(),
		PaperAmount:  o.PaperAmount.String(),
		CreditAmount: o.CreditAmount.String(),
		Info:         hexutil.Encode(o.Info[:]),
		Signature:    hexutil.Encode(sig),
	}
}
