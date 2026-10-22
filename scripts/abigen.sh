#!/usr/bin/env bash
# 从 ../my-perpetual/out 提取 ABI，生成 Go 绑定
# 用法：在任何目录执行 ./scripts/abigen.sh（合约改动后先在 my-perpetual 里 forge build）
set -euo pipefail

# 切到后端仓库根目录：不管从哪执行，下面的相对路径都成立
cd "$(dirname "$0")/.."

CONTRACTS=../my-perpetual
OUT=$CONTRACTS/out

# 先编译一次，保证 out/ 是最新的
(cd "$CONTRACTS" && forge build)

mkdir -p abi

# ---- Dealer ----
jq '.abi' "$OUT/MetaNodeDealer.sol/MetaNodeDealer.json" > abi/Dealer.abi
mkdir -p internal/chain/bindings/dealer
abigen --abi abi/Dealer.abi --pkg dealer --type Dealer \
       --out internal/chain/bindings/dealer/dealer.go

# ---- Perpetual ----
jq '.abi' "$OUT/Perpetual.sol/Perpetual.json" > abi/Perpetual.abi
mkdir -p internal/chain/bindings/perpetual
abigen --abi abi/Perpetual.abi --pkg perpetual --type Perpetual \
       --out internal/chain/bindings/perpetual/Perpetual.go

# ---- ERC20（用 MockERC20 的 ABI）----
jq '.abi' "$OUT/MockERC20.sol/MockERC20.json" > abi/ERC20.abi
mkdir -p internal/chain/bindings/erc20
abigen --abi abi/ERC20.abi --pkg erc20 --type ERC20 \
       --out internal/chain/bindings/erc20/erc20.go

echo "abigen done"