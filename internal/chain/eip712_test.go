package chain

import (
	"bytes"
	"math"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// 期望值由 Solidity 独立算出（chisel eval，与 test/utils/EIP712Test.sol 的 _packInfo 同一公式）：
//
//	bytes32((uint256(uint64(m)) << 192) | (uint256(uint64(t)) << 128) | (uint256(e) << 64) | uint256(n))
//
// 解包一侧用合约 Trading._info2* 的公式在 chisel 里验证过：第二组解出 maker -2e14、taker -1e14、expiration 1792700000。
func TestPackInfoMatchesSolidity(t *testing.T) {
	cases := []struct {
		name string
		in   OrderInfo
		want string
	}{
		{
			name: "maker 负费率（返佣）, taker 正费率",
			in:   OrderInfo{MakerFeeRate: -2e14, TakerFeeRate: 5e14, Expiration: 1792700000, Nonce: 7},
			want: "0xffff4a19df0b80000001c6bf52634000000000006ada6e600000000000000007",
		},
		{
			// taker 为负时最容易出符号扩展污染：maker 那 8 字节会被刷成全 ff
			name: "两侧都是负费率",
			in:   OrderInfo{MakerFeeRate: -2e14, TakerFeeRate: -1e14, Expiration: 1792700000, Nonce: 7},
			want: "0xffff4a19df0b8000ffffa50cef85c000000000006ada6e600000000000000007",
		},
		{
			name: "边界：int64 最小/最大、uint64 最大、nonce 0",
			in:   OrderInfo{MakerFeeRate: math.MinInt64, TakerFeeRate: math.MaxInt64, Expiration: math.MaxUint64, Nonce: 0},
			want: "0x80000000000000007fffffffffffffffffffffffffffffff0000000000000000",
		},
		{
			name: "全 0",
			in:   OrderInfo{},
			want: "0x0000000000000000000000000000000000000000000000000000000000000000",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := common.Hash(PackInfo(tc.in)).Hex()
			if got != tc.want {
				t.Fatalf("PackInfo(%+v)\n got  %s\n want %s", tc.in, got, tc.want)
			}
			// 反向：从 Solidity 的结果解包，必须还原出原始字段（同时证明 UnpackInfo 与 PackInfo 互逆）
			if back := UnpackInfo(common.HexToHash(tc.want)); back != tc.in {
				t.Fatalf("UnpackInfo(%s) = %+v, want %+v", tc.want, back, tc.in)
			}
		})
	}
}

// D7 第 5 步上链用的那张单：alice 买 0.1 张 @26000，永不过期（2100 年），nonce 1，费率 0。
func aliceBuyOrder() Order {
	return Order{
		Perp:         common.HexToAddress("0xB7f8BC63BbcaD18155201308C8f3540b07f84F5e"),
		Signer:       common.HexToAddress("0x15d34AAf54267DB7D7c367839AAf71A00a2C6A65"),
		PaperAmount:  big.NewInt(1e17),
		CreditAmount: big.NewInt(-2600e6),
		Info:         PackInfo(OrderInfo{Expiration: 4102444800, Nonce: 1}),
	}
}

var anvilDealer = common.HexToAddress("0x5FC8d32690cc91D4c39d9d3abcBD16989F875707")

// typehash 期望值：cast keccak "<类型字符串>"
func TestTypeHashes(t *testing.T) {
	if got, want := orderTypeHash.Hex(), "0x6ce22f1be37c24a60b088bc91a456d1ae077db421fa72ffe85abd6af245899e4"; got != want {
		t.Errorf("orderTypeHash = %s, want %s", got, want)
	}
	if got, want := domainTypeHash.Hex(), "0x8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f"; got != want {
		t.Errorf("domainTypeHash = %s, want %s", got, want)
	}
}

// 期望值 = 全新 anvil 部署后 cast call <Dealer> "domainSeparator()(bytes32)" 读到的链上值
func TestDomainSeparatorMatchesChain(t *testing.T) {
	got := DomainSeparator(big.NewInt(31337), anvilDealer).Hex()
	if want := "0x8938cef411dd0437552cc5e02ef0158cf902db9960519ab721b8ea2d50e81689"; got != want {
		t.Fatalf("DomainSeparator = %s, want %s", got, want)
	}
	// 换链或换合约，domainSeparator 必须变（这就是防重放的意义）
	if DomainSeparator(big.NewInt(1), anvilDealer).Hex() == got {
		t.Error("chainId 变了 domainSeparator 却没变")
	}
	if DomainSeparator(big.NewInt(31337), common.HexToAddress("0xB7f8BC63BbcaD18155201308C8f3540b07f84F5e")).Hex() == got {
		t.Error("verifyingContract 变了 domainSeparator 却没变")
	}
}

// 期望值由 chisel eval 用 Solidity 的 keccak256(abi.encode(...)) / keccak256(abi.encodePacked("\x19\x01", ds, sh)) 算出
func TestStructHashAndDigestMatchSolidity(t *testing.T) {
	o := aliceBuyOrder()
	if got, want := StructHash(o).Hex(), "0xf4ae175889b7a0e6c27e3357894190448ee2d6486f19f60531900768eb2330f3"; got != want {
		t.Errorf("StructHash = %s, want %s", got, want)
	}
	ds := common.HexToHash("0x8938cef411dd0437552cc5e02ef0158cf902db9960519ab721b8ea2d50e81689")
	if got, want := Digest(ds, o).Hex(), "0xa4d34043f222e1c72256b2e25938b6fd9b33090359b7b5fa135436fecc44aa92"; got != want {
		t.Errorf("Digest = %s, want %s", got, want)
	}
}

// math.U256Bytes 会原地改参数：StructHash 不能把订单里的负数 credit 改成 2^256 - x
func TestStructHashDoesNotMutateOrder(t *testing.T) {
	o := aliceBuyOrder()
	StructHash(o)
	if o.PaperAmount.Cmp(big.NewInt(1e17)) != 0 || o.CreditAmount.Cmp(big.NewInt(-2600e6)) != 0 {
		t.Fatalf("StructHash 改了订单：paper %s credit %s", o.PaperAmount, o.CreditAmount)
	}
}

// 每个字段都要参与哈希：改任何一个，structHash 都必须变
func TestStructHashCoversEveryField(t *testing.T) {
	base := StructHash(aliceBuyOrder())
	mutations := map[string]func(*Order){
		"perp":   func(o *Order) { o.Perp = anvilDealer },
		"signer": func(o *Order) { o.Signer = anvilDealer },
		"paper":  func(o *Order) { o.PaperAmount = big.NewInt(2e17) },
		"credit": func(o *Order) { o.CreditAmount = big.NewInt(-2601e6) },
		"info":   func(o *Order) { o.Info = PackInfo(OrderInfo{Expiration: 4102444800, Nonce: 2}) },
	}
	for name, mutate := range mutations {
		o := aliceBuyOrder()
		mutate(&o)
		if StructHash(o) == base {
			t.Errorf("改了 %s，structHash 没变", name)
		}
	}
}

// anvil #4 alice 的私钥（anvil 公开的测试私钥）
const aliceKeyHex = "47e179ec197488593b187f80a00eb0da91f1b9d0b13f8733639f19c30a34926a"

var aliceAddr = common.HexToAddress("0x15d34AAf54267DB7D7c367839AAf71A00a2C6A65")

// 期望签名：cast wallet sign --no-hash --private-key <alice> <digest>
// secp256k1 用 RFC 6979 确定性 k，同一私钥签同一 digest 结果固定，可以当向量比对
func TestSignOrderMatchesCast(t *testing.T) {
	key, err := crypto.HexToECDSA(aliceKeyHex)
	if err != nil {
		t.Fatal(err)
	}
	digest := common.HexToHash("0xa4d34043f222e1c72256b2e25938b6fd9b33090359b7b5fa135436fecc44aa92")
	sig, err := SignOrder(key, digest)
	if err != nil {
		t.Fatal(err)
	}
	want := "0xfe00db9aeeeb66357e9b9ab1fce9b62367399c4d6de163d639fd88a3bb5757196289af3876dfbf21b09d2f6938ac398165725e6a60a16aa613c70d7d18cf11131c"
	if got := hexutil.Encode(sig); got != want {
		t.Fatalf("SignOrder\n got  %s\n want %s", got, want)
	}
}

// 签 → 恢复 → 必须是 alice；恢复过程不能改调用方的 sig（它之后还要原样上链）
func TestSignThenRecover(t *testing.T) {
	key, _ := crypto.HexToECDSA(aliceKeyHex)
	digest := Digest(DomainSeparator(big.NewInt(31337), anvilDealer), aliceBuyOrder())
	sig, err := SignOrder(key, digest)
	if err != nil {
		t.Fatal(err)
	}
	if v := sig[64]; v != 27 && v != 28 {
		t.Fatalf("v = %d, OpenZeppelin ECDSA 只认 27/28", v)
	}
	before := bytes.Clone(sig)

	got, err := RecoverSigner(digest, sig)
	if err != nil {
		t.Fatal(err)
	}
	if got != aliceAddr {
		t.Fatalf("RecoverSigner = %s, want alice %s", got.Hex(), aliceAddr.Hex())
	}
	if !bytes.Equal(sig, before) {
		t.Fatal("RecoverSigner 改了调用方传入的 sig")
	}

	// 订单被篡改（digest 变了）→ 恢复出的是另一个地址，验签不通过
	tampered := aliceBuyOrder()
	tampered.CreditAmount = big.NewInt(-2000e6)
	other, err := RecoverSigner(Digest(DomainSeparator(big.NewInt(31337), anvilDealer), tampered), sig)
	if err == nil && other == aliceAddr {
		t.Fatal("订单被改了，签名却还能恢复出 alice")
	}
}

func TestRecoverSignerRejectsBadInput(t *testing.T) {
	key, _ := crypto.HexToECDSA(aliceKeyHex)
	digest := common.HexToHash("0xa4d34043f222e1c72256b2e25938b6fd9b33090359b7b5fa135436fecc44aa92")
	sig, _ := SignOrder(key, digest)

	raw := bytes.Clone(sig)
	raw[64] -= 27 // go-ethereum 原生的 0/1，合约不认，链下也要拒
	cases := map[string][]byte{
		"v 是 0/1": raw,
		"v 是 29":  append(bytes.Clone(sig[:64]), 29),
		"64 字节":   sig[:64],
		"66 字节":   append(bytes.Clone(sig), 0),
		"空签名":     nil,
	}
	for name, s := range cases {
		if _, err := RecoverSigner(digest, s); err == nil {
			t.Errorf("%s：应该报错", name)
		}
	}
}
