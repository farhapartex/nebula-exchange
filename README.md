# Nebula Exchange

A browser-based space trading game. Players mine resources on timed missions, craft and upgrade gear, and trade on a live order-book exchange and an auction house. Every item is backed by an ERC-1155 token held in an on-chain vault.

## Repository layout

| Folder | Contents | Runs in |
| --- | --- | --- |
| `frontend/` | Next.js, TypeScript, Tailwind CSS, wagmi + viem | Host |
| `backend/` | Go, Gin, PostgreSQL, Redis | Docker |
| `smart-contract/` | Solidity, Foundry, OpenZeppelin | Host (Anvil) |

## Prerequisites

- Docker with Compose v2
- Go 1.26+
- Node.js 22+
- Foundry (forge, anvil)

## Getting started

```bash
cp .env.example .env
make docker-up
make chain
```

| Service | Address |
| --- | --- |
| PostgreSQL | `localhost:5432` |
| Redis | `localhost:6379` |
| Mailpit SMTP | `localhost:1025` |
| Mailpit inbox | http://localhost:8025 |
| Anvil RPC | http://localhost:8545 |

Run `make help` to see all commands.

## Documents

- `Nebula Exchange — Product Requirements Document.md` is the product spec.
- `task_list.md` is the build plan.
- `CLAUDE.md` lists the coding and workflow rules.
