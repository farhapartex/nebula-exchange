-include .env

ANVIL_PORT ?= 8545
COMPOSE := docker compose --env-file .env
MIGRATE := $(COMPOSE) run --rm migrate

.PHONY: help docker-up docker-down docker-logs docker-ps backend-build backend-test backend-lint migrate-up migrate-down migrate-version migrate-create seed chain contracts-build contracts-test contracts-fmt

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
	@echo "  make chain                    Start a local Anvil chain on the host"
	@echo "  make contracts-build          Compile the smart contracts"
	@echo "  make contracts-test           Run Foundry tests"
	@echo "  make contracts-fmt            Format Solidity files"

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

chain:
	anvil --host 0.0.0.0 --port $(ANVIL_PORT) --chain-id 31337

contracts-build:
	cd smart-contract && forge build

contracts-test:
	cd smart-contract && forge test -vv

contracts-fmt:
	cd smart-contract && forge fmt
