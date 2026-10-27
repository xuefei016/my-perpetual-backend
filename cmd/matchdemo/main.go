package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
	"github.com/xuefei016/my-perpetual-backend/internal/config"
	"github.com/xuefei016/my-perpetual-backend/internal/engine"
)

type balance struct {
	Paper  *big.Int
	Credit *big.Int
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := must(config.Load())
	c := must(chain.New(ctx, cfg.RPCURL, cfg.Dealer, cfg.Markets[0].Perp, cfg.USDC))
	defer c.Client.Close()
	dsChain, err := c.Dealer.DomainSeparator(&bind.CallOpts{Context: ctx})
	if err != nil {
		panic(err)
	}
	if dsChain != chain.DomainSeparator(c.ChainId, cfg.Dealer) {
		panic("domain separator mismatch")
	}

	orderSenderKey := mustEnv("ORDER_SENDER_PK")
	aliceKey, bobKey, carolKey := mustEnv("ALICE_PK"), mustEnv("BOB_PK"), mustEnv("CAROL_PK")

	// 创建engine
	auth := must(c.Transactor(orderSenderKey))
	eng := engine.New(chain.NewTradeSender(c, auth), cfg.Markets[0].Perp)

	aliceAddress := must(privateKeyToAddress(aliceKey))
	bobAddress := must(privateKeyToAddress(bobKey))
	carolAddress := must(privateKeyToAddress(carolKey))
	beforeBalanceMap := must(readBalance(ctx, c, aliceAddress, bobAddress, carolAddress))

	base := uint64(time.Now().UnixNano())
	head := must(c.Client.HeaderByNumber(ctx, nil))

	maxTime := max(head.Time, uint64(time.Now().Unix()))
	o1 := must(sign(c, cfg.Dealer, cfg.Markets[0].Perp, bobKey, big.NewInt(-1e17), big.NewInt(2600e6), base+1, maxTime+3600))
	o2 := must(sign(c, cfg.Dealer, cfg.Markets[0].Perp, aliceKey, big.NewInt(-1e17), big.NewInt(2600e6), base+2, maxTime+3600))
	o3 := must(sign(c, cfg.Dealer, cfg.Markets[0].Perp, carolKey, big.NewInt(-1e17), big.NewInt(2600e6), base+3, maxTime+3600))
	o4 := must(sign(c, cfg.Dealer, cfg.Markets[0].Perp, aliceKey, big.NewInt(1.8e17), big.NewInt(-4680e6), base+4, maxTime+3600))

	orders := []*engine.BookOrder{o1, o2, o3, o4}
	var last *engine.Match
	for i, o := range orders {
		m, err := eng.PlaceOrder(ctx, o)
		if err != nil {
			panic(fmt.Errorf("place #%d: %w", i+1, err))
		}
		fmt.Printf("#%d %s fills=%d tx=%s\n", i+1, o.ID[:10], len(m.Fills), m.TxHash)
		for _, f := range m.Fills {
			fmt.Printf("    吃 %s %s paper=%s\n", f.Maker.ID[:10], f.Maker.Order.Signer.Hex()[:6], f.Amount)
		}
		last = m
	}
	afterBalanceMap := must(readBalance(ctx, c, aliceAddress, bobAddress, carolAddress))
	check := func(name string, got, want *big.Int) {
		if got.Cmp(want) != 0 {
			panic(fmt.Sprintf("%s: got %s want %s", name, got, want))
		}
	}

	// ① 期望的仓位变化从引擎的 Match 推导，公式同合约 _matchOrders（费率为 0）：
	//    maker: paperChange = ±Amount（符号随 maker 方向），creditChange = paperChange * credit / paper
	//    taker: 所有 maker 变化取反求和
	// Solidity 的 / 向零截断，对应 big.Int 的 Quo（Div 是欧几里得除法，负数除不尽时差 1）
	want := map[common.Address]balance{}
	for _, a := range []common.Address{aliceAddress, bobAddress, carolAddress} {
		want[a] = balance{Paper: new(big.Int), Credit: new(big.Int)}
	}
	taker := last.Taker.Order.Signer
	for _, f := range last.Fills {
		mo := f.Maker.Order
		paperChange := new(big.Int).Set(f.Amount)
		if f.Maker.Side == engine.Sell {
			paperChange.Neg(paperChange)
		}
		creditChange := new(big.Int).Mul(paperChange, mo.CreditAmount)
		creditChange.Quo(creditChange, mo.PaperAmount)
		want[mo.Signer].Paper.Add(want[mo.Signer].Paper, paperChange)
		want[mo.Signer].Credit.Add(want[mo.Signer].Credit, creditChange)
		want[taker].Paper.Sub(want[taker].Paper, paperChange)
		want[taker].Credit.Sub(want[taker].Credit, creditChange)
	}
	for _, a := range []common.Address{aliceAddress, bobAddress, carolAddress} {
		before, after := beforeBalanceMap[a], afterBalanceMap[a]
		check(a.Hex()[:6]+" paper", new(big.Int).Sub(after.Paper, before.Paper), want[a].Paper)
		check(a.Hex()[:6]+" credit", new(big.Int).Sub(after.Credit, before.Credit), want[a].Credit)
	}

	// ② 引擎 Remaining + 链上已成交 == |paper|：引擎和链上 orderFilledPaperAmount 是同一本账
	opts := &bind.CallOpts{Context: ctx}
	for _, o := range orders {
		filled, err := c.Dealer.GetOrderFilledAmount(opts, common.HexToHash(o.ID))
		if err != nil {
			panic(err)
		}
		check("remaining+filled "+o.ID[:10], new(big.Int).Add(o.Remaining, filled), new(big.Int).Abs(o.Order.PaperAmount))
		fmt.Printf("%s remaining=%s filled=%s\n", o.ID[:10], o.Remaining, filled)
	}

	bids, asks, err := eng.Snapshot(cfg.Markets[0].Perp, 0)
	if err != nil {
		panic(err)
	}
	fmt.Printf("bids=%v\nasks=%v\n", levels(bids), levels(asks))
	fmt.Println("reconciled ✓")
}

func readBalance(ctx context.Context, c *chain.Chain, traders ...common.Address) (map[common.Address]balance, error) {
	opts := &bind.CallOpts{Context: ctx}
	balanceMap := make(map[common.Address]balance, len(traders))
	for _, trader := range traders {
		result, err := c.BTC.BalanceOf(opts, trader)
		if err != nil {
			return nil, err
		}
		balanceMap[trader] = balance{
			Paper:  result.Paper,
			Credit: result.Credit,
		}
	}
	return balanceMap, nil
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("environment variable %q not set", key))
	}
	return v
}

func privateKeyToAddress(privateKey string) (common.Address, error) {
	if privateKey == "" {
		return common.Address{}, fmt.Errorf("private key is empty")
	}
	key, err := crypto.HexToECDSA(strings.TrimPrefix(privateKey, "0x"))
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(key.PublicKey), nil
}

func sign(c *chain.Chain, dealer, perp common.Address, keyHex string, paper, credit *big.Int, nonce uint64, expiration uint64) (*engine.BookOrder, error) {
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(keyHex, "0x"))
	if err != nil {
		return nil, err
	}
	trader := crypto.PubkeyToAddress(privateKey.PublicKey)
	oi := chain.OrderInfo{
		MakerFeeRate: 0,
		TakerFeeRate: 0,
		Expiration:   expiration,
		Nonce:        nonce,
	}
	o := chain.Order{
		Perp:         perp,
		Signer:       trader,
		PaperAmount:  paper,
		CreditAmount: credit,
		Info:         chain.PackInfo(oi),
	}

	ds := chain.DomainSeparator(c.ChainId, dealer)
	digest := chain.Digest(ds, o)
	sig, err := chain.SignOrder(privateKey, digest)
	if err != nil {
		return nil, err
	}
	got, err := chain.RecoverSigner(digest, sig)
	if err != nil || got != trader {
		return nil, fmt.Errorf("recover signer: got %s want %s, err %v", got, trader, err)
	}
	return engine.NewBookOrder(digest.Hex(), o, sig)
}

// 按合约 _matchOrders 公式，从引擎撮合结果推出每个人的期望变化
func expectedDeltas(taker *engine.BookOrder, fills []engine.Fill) map[common.Address][2]*big.Int {
	want := map[common.Address][2]*big.Int{}
	add := func(who common.Address, p, c *big.Int) {
		cur, ok := want[who]
		if !ok {
			cur = [2]*big.Int{new(big.Int), new(big.Int)}
			want[who] = cur
		}
		cur[0].Add(cur[0], p)
		cur[1].Add(cur[1], c)
	}

	for _, f := range fills {
		mo := f.Maker.Order
		paper := new(big.Int).Set(f.Amount)
		if mo.PaperAmount.Sign() < 0 { // maker 卖
			paper.Neg(paper)
		}
		credit := new(big.Int).Mul(paper, mo.CreditAmount)
		credit.Quo(credit, mo.PaperAmount) // Quo = 向零截断，和 Solidity 的 / 一致；不能用 Div

		add(mo.Signer, paper, credit)
		add(taker.Order.Signer, new(big.Int).Neg(paper), new(big.Int).Neg(credit))
	}
	return want
}

func levels(ls []engine.Level) []string {
	var out []string
	for _, l := range ls {
		out = append(out, fmt.Sprintf("%s:%s(%d)", l.Price, l.Amount, l.Count))
	}
	return out
}
