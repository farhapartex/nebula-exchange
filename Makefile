-include .env

ANVIL_PORT ?= 8545
BASE_SEPOLIA_RPC_URL ?= https://sepolia.base.org
LOCAL_RPC_URL := http://localhost:$(ANVIL_PORT)
LOCAL_NETWORK_CONFIG := ./config/networks/base-sepolia.json
COMPOSE := docker compose --env-file .env
MIGRATE := $(COMPOSE) run --rm migrate

.PHONY: help docker-up docker-down docker-logs docker-ps backend-build backend-test backend-lint migrate-up migrate-down migrate-version migrate-create seed dev-unlock-chapter stripe-listen chain contracts-build contracts-test contracts-fmt contracts-test-fork contracts-deploy-local contracts-upgrade-local contracts-export-abi

help:
	@echo "Available commands:"
	@echo "  make docker-up                Build images, run migrations and start the backend stack"
	@echo "  make docker-down              Stop the containers"
	@echo "  make docker-logs              Follow container logs"
	@echo "  make docker-ps                Show container status"
	@echo "  make backend-build            Build the server and migrate binaries into backend/bin"
	@echo "  make backend-test             Run backend tests with the race detector"
	@echo "  make backend-lint             Check formatting and run go vet"
	@echo "  make migrate-up               Apply all pending migrations"
	@echo "  make migrate-down             Roll back the last migration"
	@echo "  make migrate-version          Show the current migration version"
	@echo "  make migrate-create name=NAME Create a new up and down migration pair"
	@echo "  make seed                     Load fighters and story levels into Postgres and their images into MinIO"
	@echo "  make dev-unlock-chapter email=E chapter=C  Give a player a chapter for free (development only)"
	@echo "  make stripe-listen            Forward Stripe test webhooks to the local backend"
	@echo "  make chain                    Start a local Anvil chain on the host, forked from Base Sepolia"
	@echo "  make contracts-build          Compile the smart contracts"
	@echo "  make contracts-test           Run Foundry tests"
	@echo "  make contracts-fmt            Format Solidity files"
	@echo "  make contracts-test-fork      Run contract tests against the real Base Sepolia Chainlink feed and USDC"
	@echo "  make contracts-deploy-local   Deploy the payment vault behind a proxy to the local chain and fund test USDC"
	@echo "  make contracts-upgrade-local  Deploy a new vault implementation and upgrade the local proxy to it"
	@echo "  make contracts-export-abi     Copy the vault ABI from the Foundry build into the frontend"

docker-up:
	$(COMPOSE) up -d --build

docker-down:
	$(COMPOSE) down

docker-logs:
	$(COMPOSE) logs -f

docker-ps:
	$(COMPOSE) ps

backend-build:
	cd backend && go build -o bin/server ./cmd/server && go build -o bin/migrate ./cmd/migrate

backend-test:
	cd backend && go test -race ./...

backend-lint:
	cd backend && test -z "$$(gofmt -l .)" && go vet ./...

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down

migrate-version:
	$(MIGRATE) version

migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=create_users" && exit 1)
	@next_version=$$(printf "%06d" $$(( $$(ls backend/migrations/*.up.sql 2>/dev/null | wc -l) + 1 ))); \
	touch backend/migrations/$${next_version}_$(name).up.sql backend/migrations/$${next_version}_$(name).down.sql; \
	echo "created backend/migrations/$${next_version}_$(name).up.sql and .down.sql"

seed:
	$(COMPOSE) run --rm --build seed

dev-unlock-chapter:
	@test "$(APP_ENV)" = "development" || (echo "dev-unlock-chapter only runs with APP_ENV=development" && exit 1)
	@test -n "$(email)" -a -n "$(chapter)" || (echo "usage: make dev-unlock-chapter email=player@streetborn.test chapter=1" && exit 1)
	$(COMPOSE) exec -T postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 -c "INSERT INTO chapter_unlocks (user_id, chapter_id, source, unlocked_at) SELECT id, '$(chapter)', 'GRANT', now() FROM users WHERE lower(email) = lower('$(email)') ON CONFLICT DO NOTHING RETURNING user_id, chapter_id, source;"

stripe-listen:
	@test -n "$(STRIPE_SECRET_KEY)" || (echo "STRIPE_SECRET_KEY must be set in .env" && exit 1)
	@stripe listen --api-key "$(STRIPE_SECRET_KEY)" --forward-to localhost:$(or $(BACKEND_PORT),8080)/api/v1/webhooks/stripe --events checkout.session.completed,checkout.session.async_payment_succeeded,checkout.session.expired,charge.refunded,charge.dispute.created

chain:
	@anvil --host 0.0.0.0 --port $(ANVIL_PORT) --chain-id 31337 --fork-url "$(BASE_SEPOLIA_RPC_URL)"

contracts-build:
	cd smart-contract && forge build

contracts-test:
	cd smart-contract && forge test -vv

contracts-fmt:
	cd smart-contract && forge fmt

contracts-test-fork:
	@cd smart-contract && BASE_SEPOLIA_RPC_URL="$(BASE_SEPOLIA_RPC_URL)" forge test --match-path test/ChapterPaymentVaultFork.t.sol

contracts-deploy-local:
	@test -n "$(LOCAL_DEPLOYER_PRIVATE_KEY)" -a -n "$(PAYMENT_SIGNER_PRIVATE_KEY)" || (echo "LOCAL_DEPLOYER_PRIVATE_KEY and PAYMENT_SIGNER_PRIVATE_KEY must be set in .env" && exit 1)
	@cd smart-contract && DEPLOYER_PRIVATE_KEY="$(LOCAL_DEPLOYER_PRIVATE_KEY)" PAYMENT_SIGNER_ADDRESS="$$(cast wallet address --private-key "$(PAYMENT_SIGNER_PRIVATE_KEY)")" NETWORK_CONFIG="$(LOCAL_NETWORK_CONFIG)" DEPLOYMENT_NAME=local forge script script/DeployChapterPaymentVault.s.sol --rpc-url $(LOCAL_RPC_URL) --broadcast
	@cd smart-contract && ./script/fund-local-usdc.sh $(LOCAL_RPC_URL) "$$(jq -r .payment_token $(LOCAL_NETWORK_CONFIG))"
	@./smart-contract/script/write-contract-addresses-to-env.sh smart-contract/deployments/local.json frontend/.env.local .env
	@cat smart-contract/deployments/local.json

contracts-upgrade-local:
	@test -n "$(LOCAL_DEPLOYER_PRIVATE_KEY)" || (echo "LOCAL_DEPLOYER_PRIVATE_KEY must be set in .env" && exit 1)
	@cd smart-contract && DEPLOYER_PRIVATE_KEY="$(LOCAL_DEPLOYER_PRIVATE_KEY)" DEPLOYMENT_NAME=local forge script script/UpgradeChapterPaymentVault.s.sol --rpc-url $(LOCAL_RPC_URL) --broadcast
	@cat smart-contract/deployments/local.json

contracts-export-abi:
	@cd smart-contract && forge build
	@mkdir -p frontend/src/lib/web3/abi
	@printf 'export const chapterPaymentVaultAbi = %s as const;\n' "$$(jq -c .abi smart-contract/out/ChapterPaymentVault.sol/ChapterPaymentVault.json)" > frontend/src/lib/web3/abi/chapter-payment-vault-abi.ts
	@cd frontend && npx prettier --write src/lib/web3/abi >/dev/null
	@echo "wrote frontend/src/lib/web3/abi/chapter-payment-vault-abi.ts"
