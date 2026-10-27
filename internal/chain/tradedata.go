package chain

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// EncodeTradeData 生成 Perpetual.trade 的参数，必须能被 approveTrade 里的
// abi.decode(tradeData, (Types.Order[], bytes[], uint256[])) 解开。
//
//	orders[0] 是 taker，orders[1:] 是 maker；sigs[i] 是 orders[i] 的 65 字节签名；
//	amounts[i] 是 orders[i] 本次成交的 paper 绝对值（1e18 精度），taker 的 = 所有 maker 之和。
//
// 三个切片长度必须相等且 >= 2。
func EncodeTradeData(orders []Order, sigs [][]byte, amounts []*big.Int) ([]byte, error) {
	if len(orders) < 2 || len(orders) != len(sigs) || len(orders) != len(amounts) {
		return nil, fmt.Errorf("need >= 2 orders and equal lengths, got orders=%d sigs=%d amounts=%d",
			len(orders), len(sigs), len(amounts))
	}

	for i, a := range amounts {
		if a == nil || a.Sign() <= 0 {
			return nil, fmt.Errorf("amounts[%d] must be > 0, got %v", i, a)
		}
	}

	orderListType, err := abi.NewType("tuple[]", "", []abi.ArgumentMarshaling{
		{Name: "perp", Type: "address"},
		{Name: "signer", Type: "address"},
		{Name: "paperAmount", Type: "int128"},
		{Name: "creditAmount", Type: "int128"},
		{Name: "info", Type: "bytes32"},
	})

	if err != nil {
		return nil, err
	}

	sigListType, err := abi.NewType("bytes[]", "", nil)
	if err != nil {
		return nil, err
	}

	paperListType, err := abi.NewType("uint256[]", "", nil)
	if err != nil {
		return nil, err
	}

	return abi.Arguments{
		{Type: orderListType},
		{Type: sigListType},
		{Type: paperListType},
	}.Pack(orders, sigs, amounts)
}
