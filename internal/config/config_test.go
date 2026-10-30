package config

import (
	"strings"
	"testing"
	"time"
)

func setKeys(t *testing.T, orderSender, keeper, liquidator string) {
	t.Helper()
	t.Setenv("ORDER_SENDER_PK", orderSender)
	t.Setenv("KEEPER_PK", keeper)
	t.Setenv("LIQUIDATOR_PK", liquidator)
	t.Setenv("RPC_URL", "")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("FUNDING_INTERVAL", "")
	t.Setenv("LIQUIDATION_INTERVAL", "")
}

func TestLoadMissingKeys(t *testing.T) {
	tests := []struct {
		name                          string
		orderSender, keeper, liquidat string
		want, notWant                 []string
	}{
		{"all missing", "", "", "", []string{"ORDER_SENDER_PK", "KEEPER_PK", "LIQUIDATOR_PK"}, nil},
		{"only keeper missing", "0xaa", "", "0xcc", []string{"KEEPER_PK"}, []string{"ORDER_SENDER_PK", "LIQUIDATOR_PK"}},
		{"whitespace is missing", "  \n", "0xbb", "0xcc", []string{"ORDER_SENDER_PK"}, []string{"KEEPER_PK", "LIQUIDATOR_PK"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setKeys(t, tt.orderSender, tt.keeper, tt.liquidat)
			_, err := Load()
			if err == nil {
				t.Fatal("want error, got nil")
			}
			for _, s := range tt.want {
				if !strings.Contains(err.Error(), s) {
					t.Errorf("error %q should mention %s", err, s)
				}
			}
			for _, s := range tt.notWant {
				if strings.Contains(err.Error(), s) {
					t.Errorf("error %q should not mention %s", err, s)
				}
			}
			// 报错里不能带私钥值
			for _, v := range []string{"0xaa", "0xbb", "0xcc"} {
				if strings.Contains(err.Error(), v) {
					t.Errorf("error %q leaks key value", err)
				}
			}
		})
	}
}

func TestLoadDefaultsAndOverrides(t *testing.T) {
	setKeys(t, " 0xaa\n", "0xbb", "0xcc")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	def := Default()
	if cfg.OrderSenderKey != "0xaa" || cfg.KeeperKey != "0xbb" || cfg.LiquidatorKey != "0xcc" {
		t.Errorf("keys = %q %q %q", cfg.OrderSenderKey, cfg.KeeperKey, cfg.LiquidatorKey)
	}
	// 空串不覆盖默认值，Default 里其他字段保留
	if cfg.RPCURL != def.RPCURL || cfg.HTTPAddr != def.HTTPAddr {
		t.Errorf("empty env overrode defaults: %q %q", cfg.RPCURL, cfg.HTTPAddr)
	}
	if cfg.Dealer != def.Dealer || cfg.USDC != def.USDC || len(cfg.Markets) != 1 ||
		cfg.FundingInterval != def.FundingInterval || cfg.LiquidationInterval != def.LiquidationInterval {
		t.Errorf("defaults lost: %+v", cfg)
	}

	t.Setenv("RPC_URL", "http://10.0.0.1:8545")
	t.Setenv("HTTP_ADDR", ":9000")
	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RPCURL != "http://10.0.0.1:8545" || cfg.HTTPAddr != ":9000" {
		t.Errorf("override failed: %q %q", cfg.RPCURL, cfg.HTTPAddr)
	}
}

func TestLoadIntervals(t *testing.T) {
	setKeys(t, "0xaa", "0xbb", "0xcc")
	t.Setenv("FUNDING_INTERVAL", "15s")
	t.Setenv("LIQUIDATION_INTERVAL", "2s")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.FundingInterval != 15*time.Second || cfg.LiquidationInterval != 2*time.Second {
		t.Errorf("intervals = %v %v", cfg.FundingInterval, cfg.LiquidationInterval)
	}
	for _, bad := range []string{"15", "abc", "0s", "-1s"} {
		t.Setenv("FUNDING_INTERVAL", bad)
		if _, err := Load(); err == nil || !strings.Contains(err.Error(), "FUNDING_INTERVAL") {
			t.Errorf("FUNDING_INTERVAL=%q: err = %v", bad, err)
		}
	}
}
