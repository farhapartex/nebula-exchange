-include .env

ANVIL_PORT ?= 8545
COMPOSE := docker compose --env-file .env

.PHONY: help docker-up docker-down docker-logs docker-ps chain backend-build backend-test backend-lint

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
