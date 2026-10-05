# Street Born

Working title. A story-driven street fighting game for the browser.

A thin, hungry boy lives alone in a broken village house. One night bandits burn it down and come for him with a hammer, a knife and finally a gun. He survives, and a fighter from a street club offers him a place. The player is that boy. Each chapter is a run of fights told through short story slides, and every fight is played in a painted 2.5D arena.

Chapter 1 is free. Later chapters are unlocked with coins. Players buy tools (an iron pipe, a scrap shield and so on), the tools grow stronger the more they are used, and a strong tool can be sold to other players for coins.

## Repository layout

| Folder | Contents | Runs on |
| --- | --- | --- |
| `frontend/` | Next.js, React, TypeScript, Tailwind CSS, Three.js for the fight stage | Host |
| `backend/` | Go, Gin, GORM, PostgreSQL, Redis | Docker |
| `smart-contract/` | Solidity, Foundry, OpenZeppelin (planned for tools as NFTs later) | Host (Anvil) |

## Getting started

You need Docker with Compose v2, Go 1.26 or newer, Node.js 22 or newer, and Foundry if you work on the contracts.

```bash
cp .env.example .env
make docker-up
cd frontend && npm install && npm run dev
```

`make docker-up` builds the backend image, starts Postgres, Redis and Mailpit, runs the `migrate` container once to apply every pending migration, and then starts the API.

| Service | Address |
| --- | --- |
| API | http://localhost:8080/api/v1 |
| Health check | http://localhost:8080/api/v1/health |
| PostgreSQL | localhost:5432 |
| Redis | localhost:6379 |
| Mailpit inbox | http://localhost:8025 |
| Frontend | http://localhost:3000 |

Run `make help` to see every command.

## Backend

The backend is a single Go service split by domain (a modular monolith). Each domain is a package under `backend/internal/` and owns its tables, its rules and its routes.

| Domain | Owns |
| --- | --- |
| `identity` | Users, signup, activation links, login, refresh sessions, password reset |
| `story` | Chapters, levels, story slides, arenas, enemies, waves and difficulty |
| `combat` | Moves, special moves, which moves fighters and enemies can use |
| `progress` | Fighter profile, level progress, stars, fight sessions |
| `tools` | Tool catalogue, owned tools, mastery, loadout |
| `economy` | Coin wallet, ledger, coin packs, Stripe payments, player market |

Every domain uses the same folders:

```
internal/identity/
  models/       GORM structs mapped to the domain's tables
  repository/   Postgres access through GORM, behind interfaces
  service/      every business rule, exposed through interfaces
  handler/      HTTP handlers (the views): read the request, call the service, write the response
```

- `handler` holds no business rules. It only translates HTTP to service calls and back.
- `service` holds every business rule. Handlers and other domains only ever see its interface.
- `repository` hides GORM behind an interface, so services can be tested without a database.

Domains never reach into each other's tables. When one domain needs another, it calls that domain's service interface. Work that spans domains, such as buying a tool on the market (coins move and the tool changes owner), runs inside one database transaction through the shared `TransactionRunner`.

Shared building blocks live in `backend/internal/platform`: config, logging, the Postgres and Redis connections, the error format, response helpers, request validation, cursor pagination and HTTP middleware.

### Migrations

Migrations are plain SQL files written by hand in `backend/migrations`, numbered in pairs (`000001_create_users.up.sql` and `000001_create_users.down.sql`). GORM is used for queries only, its auto migration is never used. The `cmd/migrate` program applies them, and Docker Compose runs it as a one-shot container before the API starts, so every `make docker-up` brings the database up to date.

```bash
make migrate-create name=create_users
make migrate-up
make migrate-down
make migrate-version
```

### API conventions

- Every route is under `/api/v1`. Resource names are plural (`/users`, `/levels`, `/tools`, `/market-listings`). The few exceptions are `/me`, `/auth/...`, `/health` and webhooks.
- A successful response is `{"data": ...}`.
- An error is `{"error": {"code": "...", "message": "...", "details": {...}}}` with a 4xx or 5xx status. Unknown errors are logged and returned as `INTERNAL_ERROR` without internal details.
- Lists use cursor pagination: `?cursor=&limit=` returns `{"data": [...], "pagination": {"next_cursor": "...", "limit": 20}}`. The default limit is 20 and the maximum is 100.
- State-changing requests must send `X-Nebula-Client: web`. Webhooks are exempt.
- Coin amounts are integers in micro-units (1 coin is 1,000,000) and travel as JSON strings.

### Identity endpoints

| Method and path | Purpose |
| --- | --- |
| `POST /auth/signup` | Create an unverified account and queue the activation email |
| `GET /auth/activations/{token}` | Show which account an activation link belongs to |
| `POST /auth/activations` | Activate the account. The player then logs in |
| `POST /auth/login` | Return an access token and set the refresh cookie |
| `POST /auth/refresh` | Rotate the refresh cookie and return a new access token |
| `POST /auth/logout` | End the current session |
| `GET /me` | The logged in player's profile |

Access tokens last 15 minutes and are sent as `Authorization: Bearer`. The refresh token lives in an httpOnly cookie for 30 days and is replaced on every refresh. If an old refresh token is ever used again, every session in that login chain is ended.

Emails are never sent inside a request. They are written to the `email_outbox` table in the same transaction as the change that caused them, and a background worker sends them with retries. In development they land in Mailpit.

### Tests

`make backend-test` runs every test. Tests that need a database create a fresh one on the Postgres from `make docker-up` and drop it afterwards. Without that database they are skipped; set `REQUIRE_DATABASE_TESTS=true` to make them fail instead.

## Database

The table design is in `backend/docs/database-design.drawio`. Open it with draw.io or the draw.io extension for VS Code. Tables are grouped by the domains above.

## Game and economy decisions

These were agreed before the backend was started. Anything not listed here is still open and gets decided before it is built.

**Currency.** There is one in-game currency, coins. Players buy coins with Stripe in fixed coin packs. Coins can never be turned back into money. A level pays coins only the first time it is cleared; replays give experience but no coins, so coins cannot be farmed.

**Chapters.** Chapter 1 is free. Each later chapter is a one-time unlock paid in coins, so Stripe only ever sells coin packs.

**Selling tools.** Tools are sold player to player on a market. The game takes a 5 percent fee, and that fee is removed from the economy rather than paid to anyone. Listings expire after 7 days. The game does not buy tools back, because that would create coins out of nothing.

**Tool growth.** Every tool is a unique copy with its own mastery. Mastery points come from fights where the tool was used, hits landed with it, and special moves landed with it. A tool has mastery levels 1 to 10, set by its mastery curve. Each level raises its stats, and some levels unlock special moves for that tool. The market sets the price.

**Moves.** Moves are data, not code. A move can be basic, special or a finisher, has its own input sequence, cost and cooldown, and may need a weapon of a certain kind. The story can teach the boy new moves, and tools unlock their own moves as they grow. Enemies use special moves too.

**Difficulty.** Each level has its own difficulty multipliers and star rules (for example: win, finish above half health, finish within 60 seconds). A level can have several enemies in waves, and each wave can be made stronger on its own. The first chapter uses this for the hammer, knife and gun bandits.

**Loadout.** The boy carries two tools: one weapon and one guard.

**Accounts.** Signup sends an activation link. Activation makes the account active and the player can start Chapter 1 right away. There is no onboarding payment.

**Fair play.** The server gives every fight a random seed when it starts and checks that the reported result is believable (time taken, damage dealt and taken, moves used). The input log of every fight is stored, so a full server-side replay check can be added later without changing the data.
