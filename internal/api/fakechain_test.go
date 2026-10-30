package api

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
	"github.com/xuefei016/my-perpetual-backend/internal/chaintest"
)

var testPriceSrc = common.HexToAddress("0x610178dA211FEF7D417bC0e6FeD39F05609AD788")

type fakeAccount = chaintest.Account

var errNodeDown = chaintest.ErrNodeDown

func newFakeBackend(t *testing.T) *chaintest.Backend {
	t.Helper()
	fb, err := chaintest.New(testDealer, testPerp, testPriceSrc)
	if err != nil {
		t.Fatal(err)
	}
	return fb
}

// newFakeChain ChainId 故意不用 anvil 的 31337（防止实现里写死链 ID）。
func newFakeChain(t *testing.T, fb *chaintest.Backend) *chain.Chain {
	t.Helper()
	c, err := fb.Chain(testChainID)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
