package chain

// 链下订单签名，和合约逐字节对齐：
//   - Types.Order / ORDER_TYPEHASH       src/libraries/Types.sol
//   - _structHash、info 解包             src/libraries/Trading.sol
//   - domainSeparator                    src/MetaNodeStorage.sol + src/libraries/EIP712.sol
//   - 验签（signer / operator / ERC1271） src/MetaNodeExternal.sol approveTrade
// 独立重算的对照实现在 test/utils/EIP712Test.sol。

import (
	"crypto/ecdsa"
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
)

// TODO(D7): 两个 typehash 常量（字符串必须和合约里一字不差，多一个空格哈希就全变）
//
//	orderTypeHash  = keccak256("Order(address perp,address signer,int128 paperAmount,int128 creditAmount,bytes32 info)")
//	domainTypeHash = keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)")
var orderTypeHash = crypto.Keccak256Hash([]byte("Order(address perp,address signer,int128 paperAmount,int128 creditAmount,bytes32 info)"))
var domainTypeHash = crypto.Keccak256Hash([]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"))

// Order 与 Types.Order 一一对应，字段顺序也一致（StructHash 按这个顺序编码）。
type Order struct {
	Perp         common.Address // 订单属于哪个市场；合约要求 == 调 approveTrade 的 Perpetual
	Signer       common.Address // 仓位归属人；签名人可以是它本人或它授权的 operator
	PaperAmount  *big.Int       // int128，1e18 = 1 张；> 0 买（开多），< 0 卖（开空）
	CreditAmount *big.Int       // int128，USDC 6 位小数；必须与 PaperAmount 异号
	Info         [32]byte       // PackInfo 打包：makerFee | takerFee | expiration | nonce
}

// OrderInfo 是 Order.Info 拆开后的四个字段（高位 → 低位，各 64 位）。
type OrderInfo struct {
	MakerFeeRate int64  // 1e18 = 100%，可为负（负费率 = 返佣给 maker）
	TakerFeeRate int64  // 同上
	Expiration   uint64 // unix 秒；合约要求 >= block.timestamp
	Nonce        uint64 // 只为让相同内容的两张单哈希不同
}

// PackInfo 把四个字段打包成 bytes32，与 Trading._info2* 的解包互逆。
// 负费率要先转成 uint64 取补码位再移位，否则符号扩展会把高位污染成全 1。
func PackInfo(i OrderInfo) [32]byte {
	var out [32]byte
	binary.BigEndian.PutUint64(out[0:8], uint64(i.MakerFeeRate))
	binary.BigEndian.PutUint64(out[8:16], uint64(i.TakerFeeRate))
	binary.BigEndian.PutUint64(out[16:24], i.Expiration)
	binary.BigEndian.PutUint64(out[24:32], i.Nonce)
	return out
}

// UnpackInfo 是 PackInfo 的逆运算，对应合约的 _info2MakerFeeRate / _info2TakerFeeRate / _info2Expiration。
func UnpackInfo(info [32]byte) OrderInfo {
	makerFeeRate := binary.BigEndian.Uint64(info[0:8])
	takerFeeRate := binary.BigEndian.Uint64(info[8:16])
	expiration := binary.BigEndian.Uint64(info[16:24])
	nonce := binary.BigEndian.Uint64(info[24:32])
	return OrderInfo{
		MakerFeeRate: int64(makerFeeRate),
		TakerFeeRate: int64(takerFeeRate),
		Expiration:   expiration,
		Nonce:        nonce,
	}
}

// DomainSeparator 重算 Dealer 构造时存下的 domainSeparator：name "MetaNode"、version "1"、chainId、verifyingContract = Dealer。
// 验收：结果必须等于链上 cast call <Dealer> "domainSeparator()(bytes32)"。
func DomainSeparator(chainID *big.Int, dealer common.Address) common.Hash {
	return crypto.Keccak256Hash(domainTypeHash.Bytes(), crypto.Keccak256([]byte("MetaNode")), crypto.Keccak256([]byte("1")), common.LeftPadBytes(chainID.Bytes(), 32), common.LeftPadBytes(dealer.Bytes(), 32))
}

// StructHash = keccak256(abi.encode(orderTypeHash, perp, signer, paperAmount, creditAmount, info))。
// abi.encode 每个字段占 32 字节：地址左补 0；int128 按 int256 符号扩展（负数左边补 0xff）。
func StructHash(o Order) common.Hash {
	return crypto.Keccak256Hash(orderTypeHash.Bytes(), common.LeftPadBytes(o.Perp.Bytes(), 32), common.LeftPadBytes(o.Signer.Bytes(), 32),
		math.U256Bytes(new(big.Int).Set(o.PaperAmount)), math.U256Bytes(new(big.Int).Set(o.CreditAmount)),
		o.Info[:],
	)
}

// Digest = keccak256(0x19 0x01 ‖ domainSeparator ‖ StructHash(o))，就是钱包实际签的 32 字节，
// 也是链上 orderFilledPaperAmount 的 key（引擎用它当订单 ID）。
func Digest(domainSeparator common.Hash, o Order) common.Hash {
	return crypto.Keccak256Hash([]byte{0x19, 0x01}, domainSeparator.Bytes(), StructHash(o).Bytes())
}

// SignOrder 用私钥签 digest，返回 65 字节 r ‖ s ‖ v。
// crypto.Sign 给出的 v 是 0/1，OpenZeppelin ECDSA 只认 27/28，要加 27。
func SignOrder(key *ecdsa.PrivateKey, digest common.Hash) ([]byte, error) {
	sig, err := crypto.Sign(digest[:], key)
	if err != nil {
		return nil, err
	}
	sig[64] += 27
	return sig, nil
}

// RecoverSigner 从签名恢复签名人地址（v 为 27/28），API 收单时用它做和合约一样的验签。
func RecoverSigner(digest common.Hash, sig []byte) (common.Address, error) {
	if len(sig) != 65 {
		return common.Address{}, errors.New("invalid signature")
	}
	if sig[64] != 27 && sig[64] != 28 {
		return common.Address{}, errors.New("invalid signature")
	}
	cp := make([]byte, 65)
	copy(cp, sig)
	cp[64] -= 27
	pubkey, err := crypto.SigToPub(digest[:], cp)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*pubkey), nil
}
