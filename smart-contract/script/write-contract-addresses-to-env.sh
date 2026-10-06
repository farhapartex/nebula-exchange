#!/usr/bin/env bash
set -euo pipefail

deployment_file="${1:?usage: write-contract-addresses-to-env.sh <deployment-json> <frontend-env-file> <backend-env-file>}"
frontend_env_file="${2:?usage: write-contract-addresses-to-env.sh <deployment-json> <frontend-env-file> <backend-env-file>}"
backend_env_file="${3:?usage: write-contract-addresses-to-env.sh <deployment-json> <frontend-env-file> <backend-env-file>}"

vault_address=$(jq -r .chapter_payment_vault "$deployment_file")
payment_token_address=$(jq -r .payment_token "$deployment_file")

upsert_env_value() {
  local env_file="$1"
  local variable_name="$2"
  local variable_value="$3"
  touch "$env_file"
  if grep -q "^${variable_name}=" "$env_file"; then
    sed -i.bak "s|^${variable_name}=.*|${variable_name}=${variable_value}|" "$env_file"
    rm -f "${env_file}.bak"
  else
    printf '%s=%s\n' "$variable_name" "$variable_value" >> "$env_file"
  fi
}

upsert_env_value "$frontend_env_file" NEXT_PUBLIC_CHAPTER_PAYMENT_VAULT_ADDRESS "$vault_address"
upsert_env_value "$frontend_env_file" NEXT_PUBLIC_PAYMENT_TOKEN_ADDRESS "$payment_token_address"
upsert_env_value "$backend_env_file" CHAPTER_PAYMENT_VAULT_ADDRESS "$vault_address"
echo "contract addresses written to $frontend_env_file and $backend_env_file"
