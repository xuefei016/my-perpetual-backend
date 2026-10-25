package chain

import "math/big"

// EncodeTradeData 生成 Perpetual.trade 的参数，必须能被 approveTrade 里的
// abi.decode(tradeData, (Types.Order[], bytes[], uint256[])) 解开。
//
//	orders[0] 是 taker，orders[1:] 是 maker；sigs[i] 是 orders[i] 的 65 字节签名；
//	amounts[i] 是 orders[i] 本次成交的 paper 绝对值（1e18 精度），taker 的 = 所有 maker 之和。
//
// 三个切片长度必须相等且 >= 2。
func EncodeTradeData(orders []Order, sigs [][]byte, amounts []*big.Int) ([]byte, error) {
	panic("TODO(D8): EncodeTradeData")
}
