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
make seed
cd frontend && npm install && npm run dev
```

`make docker-up` builds the backend image, starts Postgres, Redis, Mailpit and MinIO, runs the `migrate` container once to apply every pending migration, and then starts the API. `make seed` loads everything under `backend/seeds` (the starter fighter, then every story level with its enemies, slides and images) into Postgres and MinIO. It is safe to run again after editing the content.

| Service | Address |
| --- | --- |
| API | http://localhost:8080/api/v1 |
| Health check | http://localhost:8080/api/v1/health |
| PostgreSQL | localhost:5432 |
| Redis | localhost:6379 |
| Mailpit inbox | http://localhost:8025 |
| MinIO API | http://localhost:9000 |
| MinIO console | http://localhost:9001 |
| Frontend | http://localhost:3000 |

Run `make help` to see every command.

## Backend

The backend is a single Go service split by domain (a modular monolith). Each domain is a package under `backend/internal/` and owns its tables, its rules and its routes.

| Domain | Owns |
| --- | --- |
| `identity` | Users, signup, activation links, login, refresh sessions, password reset |
| `story` | Chapters, levels, story slides, arenas, enemies, waves and difficulty |
| `combat` | Moves, special moves, which moves fighters and enemies can use |
| `progress` | Fighter profile, level progress, stars, fight sessions, chapter unlocks |
| `payment` | Purchase plans, chapter payments, Stripe checkout and webhooks |
| `tools` | Tool catalogue, owned tools, mastery, loadout |
| `economy` | Coin wallet, ledger, coin packs, player market |

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
| `GET /me` | The logged in player: `name`, `email`, `current_level`, `story_level`, `total_win`, `total_lose`, `current_level_win`, `current_level_lose`, `paid_chapters`, `unpaid_chapters` (see below) |

Access tokens last 15 minutes and are sent as `Authorization: Bearer`. The refresh token lives in an httpOnly cookie for 30 days and is replaced on every refresh. If an old refresh token is ever used again, every session in that login chain is ended.

Emails are never sent inside a request. They are written to the `email_outbox` table in the same transaction as the change that caused them, and a background worker sends them with retries. In development they land in Mailpit.

What the `/me` progress fields mean:

- `current_level` is the chapter the player is in, which is also their fighter level. A new player is at 1. It is the chapter of the next level they have not won yet, so it moves to the next chapter once every level of a chapter is won.
- `story_level` is how many levels of the current chapter the player has won: 0 for a new player, 1 after winning level 1.
- `total_win` and `total_lose` count every finished fight in the whole game. `current_level_win` and `current_level_lose` count only fights in the current chapter.
- `current_level_price` is what the current chapter costs, as a string of US cents (`"499"` is USD 4.99), or null for a free chapter.
- `is_current_level_paid` is true when the player owns the current chapter (or it costs nothing), so every level in it can be played.
- `paid_chapters` is how many paid chapters the player owns. `unpaid_chapters` is how many published paid chapters they do not own yet. Free chapters are in neither count.

### Payment endpoints

| Method and path | Purpose |
| --- | --- |
| `GET /chapters` | Every published chapter in order: `id`, `number`, `title`, `is_free`, `price_cents`, `level_count` and `is_owned` for the logged in player |
| `GET /plans` | The ways to buy chapters, each with the exact options and prices for the logged in player |
| `POST /checkout-sessions` | Body `{"plan_id": "chapter-bundle", "chapter_count": 3}`. Starts a Stripe Checkout and returns `id`, `checkout_url` and `expires_at`. The frontend sends the player to `checkout_url` |
| `GET /checkout-sessions/{id}` | The checkout `status`: `OPEN`, `PAID` or `EXPIRED`. The subscription page checks it after Stripe sends the player back |
| `GET /subscriptions` | The player's chapter purchases, newest first: plan, chapters, amounts, `status` (`PAID`, `REFUNDED` or `DISPUTED`), `paid_at` and `refunded_at` |
| `POST /webhooks/stripe` | Stripe events, checked against `STRIPE_WEBHOOK_SECRET`. Not for the frontend |

There are three plans, seeded from `backend/seeds/plans/plans.json`:

- `single-chapter` sells the next chapter the player does not own, at full price.
- `chapter-bundle` lets the player pick how many of the next chapters to buy, from 2 up to one less than all of them. 2 chapters get 5 percent off, 3 or more get 10 percent off.
- `all-chapters` sells every chapter that is out now and not owned yet, with 20 percent off. It needs at least 2 chapters to buy, and it grows by itself as new chapters are published.

Each plan has `is_available` and an `options` list. An option has `chapter_count`, the `chapters` it covers, `subtotal_cents`, `discount_percent`, `discount_cents` and `total_cents`, all money as strings of US cents. The server works out every price from the chapters table and the plan's discount tiers, and a checkout works it out again from `plan_id` and `chapter_count`, so a client can never send its own price. Discounts are rounded to the nearest cent.

How a chapter payment works:

- `POST /checkout-sessions` saves a `payments` row (status `OPEN`) with the exact chapters and prices, then creates a hosted Stripe Checkout for that total, card only, valid for 31 minutes. A plan or count that is not offered returns 422.
- A player has at most one open checkout. Starting a new one expires the older one in Stripe first, so it can never be paid twice. If the older one was already paid, the new request returns 409 and the chapters are unlocked.
- After paying, Stripe sends the player to `/subscription?checkout={id}`. Cancelling sends them to `/fight?checkout=cancelled`.
- The webhook is the source of truth. `checkout.session.completed` marks the payment `PAID` and unlocks its chapters, but only when the paid amount and currency match the payment. Every event id is stored in `stripe_webhook_events`, so a repeated event changes nothing.
- `GET /checkout-sessions/{id}` also asks Stripe while a checkout is open, so the player sees the result even if the webhook is late.
- A full refund (`charge.refunded`) or a dispute (`charge.dispute.created`) locks the chapters of that payment again. Wins and stars stay. A chapter that another paid purchase also covers stays unlocked. A partial refund changes nothing.
- Without `STRIPE_SECRET_KEY` and `STRIPE_WEBHOOK_SECRET`, checkout and the webhook return 503. Plans and chapters still load.

To test payments locally, run `make stripe-listen` in a second terminal. It forwards Stripe test events to the backend; the signing secret it prints must be the `STRIPE_WEBHOOK_SECRET` in `.env`. Pay with the test card `4242 4242 4242 4242`, any future date and any CVC.

### Story endpoints

| Method and path | Purpose |
| --- | --- |
| `GET /stories?level={level_id}` | The story slides of one level, in order. `level` is required |

The list is cursor paginated like every other list. Each item has `id`, `position`, `kind` (`SLIDE` or `CALL_TO_ACTION`), `eyebrow`, `heading`, `body`, `image`, `palette` and `button_label`. The call to action is always the last item and is the only one with a button label. A level that does not exist or is not published returns 404.

Images live in a private MinIO bucket. `image` is a presigned link that works for one hour, so the page should load the story again rather than keep links around.

### Fight endpoints

| Method and path | Purpose |
| --- | --- |
| `GET /me/next-level` | The next level to play: `status` (`AVAILABLE`, `LOCKED` for a chapter that has to be bought, `COMING_SOON` when the next chapter has no levels yet), `chapter` (number and title) and `level` (id, number, title, teaser, time limit, best stars, attempts). The fight hub shows it as the next fight |
| `GET /levels/{level_id}/fight-setup` | Everything the fight screen needs: time limit, arena and stage, the player's fighter and the enemy of the first wave |
| `POST /fight-sessions` | Body `{"level": "1-1"}`. Called when the player presses Play on the last story slide. Records that the level was started but not finished |
| `POST /fight-sessions/{id}/results` | Body `{"outcome": "WON", "duration_ms": 40000, "damage_dealt": 95, "damage_taken": 30}`. Sent once when the fight ends |

The response has the fight `id`, `level`, `status` (`STARTED`), `started_at` and a server `seed` (as a string, because it does not fit in a JavaScript number). Rules:

- Chapter 1 level 1 is always open. Every other level opens only after the level before it has been won, across chapters too. Levels in a paid chapter need the chapter to be owned, unless the level itself is free. A level waiting on the previous win returns 403 `LEVEL_LOCKED`; a level in a chapter the player does not own returns 403 `CHAPTER_LOCKED`. `/me/next-level` shows the chapter `price` (US cents) and `is_paid`.
- Pressing Play again on a level with an unfinished fight marks that fight `ABANDONED` and starts a new one. There is never more than one open fight per player and level.
- Every start adds one attempt to the player's progress for that level.

How a result is handled:

- The frontend reports what happened. The server decides whether it is believable and works out the stars itself. Rewards are always worked out on the server.
- A fight can only be won by knocking the enemy out. When time runs out it is a loss, whatever the health bars show.
- A win adds 1 to `total_win`, marks the level completed (which unlocks the next level) and keeps the best stars. A loss adds 1 to `total_lose`. Quitting or closing the tab changes nothing.
- Stars come from the level's time slots, counted in fight time, so pauses do not count. Level 1: 3 stars within 45 seconds, 2 within 70, 1 within 90.
- A report is rejected (`FIGHT_RESULT_REJECTED`, nothing counted, reason stored on the fight) when, for example, the enemy was not knocked out in a win, the fight is longer than the time limit or than the real time since it started, the damage is more than the fighters can deal in that time, or the fight was paused for more than 10 minutes in total.
- A fight takes one result. A second one returns 409.

The player's fighter comes from their `fighter_profiles` row. A player gets that row the first time they press Play, copied from the default fighter template (the boy). Until then the fight setup and `/me` use the template itself, so `current_level` starts at the template's starting level, which is 1.

### Game content

`backend/seeds/fighters/*.json` holds fighter templates. Exactly one of them is the default for new players.

The purchase plans live in `backend/seeds/plans/plans.json`. The seed checks the plan kinds and discount tiers and updates plans by id.

Each level lives in `backend/seeds/story/<level>/`: a `level.json` with the chapter, arena, level, enemy waves and slides, and an `images` folder. The seed checks the package before it changes anything, uploads the images under `story/<level_id>/`, then saves the rows in one transaction.

### Tests

`make backend-test` runs every test. Tests that need a database create a fresh one on the Postgres from `make docker-up` and drop it afterwards. Without that database they are skipped; set `REQUIRE_DATABASE_TESTS=true` to make them fail instead.

## Database

The table design is in `backend/docs/database-design.drawio`. Open it with draw.io or the draw.io extension for VS Code. Tables are grouped by the domains above.

## Game and economy decisions

These were agreed before the backend was started. Anything not listed here is still open and gets decided before it is built.

**Currency.** There is one in-game currency, coins, used for tools and the market. 1 coin is worth 1 US cent. Players buy coins with Stripe in fixed coin packs: 500 coins for USD 4.99, 1,100 for USD 9.99 and 2,400 for USD 19.99. Coins can never be turned back into money. Chapters are paid with a card directly, not with coins. A level pays coins only the first time it is cleared; replays give experience but no coins, so coins cannot be farmed.

**Chapters.** Only chapter 1 level 1 is free. Every other level needs its whole chapter, bought once with a card through Stripe (chapter 1 costs USD 4.99). A player can buy the next chapter, a number of chapters or all of them at once: 2 chapters get 5 percent off, 3 or more get 10 percent off and all remaining chapters get 20 percent off. A refunded or disputed chapter payment locks those chapters again; wins and stars stay. In development `make dev-unlock-chapter` also gives a chapter to a player without paying. Coins (1 coin is 1 US cent, sold in packs) stay for tools and the market.

**Selling tools.** Tools are sold player to player on a market. The game takes a 5 percent fee, and that fee is removed from the economy rather than paid to anyone. Listings expire after 7 days. The game does not buy tools back, because that would create coins out of nothing.

**Tool growth.** Every tool is a unique copy with its own mastery. Mastery points come from fights where the tool was used, hits landed with it, and special moves landed with it. A tool has mastery levels 1 to 10, set by its mastery curve. Each level raises its stats, and some levels unlock special moves for that tool. The market sets the price.

**Moves.** Moves are data, not code. A move can be basic, special or a finisher, has its own input sequence, cost and cooldown, and may need a weapon of a certain kind. The story can teach the boy new moves, and tools unlock their own moves as they grow. Enemies use special moves too.

**Difficulty.** Each level has its own difficulty multipliers and star rules (for example: win, finish above half health, finish within 60 seconds). A level can have several enemies in waves, and each wave can be made stronger on its own. The first chapter uses this for the hammer, knife and gun bandits.

**Loadout.** The boy carries two tools: one weapon and one guard.

**Accounts.** Signup sends an activation link. Activation makes the account active and the player can start Chapter 1 right away. There is no onboarding payment.

**Fair play.** The server gives every fight a random seed when it starts and checks that the reported result is believable (time taken, damage dealt and taken). Next, the engine will use that seed and record an input log, so a full server-side replay check can be added later without changing the data.

**Fight results.** Knock the enemy out to win; running out of time is a loss. Wins and losses are counted for the player, a loss can be replayed straight away, and quitting is not counted. Pauses can add up to at most 10 minutes per fight. Stars depend on how fast the fight was won (level 1: 45, 70 and 90 seconds for 3, 2 and 1 stars). Rewards stay at 0 until the economy is designed.
