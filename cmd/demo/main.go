// Command demo 是 D5 的演示：读风控参数和仓位，alice(#4) approve + deposit 1000 USDC。
// 用法：TRADER_PK=<#4 私钥> go run ./cmd/demo
package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
)

func main() {
	const rpcURL = "http://127.0.0.1:8545"

	// 全新 anvil 上跑 Deploy.s.sol 得到的确定性地址
	var (
		dealerAddr = common.HexToAddress("0x5FC8d32690cc91D4c39d9d3abcBD16989F875707")
		btcAddr    = common.HexToAddress("0xB7f8BC63BbcaD18155201308C8f3540b07f84F5e")
		usdcAddr   = common.HexToAddress("0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9")
		alice      = common.HexToAddress("0x15d34AAf54267DB7D7c367839AAf71A00a2C6A65") // anvil #4
		carol      = common.HexToAddress("0x976EA74026E726554dB657fA54763abd0C3a0aa9") // anvil #6
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := chain.New(ctx, rpcURL, dealerAddr, btcAddr, usdcAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Client.Close()

	fmt.Println(c.ChainId)
	opts := &bind.CallOpts{Context: ctx}
	p, err := c.Dealer.GetRiskParams(opts, btcAddr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s  IM %s  MM %s  折扣 %s  保险费 %s  价格源 %s  registered=%v\n",
		p.Name, formatRatio(p.InitialMarginRatio), formatRatio(p.LiquidationThreshold),
		formatRatio(p.LiquidationPriceOff), formatRatio(p.InsuranceFeeRate), p.MarkPriceSource.Hex(), p.IsRegistered)
	aliceAddressArray, err := c.Dealer.GetPositions(opts, alice)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("alice position: ", aliceAddressArray)
	carolAddressArray, err := c.Dealer.GetPositions(opts, carol)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("carol position: ", carolAddressArray)

	primaryCredit, err := c.Dealer.GetPrimaryCredit(opts, alice)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("alice primary credit: ", formatUSDC(primaryCredit))

	amount := big.NewInt(1_000_000_000)
	walletBefore, err := c.USDC.BalanceOf(opts, alice)
	if err != nil {
		log.Fatal(err)
	}
	creditBefore, err := c.Dealer.GetPrimaryCredit(opts, alice)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("alice wallet before:", formatUSDC(walletBefore), "credit before:", formatUSDC(creditBefore))

	auth, err := c.Transactor(os.Getenv("TRADER_PK"))
	if err != nil {
		log.Fatal(err)
	}

	approveTx, err := c.USDC.Approve(auth, dealerAddr, amount)
	if err != nil {
		log.Fatal(err)
	}
	recipt, err := chain.WaitSuccess(ctx, c, approveTx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("approve tx", recipt.TxHash.Hex(), "gas", recipt.GasUsed, "block", recipt.BlockNumber)
	depositTx, err := c.Dealer.Deposit(auth, amount, big.NewInt(0), alice)
	if err != nil {
		log.Fatal(err)
	}

	recipt, err = chain.WaitSuccess(ctx, c, depositTx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("deposit tx", recipt.TxHash.Hex())

	walletAfter, err := c.USDC.BalanceOf(opts, alice)
	if err != nil {
		log.Fatal(err)
	}
	creditAfter, err := c.Dealer.GetPrimaryCredit(opts, alice)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("alice wallet after:", formatUSDC(walletAfter), "credit after:", formatUSDC(creditAfter))
}

func formatFixed(x *big.Int, decimal int) string {
	unit := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimal)), nil)
	abs := new(big.Int).Abs(x)
	intPart, fracPart := new(big.Int).QuoRem(abs, unit, new(big.Int))
	frac := fracPart.String()
	frac = strings.Repeat("0", decimal-len(frac)) + frac
	frac = strings.TrimRight(frac, "0")
	s := intPart.String()
	if frac != "" {
		s += "." + frac
	}

	if x.Sign() < 0 {
		s = "-" + s
	}
	return s
}

func formatUSDC(x *big.Int) string {
	return formatFixed(x, 6)
}

func formatRatio(x *big.Int) string {
	return formatFixed(x, 16) + "%"
}
