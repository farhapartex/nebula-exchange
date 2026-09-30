-include .env

ANVIL_PORT ?= 8545
COMPOSE := docker compose --env-file .env

.PHONY: help docker-up docker-down docker-logs docker-ps chain backend-build backend-test backend-lint migrate-up migrate-down migrate-version migrate-force migrate-create sqlc-generate contracts-build contracts-test contracts-fmt dev-activate-user dev-reset-rate-limits dev-set-status dev-credit-nc dev-grant-item dev-complete-payment dev-finish-missions

help:
	@echo "Available commands:"
	@echo "  make docker-up     Build and start the backend with Postgres, Redis and Mailpit"
	@echo "  make docker-down   Stop the containers"
	@echo "  make docker-logs   Follow container logs"
	@echo "  make docker-ps     Show container status"
	@echo "  make chain         Start a local Anvil chain on the host"
	@echo "  make backend-build Build the backend binary into backend/bin"
	@echo "  make backend-test  Run backend tests with the race detector"
	@echo "  make backend-lint  Check formatting and run go vet"
	@echo "  make migrate-up                 Apply all pending migrations"
	@echo "  make migrate-down               Roll back the last migration"
	@echo "  make migrate-version            Show the current migration version"
	@echo "  make migrate-force version=N    Mark the database as being at version N"
	@echo "  make migrate-create name=NAME   Create a new up/down migration pair"
	@echo "  make sqlc-generate              Generate typed Go code from SQL queries"
	@echo "  make contracts-build            Compile the smart contracts"
	@echo "  make contracts-test             Run Foundry tests"
	@echo "  make contracts-fmt              Format Solidity files"
	@echo "  make dev-activate-user email=E            Activate an account directly (development only)"
	@echo "  make dev-reset-rate-limits                 Clear rate limits and login lockouts"
	@echo "  make dev-set-status email=E status=S       Set an account status"
	@echo "  make dev-credit-nc email=E amount=A [bucket=B]  Credit test NC through the ledger"
	@echo "  make dev-grant-item email=E item=I [quantity=Q] Grant test items through the ledger"
	@echo "  make dev-complete-payment id=PAYMENT_ID     Simulate a successful card checkout"
	@echo "  make dev-finish-missions email=E           End a player's running missions now"

docker-up:
	$(COMPOSE) up -d --build

docker-down:
	$(COMPOSE) down

docker-logs:
	$(COMPOSE) logs -f

docker-ps:
	$(COMPOSE) ps

chain:
	anvil --host 0.0.0.0 --port $(ANVIL_PORT) --chain-id 31337

backend-build:
	cd backend && go build -o bin/server ./cmd/server

backend-test:
	cd backend && go test -race ./...

backend-lint:
	cd backend && test -z "$$(gofmt -l .)" && go vet ./...

MIGRATE := $(COMPOSE) run --rm migrate

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

migrate-version:
	$(MIGRATE) version

migrate-force:
	@test -n "$(version)" || (echo "usage: make migrate-force version=N" && exit 1)
	$(MIGRATE) force $(version)

migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=create_users" && exit 1)
	$(COMPOSE) run --rm --no-deps migrate create -ext sql -dir /migrations -seq $(name)

sqlc-generate:
	cd backend && go tool sqlc generate

contracts-build:
	cd smart-contract && forge build

contracts-test:
	cd smart-contract && forge test -vv

contracts-fmt:
	cd smart-contract && forge fmt

dev-activate-user:
	@test -n "$(email)" || (echo "usage: make dev-activate-user email=pilot@nebula.test" && exit 1)
	$(COMPOSE) exec -T postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 -c "UPDATE users SET is_active = true, activated_at = now(), status = 'PENDING_PAYMENT', updated_at = now() WHERE lower(email) = lower('$(email)') AND is_active = false RETURNING email, username, status;"

dev-reset-rate-limits:
	$(COMPOSE) exec -T redis sh -c "redis-cli --scan --pattern 'nebula:ratelimit:*' | xargs -r redis-cli del; redis-cli --scan --pattern 'nebula:login-lockout:*' | xargs -r redis-cli del"

dev-set-status:
	@test -n "$(email)" -a -n "$(status)" || (echo "usage: make dev-set-status email=pilot@nebula.test status=ACTIVE" && exit 1)
	$(COMPOSE) exec -T postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 -c "UPDATE users SET status = '$(status)', updated_at = now() WHERE lower(email) = lower('$(email)') RETURNING email, username, status, is_active;"

DEVTOOL := cd backend && DATABASE_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable go run ./cmd/devtool

dev-credit-nc:
	@test -n "$(email)" -a -n "$(amount)" || (echo "usage: make dev-credit-nc email=pilot@nebula.test amount=25 [bucket=card|crypto|earned_pending|earned]" && exit 1)
	$(DEVTOOL) credit-nc --email '$(email)' --amount '$(amount)' --bucket '$(or $(bucket),card)'

dev-grant-item:
	@test -n "$(email)" -a -n "$(item)" || (echo "usage: make dev-grant-item email=pilot@nebula.test item=401 [quantity=10]" && exit 1)
	$(DEVTOOL) grant-item --email '$(email)' --item '$(item)' --quantity '$(or $(quantity),1)'

dev-complete-payment:
	@test -n "$(id)" || (echo "usage: make dev-complete-payment id=PAYMENT_ID" && exit 1)
	$(DEVTOOL) complete-payment --id '$(id)'

dev-finish-missions:
	@test -n "$(email)" || (echo "usage: make dev-finish-missions email=pilot@nebula.test" && exit 1)
	$(COMPOSE) exec -T postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 -c "UPDATE missions SET ends_at = now(), started_at = LEAST(started_at, now() - interval '1 second') WHERE status = 'RUNNING' AND user_id = (SELECT id FROM users WHERE lower(email) = lower('$(email)')) RETURNING id, zone_id;"
