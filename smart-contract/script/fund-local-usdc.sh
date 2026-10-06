#!/usr/bin/env bash
set -euo pipefail

rpc_url="${1:?usage: fund-local-usdc.sh <rpc-url> <usdc-address>}"
usdc_address="${2:?usage: fund-local-usdc.sh <rpc-url> <usdc-address>}"
usdc_units_per_account=1000000000
gas_money_wei=0x8AC7230489E80000
anvil_accounts=(
  0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
  0x70997970C51812dc3A010C7d01b50e0d17dc79C8
  0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC
  0x90F79bf6EB2c4f870365E785982E1f101E93b906
  0x15d34AAf54267DB7D7c367839AAf71A00a2C6A65
)

master_minter=$(cast call "$usdc_address" "masterMinter()(address)" --rpc-url "$rpc_url")
cast rpc anvil_impersonateAccount "$master_minter" --rpc-url "$rpc_url" >/dev/null
cast rpc anvil_setBalance "$master_minter" "$gas_money_wei" --rpc-url "$rpc_url" >/dev/null
total_units=$((usdc_units_per_account * ${#anvil_accounts[@]}))
cast send "$usdc_address" "configureMinter(address,uint256)" "$master_minter" "$total_units" \
  --from "$master_minter" --unlocked --rpc-url "$rpc_url" >/dev/null

for account in "${anvil_accounts[@]}"; do
  cast send "$usdc_address" "mint(address,uint256)" "$account" "$usdc_units_per_account" \
    --from "$master_minter" --unlocked --rpc-url "$rpc_url" >/dev/null
  echo "$account $(cast call "$usdc_address" "balanceOf(address)(uint256)" "$account" --rpc-url "$rpc_url") USDC units"
done

cast rpc anvil_stopImpersonatingAccount "$master_minter" --rpc-url "$rpc_url" >/dev/null
