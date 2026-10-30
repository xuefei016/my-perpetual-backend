package chain

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/xuefei016/my-perpetual-backend/internal/chain/bindings/dealer"
	"github.com/xuefei016/my-perpetual-backend/internal/chain/bindings/erc20"
	"github.com/xuefei016/my-perpetual-backend/internal/chain/bindings/perpetual"
	"github.com/xuefei016/my-perpetual-backend/internal/chain/bindings/pricesource"
)

type Chain struct {
	Client  *ethclient.Client
	Backend bind.ContractBackend // 绑定任意合约用（= Client；单测里换成假后端）
	ChainId *big.Int
	Dealer  *dealer.Dealer
	BTC     *perpetual.Perpetual
	USDC    *erc20.ERC20
}

func New(ctx context.Context, rpcURL string, dealerAddr, btcAddr, usdcAddr common.Address) (*Chain, error) {
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", rpcURL, err)
	}
	chainId, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("chain id: %w", err)
	}
	deal, err := dealer.NewDealer(dealerAddr, client)
	if err != nil {
		return nil, fmt.Errorf("bind dealer: %w", err)
	}

	btc, err := perpetual.NewPerpetual(btcAddr, client)
	if err != nil {
		return nil, fmt.Errorf("bind btc: %w", err)
	}

	usdc, err := erc20.NewERC20(usdcAddr, client)
	if err != nil {
		return nil, fmt.Errorf("bind usdc: %w", err)
	}
	return &Chain{
		Client:  client,
		Backend: client,
		ChainId: chainId,
		Dealer:  deal,
		BTC:     btc,
		USDC:    usdc,
	}, nil
}

func (c *Chain) Transactor(keyHex string) (*bind.TransactOpts, error) {
	if keyHex == "" {
		return nil, fmt.Errorf("keyHex is empty")
	}
	privateKey := strings.TrimPrefix(keyHex, "0x")
	key, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, fmt.Errorf("private key: %w", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key, c.ChainId)
	if err != nil {
		return nil, fmt.Errorf("transactor: %w", err)
	}
	return auth, nil
}

func WaitSuccess(ctx context.Context, c *Chain, tx *types.Transaction) (*types.Receipt, error) {
	receipt, err := bind.WaitMined(ctx, c.Client, tx)
	if err != nil {
		return nil, fmt.Errorf("wait tx: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("wait tx %s receipt status: %d", receipt.TxHash.Hex(), receipt.Status)
	}
	return receipt, nil
}

// MarkPrice 读 perp 市场的标记价（USDC 6 位 / 1 张）：先从 Dealer 风控参数拿价格源地址，再调它的 getMarkPrice。
func (c *Chain) MarkPrice(opts *bind.CallOpts, perp common.Address) (*big.Int, error) {
	rp, err := c.Dealer.GetRiskParams(opts, perp)
	if err != nil {
		return nil, fmt.Errorf("get risk params: %w", err)
	}
	ps, err := pricesource.NewPriceSource(rp.MarkPriceSource, c.Backend)
	if err != nil {
		return nil, fmt.Errorf("bind price source: %w", err)
	}
	mark, err := ps.GetMarkPrice(opts)
	if err != nil {
		return nil, fmt.Errorf("get mark price: %w", err)
	}
	return mark, nil
}
