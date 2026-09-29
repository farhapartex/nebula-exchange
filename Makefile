-include .env

ANVIL_PORT ?= 8545
COMPOSE := docker compose --env-file .env

.PHONY: help docker-up docker-down docker-logs docker-ps chain

help:
	@echo "Available commands:"
	@echo "  make docker-up     Start Postgres, Redis and Mailpit"
	@echo "  make docker-down   Stop the containers"
	@echo "  make docker-logs   Follow container logs"
	@echo "  make docker-ps     Show container status"
	@echo "  make chain         Start a local Anvil chain on the host"

docker-up:
	$(COMPOSE) up -d

docker-down:
	$(COMPOSE) down

docker-logs:
	$(COMPOSE) logs -f

docker-ps:
	$(COMPOSE) ps

chain:
	anvil --host 0.0.0.0 --port $(ANVIL_PORT) --chain-id 31337
