-include .env

ANVIL_PORT ?= 8545
COMPOSE := docker compose --env-file .env

.PHONY: help docker-up docker-down docker-logs docker-ps chain backend-build backend-test backend-lint migrate-up migrate-down migrate-version migrate-force migrate-create sqlc-generate contracts-build contracts-test contracts-fmt

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
