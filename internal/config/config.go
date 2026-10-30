// Package config 汇总后端运行需要的全部配置：RPC、合约地址、各角色私钥、轮询间隔。
// 地址默认值是全新 anvil 上跑 Deploy.s.sol 得到的确定性地址；私钥只从环境变量读，不写进代码。
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Market 一个永续市场：名字只用于展示，Perp 是该市场 Perpetual 合约地址。
type Market struct {
	Name string
	Perp common.Address
}

type Config struct {
	RPCURL   string // 例 "http://127.0.0.1:8545"
	HTTPAddr string // API 监听地址，例 ":8080"

	Dealer  common.Address // EIP-712 的 verifyingContract 也是它
	USDC    common.Address
	Markets []Market

	// 三个角色的私钥（hex，可带 0x），分别对应 anvil #1 / #3 / 清算人账户
	OrderSenderKey string // 调 Perpetual.Trade 的撮合账户（必须在 Dealer 的 validOrderSender 白名单里）
	KeeperKey      string // 调 Dealer.UpdateFundingRate 的 fundingRateKeeper
	LiquidatorKey  string // 调 Perpetual.Liquidate 的清算人（自己要有保证金接仓位）

	FundingInterval     time.Duration // 资金费 keeper 多久跑一次
	LiquidationInterval time.Duration // 清算机器人多久扫一次
}

// Default 返回本地 anvil 的默认配置（不含私钥）。
func Default() Config {
	return Config{
		RPCURL:   "http://127.0.0.1:8545",
		HTTPAddr: ":8080",
		Dealer:   common.HexToAddress("0x5FC8d32690cc91D4c39d9d3abcBD16989F875707"),
		USDC:     common.HexToAddress("0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9"),
		Markets: []Market{
			{Name: "BTC-USDC", Perp: common.HexToAddress("0xB7f8BC63BbcaD18155201308C8f3540b07f84F5e")},
		},
		FundingInterval:     time.Hour,
		LiquidationInterval: 10 * time.Second,
	}
}

// Load 在 Default 基础上用环境变量覆盖：ORDER_SENDER_PK / KEEPER_PK / LIQUIDATOR_PK 必填，
// 缺了一次性列出全部缺失项；RPC_URL / HTTP_ADDR / FUNDING_INTERVAL / LIQUIDATION_INTERVAL 可选，非空才覆盖默认值。
func Load() (Config, error) {
	cfg := Default()
	orderSenderPk := strings.TrimSpace(os.Getenv("ORDER_SENDER_PK"))
	keeperPk := strings.TrimSpace(os.Getenv("KEEPER_PK"))
	liquidatorPk := strings.TrimSpace(os.Getenv("LIQUIDATOR_PK"))
	var missing []string
	if orderSenderPk == "" {
		missing = append(missing, "ORDER_SENDER_PK")
	}
	if keeperPk == "" {
		missing = append(missing, "KEEPER_PK")
	}
	if liquidatorPk == "" {
		missing = append(missing, "LIQUIDATOR_PK")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	cfg.OrderSenderKey = orderSenderPk
	cfg.KeeperKey = keeperPk
	cfg.LiquidatorKey = liquidatorPk

	rpcUrl := os.Getenv("RPC_URL")
	if rpcUrl != "" {
		cfg.RPCURL = rpcUrl
	}
	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr != "" {
		cfg.HTTPAddr = httpAddr
	}

	// 两个间隔可选，格式同 time.ParseDuration（例 "15s"）；端到端演示时调短
	for _, d := range []struct {
		env string
		dst *time.Duration
	}{
		{"FUNDING_INTERVAL", &cfg.FundingInterval},
		{"LIQUIDATION_INTERVAL", &cfg.LiquidationInterval},
	} {
		raw := os.Getenv(d.env)
		if raw == "" {
			continue
		}
		v, err := time.ParseDuration(raw)
		if err != nil || v <= 0 {
			return Config{}, fmt.Errorf("invalid %s %q: must be a positive duration like 15s", d.env, raw)
		}
		*d.dst = v
	}

	return cfg, nil
}
