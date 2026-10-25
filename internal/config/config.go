// Package config 汇总后端运行需要的全部配置：RPC、合约地址、各角色私钥、轮询间隔。
// 地址默认值是全新 anvil 上跑 Deploy.s.sol 得到的确定性地址；私钥只从环境变量读，不写进代码。
package config

import (
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

// Load 在 Default 基础上用环境变量覆盖，并校验必填项。
//
// TODO(D9): 读 ORDER_SENDER_PK / KEEPER_PK / LIQUIDATOR_PK（必填，缺了返回错误），
// 以及可选的 RPC_URL / HTTP_ADDR。现在直接返回 Default()，私钥全空。
func Load() (Config, error) {
	return Default(), nil
}
