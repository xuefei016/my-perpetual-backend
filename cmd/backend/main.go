// Command backend 组装并启动整个后端：配置 → 连链 → 撮合引擎 → keeper → HTTP。
// 用法：ORDER_SENDER_PK=… KEEPER_PK=… LIQUIDATOR_PK=… go run ./cmd/backend
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xuefei016/my-perpetual-backend/internal/api"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
	"github.com/xuefei016/my-perpetual-backend/internal/config"
	"github.com/xuefei016/my-perpetual-backend/internal/engine"
	"github.com/xuefei016/my-perpetual-backend/internal/keeper"
)

func main() {
	// Ctrl-C / kill 时取消 ctx，keeper 和 HTTP 一起退出
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	c, err := chain.New(ctx, cfg.RPCURL, cfg.Dealer, cfg.Markets[0].Perp, cfg.USDC)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Client.Close()

	orderSender, err := c.Transactor(cfg.OrderSenderKey)
	if err != nil {
		log.Fatal("order sender: ", err)
	}
	keeperAuth, err := c.Transactor(cfg.KeeperKey)
	if err != nil {
		log.Fatal("keeper: ", err)
	}
	liquidatorAuth, err := c.Transactor(cfg.LiquidatorKey)
	if err != nil {
		log.Fatal("liquidator: ", err)
	}

	perps := make([]common.Address, len(cfg.Markets))
	for i, m := range cfg.Markets {
		perps[i] = m.Perp
	}

	eng := engine.New(chain.NewTradeSender(c, orderSender), perps...)
	funding := keeper.NewFundingKeeper(c, keeperAuth, perps, cfg.FundingInterval)
	liquidator := keeper.NewLiquidator(c, liquidatorAuth, perps[0], eng, cfg.LiquidationInterval)

	// 协程1: 用于更新资金费率
	go func() {
		if err := funding.Run(ctx); err != nil {
			log.Println("funding keeper:", err)
		}
	}()

	// 协程2：清算机器人用于清算
	go func() {
		if err := liquidator.Run(ctx); err != nil {
			log.Println("liquidator:", err)
		}
	}()

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: api.New(eng, c, cfg.Dealer).Handler()}
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	log.Println("listening on", cfg.HTTPAddr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
