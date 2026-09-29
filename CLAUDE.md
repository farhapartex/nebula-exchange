# Nebula Exchange — Working Rules

These rules apply to every part of the project: frontend, backend and smart contracts. They are strict.

## Repository

- One git repository with top-level folders: `frontend/`, `backend/`, `smart-contract/`.
- `task_list.md` is the build plan. The PRD is `Nebula Exchange — Product Requirements Document.md`.

## Code style

- No comments in code. Names must explain the code.
- Variable, function, type and file names must be meaningful and descriptive.
- Code is modular and split by scope. Each file has one clear responsibility. Never put unrelated code in one file.
- Reusable functions go in `utils` files. Reusable UI goes in shared component folders.

## Backend API

- Resource URLs are plural (`/users`, `/orders`, `/auctions`, `/wallets`, `/shop-items`). Fixed exceptions: `/me`, `/auth/*`, `/ws`, `/metadata/{id}.json`.
- Success responses are wrapped as `{"data": ...}`.
- Error responses are `{"error": {"code": "...", "message": "...", "details": {...}}}` with a 4xx/5xx status.
- Every list endpoint uses cursor pagination: `?cursor=&limit=` returns `{"data": [...], "pagination": {"next_cursor": "...", "limit": 20}}`.

## Decisions

- Never decide system design or business logic alone. Ask first, and continue only after confirmation. This includes filling gaps or ambiguities in the PRD.

## UI workflow

- For any task with UI work, build the UI first with mock data, then stop and wait for review.
- Backend or Solidity work for that task starts only after the user says "go".

## Git

- Commit by scope (for example frontend, backend, smart-contract, infra), not by task. One task can produce several commits.
- Commits use the repository owner's git identity. No co-author lines, no tool attribution of any kind.
- Commit messages are short and human, with a scope prefix: `backend: add signup endpoint`. One line, two at most.

## Docker

- Docker files and compose config may be written, but containers are never started (`docker compose up`, `docker run`, etc.) until the user says so.
