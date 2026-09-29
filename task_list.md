# Nebula Exchange — Task List

Tasks are small vertical slices. We finish each one (UI, backend and contract, as the task needs), test it, and only then start the next. The order is the build order: each task builds only on tasks above it.

**Layer tags:** `UI` Next.js frontend · `BE` Go backend · `SC` Solidity contracts · `INFRA` tooling, CI, deployment. An arrow such as `UI → BE` gives the order the layers are built in.

## Workflow for every task

1. **UI first.** If a task has UI work, I build the UI first against mock data (MSW handlers that match the planned API contract), so every screen can be clicked through before any backend exists.
2. **⏸ Checkpoint.** I stop and ask you to verify the UI. I don't start any backend or Solidity work for that task until you say **"go"**. Changes you ask for are made in the UI and then shown to you again.
3. **Backend and/or Solidity.** After your "go", I build the backend and contract parts, swap the mocks for the real API or contract, and test it end to end.
4. **Done.** Tests pass, it runs locally, and the status here is set to Done.

Tasks with nothing visible to review (infra, config, pure logic, jobs) have no checkpoint. Docker config is written but containers are only started when you say so. Commits are made per scope (frontend, backend, smart-contract, infra), so one task can have several commits. UI-only tasks still get a checkpoint before they're marked Done.

**Definition of done**
- Code follows the rules in `CLAUDE.md` (no comments, meaningful names, modular files, shared utils).
- Code is written with tests: Go unit tests, Foundry tests, and component tests where they make sense.
- It runs locally: backend services in Docker, frontend and Anvil on the host.
- Money and item changes go through the ledger only. No floats anywhere.
- The mocks for the task's endpoints are replaced by the real API. The mock handlers stay available for UI development.

**Open decisions (proposed defaults for PRD gaps — NOT confirmed; each needs your approval before its task is built)**
1. NC has four buckets: `card`, `crypto`, `earned_pending`, `earned`. Spending order is `card → earned_pending → earned → crypto`.
2. The contract circular dependency is solved by deploying `NebulaItems` first, then the Vault, then calling `NebulaItems.setVault(vault)`. This setter can only be called once.
3. A pending player can pay the entry fee from `crypto` NC, so an underpaid USDC entry fee can be completed later.
4. A market order on an empty opposite book is rejected with `NO_LIQUIDITY`.
5. The withdrawal review value for legendaries uses the last auction price.
6. The cargo cap uses a `rarity_rank` column on items. Common resources are reduced first.
7. When a resting buy order fills as maker, the fee difference between taker and maker is released from the hold.
8. The Vault gets an extra `payWithPermit()` so a USDC payment can be a single transaction.

---

## Milestone 0 — Foundation

### T-001 · Monorepo scaffold
**Layers:** INFRA  
**Status:** Done

**Description:** One git repo with `frontend/`, `backend/` and `smart-contract/`, plus a root `Makefile`, `.gitignore`, `.editorconfig`, `README.md` and `CLAUDE.md` (coding rules). The Makefile starts with a `help` target and grows as each task adds commands.

**Outcome:** A clean repo where each app has its own folder and one command lists the available dev tasks.

### T-002 · Local dev environment (docker-compose)
**Layers:** INFRA  
**Status:** Done

**Description:** A `docker-compose.yml` for backend-related services only, using the latest images: `postgres:latest`, `redis:latest` and `axllent/mailpit:latest` (catches outgoing email). Named volumes, health checks and a root `.env.example`. The frontend and smart contracts run outside Docker: `npm run dev` for the frontend, and Anvil from the local Foundry install (`make chain`).

**Outcome:** `make docker-up` starts Postgres, Redis and Mailpit, each on a known port. Containers are only started when you say so.

### T-003 · Go backend skeleton
**Layers:** BE  
**Status:** Todo

**Description:** `cmd/server/main.go` with Gin, env config loading, a structured logger (slog), request-ID middleware, `GET /api/v1/health` and graceful shutdown. Sets up the `internal/platform` package layout from PRD section 11.

A `backend/Dockerfile` and a `backend` service in docker-compose that depends on Postgres and Redis and reaches the host's Anvil through `host.docker.internal`.

**Outcome:** The backend runs in Docker next to its services, serves health with a request ID in the logs, and shuts down cleanly on SIGTERM.

### T-004 · Database tooling
**Layers:** BE  
**Status:** Todo

**Description:** A pgx connection pool, golang-migrate with an `up`/`down` Make target, sqlc config and code generation, and a test helper that creates an isolated test database per package.

**Outcome:** Migrations run, sqlc generates typed queries, and `go test` can use a real Postgres.

### T-005 · API response envelope and validation
**Layers:** BE  
**Status:** Todo

**Description:** Success responses as `{"data": ...}`, the error response format and fixed error-code list, a request-binding and validation helper, and a shared pagination helper used by every list endpoint.

**Outcome:** Every handler returns consistent errors, and there's a shared way to validate input and paginate lists.

### T-006 · Next.js skeleton and theme
**Layers:** UI  
**Status:** Todo

**UI:** Next.js App Router + TypeScript + Tailwind with the dark space theme tokens (colors, up/down colors, fonts), ESLint + Prettier, and a base layout with a placeholder page.

⏸ **Checkpoint:** you verify the theme and base layout.

**Outcome:** `npm run dev` shows the themed placeholder page, and lint and typecheck pass.

### T-007 · UI primitives
**Layers:** UI  
**Status:** Todo

**UI:** Button, Input, Select, Card, Modal, ConfirmDialog, Tabs, Toast/ToastProvider, Skeleton, EmptyState, ErrorState, StatusBadge. Shown on a `/dev/ui` preview page.

⏸ **Checkpoint:** you review each component on `/dev/ui`.

**Outcome:** A reusable component set that every later screen is built from.

### T-008 · API client, mock layer and money formatting
**Layers:** UI  
**Status:** Todo

**UI:** A typed fetch client (base URL, JSON, error-envelope parsing, `Idempotency-Key` on POST, custom CSRF header) and a TanStack Query provider. An MSW mock layer, switched on with `NEXT_PUBLIC_API_MOCKS=true`, where every later task adds its handlers first. A decimal library helper for micro-units, plus the `NcAmount` and `PriceChange` components, added to `/dev/ui`.

⏸ **Checkpoint:** you verify amount and price-change formatting on `/dev/ui`.

**Outcome:** Screens can be built against mocks before the backend exists, and amounts display exactly, with no JS floats.

### T-009 · Foundry project skeleton
**Layers:** SC  
**Status:** Todo

**Description:** `forge init`, OpenZeppelin v5 install, remappings, `foundry.toml` profiles (default, ci), `forge fmt` config, and an empty test to confirm the setup.

**Outcome:** `forge build` and `forge test` pass on an empty project.

### T-010 · CI pipeline
**Layers:** INFRA  
**Status:** Todo

**Description:** GitHub Actions jobs: Go lint and test with a Postgres service, `forge test`, and frontend lint, typecheck and build.

**Outcome:** Every push runs all three test suites automatically.

### T-011 · Background job scheduler
**Layers:** BE  
**Status:** Todo

**Description:** An in-process scheduler in `platform/scheduler`: register a job with an interval, a Postgres advisory lock so only one instance runs jobs, panic recovery, and logging per run.

**Outcome:** Any module can register a periodic job, and it runs safely on one leader instance.

---

## Milestone 1 — Auth and account

### T-012 · Email module
**Layers:** BE  
**Status:** Todo

**Description:** A `notify/email` interface with an SMTP implementation (Mailpit in dev) and HTML/text templates for codes and links.

**Outcome:** The backend can send templated emails, visible in Mailpit.

### T-013 · Signup
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/signup` page with email, username, password with a strength meter, the 18+ and Terms checkbox, inline field errors, and a mocked `POST /auth/signup` (success, username taken, email taken).

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `users` migration. `POST /auth/signup` with username rules, case-insensitive uniqueness, argon2id and status `UNVERIFIED`. Sends a verification code.

**Outcome:** A user can sign up and receives a 6-digit code by email.

### T-014 · Email verification
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/verify` page with a 6-digit code input, a resend button with a 60-second countdown, and attempts left. Mocks cover success, wrong code and expired code.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `email_codes` table (hashed code, 15-minute expiry, max 5 attempts), `POST /auth/verify-email` and `POST /auth/resend-code` with a 60-second cooldown.

**Outcome:** A user can verify their email and becomes `PENDING_PAYMENT`.

### T-015 · Login and sessions
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/login` page. A frontend auth provider that keeps the access token in memory, refreshes it silently, and exposes `useMe()`. Mocked login, refresh and `GET /me`.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `POST /auth/login` returns an access JWT (15 min) and a refresh cookie (httpOnly, Secure, SameSite=Strict). `refresh_tokens` table and `GET /me`.

**Outcome:** A user can log in and stays logged in across page reloads through the refresh cookie.

### T-016 · Refresh rotation, reuse detection and logout
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A logout item in the profile menu, and a "session expired, please log in again" state.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `POST /auth/refresh` rotates the token on every use. Reusing an old token revokes the whole session chain. `POST /auth/logout`.

**Outcome:** Refresh tokens are single-use. A stolen token being reused logs out all of that user's sessions.

### T-017 · Rate limiter and login lockout
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A lockout message on the login page ("Try again in 14 minutes") and a generic "too many requests" toast for `RATE_LIMITED`.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** Generic Redis rate-limit middleware (per IP or per user, configurable), applied to signup (5/hour/IP) and login (10 per 15 min per IP). Locks an email for 15 minutes after 5 failed logins.

**Outcome:** Brute-force protection is in place, and the limiter can be reused for later limits.

### T-018 · Forgot and reset password
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/forgot` email form and `/reset?token=` new-password form, including the expired-link state.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `POST /auth/forgot-password` sends a single-use link valid for 30 minutes. `POST /auth/reset-password` changes the password and revokes all sessions.

**Outcome:** A user can recover access to their account by email.

### T-019 · Account-state guards
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** Route guards that redirect by state (unverified to `/verify`, pending to `/onboarding/pay`) and an `AccountStateBanner` for `PENDING_PAYMENT` and `FROZEN`. The mocked `GET /me` can switch between account states.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** Middleware such as `RequireAuth`, `RequireState(ACTIVE)` and `RequireAdmin`.

**Outcome:** Permissions are enforced on the server, and the UI sends users to the right place for their state.

### T-020 · App shell
**Layers:** UI  
**Status:** Todo

**UI:** The logged-in layout: `TopNav` with all section links, `ProfileMenu`, placeholder slots for `BalanceChip` and `NotificationBell`, and a mobile nav drawer. Each section gets an empty placeholder page.

⏸ **Checkpoint:** you verify the layout on desktop and at 375 px.

**Outcome:** A navigable app frame that works on phones.

### T-021 · Idempotency, CSRF, CORS and security headers
**Layers:** BE + frontend config  
**Status:** Todo

**Description:** `idempotency_keys` table and middleware (stores the request hash and response for 24 hours, returns the stored response on a retry). Custom-header CSRF check on state-changing requests. CORS limited to the frontend origin. CSP and HSTS headers in Next.js.

**Outcome:** A double-submitted POST performs the action once, and the API is protected against cross-site requests.

### T-022 · Two-factor authentication (TOTP)
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A 2FA section in settings with the QR code, secret and verify step, plus disable. A reusable `TwoFactorPrompt` modal for sensitive actions.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `POST /auth/2fa/setup|enable|disable` with the secret encrypted at rest, a `RequireTOTP` helper, and an email when 2FA changes.

**Outcome:** A user can turn 2FA on or off, and later features can ask for a TOTP code.

### T-023 · Login with 2FA
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A TOTP step on the login page after the password step.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** If 2FA is on, login returns a short-lived challenge, and the TOTP code completes the login.

**Outcome:** Accounts with 2FA need a code to log in.

### T-024 · Settings: profile, password and sessions
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/settings` page with username change, password change, and a list of active sessions with revoke.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `PATCH /me`, password change, session list and revoke endpoints.

**Outcome:** A user can manage their profile and active sessions.

### T-025 · Cleanup job
**Layers:** BE  
**Status:** Todo

**Description:** A daily job that deletes unverified accounts older than 7 days, and expired codes, nonces and idempotency keys.

**Outcome:** Stale auth data is removed automatically.

---

## Milestone 2 — Ledger core

### T-026 · Ledger schema
**Layers:** BE  
**Status:** Todo

**Description:** Migrations for `ledger_accounts`, `ledger_journals` (unique on type + ref_type + ref_id), `ledger_entries`, `ledger_balances` (CHECK available ≥ 0 and held ≥ 0, with the dispute exception) and `ledger_holds`.

**Outcome:** The database itself rejects invalid balances and duplicate journals.

### T-027 · System accounts and account resolution
**Layers:** BE  
**Status:** Todo

**Description:** Seed the system accounts (`stripe_clearing`, `crypto_clearing`, `fees`, `treasury`, `mint`, `burn`, `withdrawn`, `unclaimed`). `GetOrCreateAccount(owner, asset, bucket)`.

**Outcome:** Every owner, asset and bucket combination maps to exactly one account.

### T-028 · ledger.Post
**Layers:** BE  
**Status:** Todo

**Description:** Post a journal inside the caller's transaction: check that entries sum to zero per asset, lock balances in ascending account-ID order, apply the entries, and return a clear error on a duplicate or a negative balance.

**Outcome:** The only way to move value, and it is atomic and can't be posted twice.

### T-029 · Hold, Release and Capture
**Layers:** BE  
**Status:** Todo

**Description:** Hold rows that move value from available to held, with partial release and partial capture to target accounts, and status transitions `ACTIVE → CAPTURED/RELEASED`.

**Outcome:** Orders, bids, missions and withdrawals can reserve value safely.

### T-030 · Spend helper and bucket logic
**Layers:** BE  
**Status:** Todo

**Description:** `Spend(user, amount, ref)` debits across buckets in the fixed order. Helpers for total, spendable and withdrawable NC per user.

**Outcome:** Every NC debit follows the same bucket rules.

### T-031 · Ledger test suite
**Layers:** BE  
**Status:** Todo

**Description:** Unit tests for every function, concurrency tests (parallel holds and spends on the same account, no deadlocks, no overspend), and randomized property tests (totals stay balanced).

**Outcome:** High confidence that the ledger can't create or lose value.

### T-032 · ledger_check job
**Layers:** BE  
**Status:** Todo

**Description:** An hourly job that checks every journal sums to zero, every balance equals the sum of its entries, and held equals the sum of active holds. Logs an alert on any mismatch.

**Outcome:** Any ledger bug is detected automatically.

### T-033 · Balances and BalanceChip
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `BalanceChip` in the TopNav with a bucket-breakdown popover (card, crypto, earned pending, earned, held), using a mocked `GET /me/balances`.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `GET /me/balances` returns totals per bucket and held amounts. A dev-only seed command credits test NC.

**Outcome:** A player always sees their NC total and where it came from.

### T-034 · Ledger history
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/history` page with the Ledger tab: a sortable, paginated `DataTable` of journals with type and amounts, from mock data.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `GET /me/ledger?cursor=` lists the player's journals with their entries.

**Outcome:** A player can audit every movement of their NC and items.

---

## Milestone 3 — Catalog and inventory

### T-035 · Catalog tables and seed
**Layers:** BE  
**Status:** Todo

**Description:** Migrations and seed data for `items` (with `rarity_rank`, attributes), `recipes`, `upgrades`, `zones`, `loot_tables` and `shop_skus`, using the section 4 values.

**Outcome:** All game config lives in the database and can be changed without a deploy.

### T-036 · Item components and catalog APIs
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `ItemIcon` and `ItemCard` (icon, name, tier, quantity) with placeholder art for every category, shown on `/dev/ui` with mocked catalog data.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `GET /items`, `/items/{id}`, `/recipes`, `/upgrades`, `/zones`, `/shop-items`, with caching.

**Outcome:** The frontend can show any item consistently.

### T-037 · Inventory
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/inventory` page with a category filter and an `ItemCard` grid showing available vs held quantities. Item action buttons are placeholders until their features exist.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `GET /me/inventory` returns available and held quantities per item.

**Outcome:** A player can see everything they own and what is currently locked.

### T-038 · Item detail page (basic)
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/items/[id]` with the image, attributes, and where the item is used (recipes, upgrades). Supply and price sections come later.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** A "used in" lookup added to `GET /items/{id}`.

**Outcome:** Every item has a public info page.

### T-039 · Token metadata endpoint
**Layers:** BE  
**Status:** Todo

**Description:** `GET /metadata/{id}.json` in OpenSea format, served from the catalog. Static image hosting.

**Outcome:** The ERC-1155 URI resolves to valid metadata.

---

## Milestone 4 — Activation with card payments

### T-040 · Payments table and purpose runner
**Layers:** BE  
**Status:** Todo

**Description:** `payments` and `external_events` migrations. A `purpose` package with `ENTRY_FEE`, `TOPUP` and `SHOP_PURCHASE` handlers, each running as an NC debit inside the same transaction as the credit. `POST /payments`, `GET /payments/{id}`, `GET /payments`.

**Outcome:** One code path runs the result of every payment, whichever rail it came from.

### T-041 · Activation and starter pack
**Layers:** BE  
**Status:** Todo

**Description:** The `ENTRY_FEE` purpose debits 5 NC, sets the account to `ACTIVE`, and grants the starter pack (Scout, Drill T1, 10 Fuel, 2 NC card bonus) as `STARTER_PACK` journals in the same transaction. Tested with a direct credit.

**Outcome:** Activation is atomic: either everything is granted or nothing is.

### T-042 · Entry fee page and payment result (card)
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/onboarding/pay` with a starter pack preview, the price and the card option. `/payment/result?id=` polls every 2 seconds for up to 60 seconds and shows success, failure or retry. Mocks simulate a Stripe redirect and each result.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** Create a Stripe Checkout Session with `metadata.payment_id`. `POST /webhooks/stripe` verifies the signature, stores the event ID (duplicates ignored), and on `checkout.session.completed` marks the payment succeeded, credits `card`, and runs the purpose. Tested with the Stripe CLI.

**Outcome:** A new player can pay by card and lands on an active account. **This is the first end-to-end player journey.**

### T-043 · Hangar v1 and onboarding checklist
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/hangar` with a balance summary and an onboarding checklist (paid, first mission, first craft, first trade).

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** An endpoint that derives checklist progress from the player's history.

**Outcome:** A newly active player has a home screen that shows what to do next.

### T-044 · payment_expiry job
**Layers:** BE  
**Status:** Todo

**Description:** Expires pending card payments after 24 hours and crypto payments after 30 minutes.

**Outcome:** Stale payment intents never stay pending.

---

## Milestone 5 — Wallet and crypto payments

### T-045 · MockUSDC contract
**Layers:** SC  
**Status:** Todo

**Description:** ERC-20 with 6 decimals, ERC-2612 permit, and `faucet()` that mints 1,000 test USDC per address per day. Full tests.

**Outcome:** Testers have free test USDC.

### T-046 · NebulaVault v1: pay
**Layers:** SC  
**Status:** Todo

**Description:** Vault with AccessControl, Pausable and ReentrancyGuard. `pay(ref, amount)` and `payWithPermit(...)`, a used-ref mapping, `PaymentReceived` event, pause. Unit, revert, replay and fuzz tests.

**Outcome:** An on-chain payment entry point where a ref can't be used twice.

### T-047 · Deploy script v1 and address export
**Layers:** SC + INFRA  
**Status:** Todo

**Description:** `script/Deploy.s.sol` for MockUSDC and the Vault on Anvil and Base Sepolia. Writes `deployments/<chainId>.json`. Basescan verification.

**Outcome:** One command deploys the contracts, and both apps read the addresses from a single file.

### T-048 · Contract bindings pipeline
**Layers:** SC + BE + frontend codegen  
**Status:** Todo

**Description:** Generate Go bindings (abigen) and TypeScript ABIs (wagmi CLI) from the forge output, run by a `make bindings` target.

**Outcome:** The backend and frontend use typed contract calls that always match the deployed ABI.

### T-049 · Wallet connect
**Layers:** UI  
**Status:** Todo

**UI:** wagmi + viem config (Anvil and Base Sepolia), `WalletConnectButton` supporting MetaMask, Coinbase Wallet and WalletConnect, and a chain-switch prompt when the wallet is on the wrong network.

⏸ **Checkpoint:** you connect a real wallet and verify the flow.

**Outcome:** A player can connect a wallet and is kept on the correct chain.

### T-050 · SIWE wallet linking
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `WalletLinkCard` on `/wallet`: connect, sign the SIWE message, and show the linked address. Mocks cover success and "address already linked to another account".

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `wallets` and `siwe_nonces` tables. `POST /wallets/nonces` and `POST /wallets` verify the signature, nonce, domain and chain, and enforce one wallet per user and one user per address. Confirmation email.

**Outcome:** A player proves ownership of an address and links it to their account.

### T-051 · Wallet change and unlink
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** Change and unlink actions on `WalletLinkCard` using `TwoFactorPrompt`, plus a notice about the 48-hour withdrawal block.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** Changing the wallet requires 2FA, sends an email, and sets a 48-hour withdrawal block. `DELETE /wallets/{id}` is allowed only when there are no pending withdrawals.

**Outcome:** Wallet changes are protected against account takeover.

### T-052 · Test USDC faucet panel
**Layers:** UI → SC  
**Status:** Todo

**UI (step 1):** `TestUsdcFaucet` on `/wallet` (testnet only) with a claim button, the USDC balance and the next-claim time, using a mocked contract response.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Solidity (step 2):** Wire the panel to the deployed MockUSDC `faucet()` and `balanceOf`.

**Outcome:** Testers can get USDC in one click.

### T-053 · Chain indexer framework
**Layers:** BE  
**Status:** Todo

**Description:** Polls `eth_getLogs` every 3 seconds from a stored block cursor, waits 5 confirmations, dedupes by tx hash + log index in `external_events`, dispatches to handlers, and retries with backoff when the RPC is down.

**Outcome:** A reliable pipeline from on-chain events to backend handlers.

### T-054 · Crypto entry fee payment
**Layers:** UI → BE + SC  
**Status:** Todo

**UI (step 1):** A reusable crypto pay flow (permit or approve, then `pay`) and `TxStatus` with a Basescan link and a "Confirming (n/5)" count. Adds the Wallet option to the entry fee page, with mocked confirmations.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend + Solidity (step 2):** A crypto intent (payment_ref = keccak256(uuid), exact amount, 30-minute expiry). A `PaymentReceived` handler that credits `crypto` and runs the purpose, covering underpay, overpay, late payment and wrong-payer cases. The UI is wired to the real Vault.

**Outcome:** A player can pay the entry fee with USDC end to end.

### T-055 · PaymentMethodPicker
**Layers:** UI  
**Status:** Todo

**UI:** A generic picker (NC balance, Card, Wallet) that disables options not allowed for the current use case. Replaces the ad-hoc choices on the entry fee page.

⏸ **Checkpoint:** you verify the picker in each use case on `/dev/ui`.

**Outcome:** One component handles payment choice everywhere.

### T-056 · NC top-up
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `TopUpPanel` on `/wallet` with amount presets (5, 10, 25, 50, 100), card or wallet, and the limit-reached state.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** The `TOPUP` purpose for card and crypto, with daily and 30-day card limits.

**Outcome:** A player can add NC by card or USDC.

### T-057 · Payments history
**Layers:** UI  
**Status:** Todo

**UI:** A Payments tab on `/history` with status, method, amount, and a tx hash or Stripe reference, using the existing `GET /payments`.

⏸ **Checkpoint:** you verify the UI.

**Outcome:** A player can see every payment they made.

---

## Milestone 6 — Shop

### T-058 · Shop page
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/shop` with a `ShopItemCard` grid.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `GET /shop-items` returns SKUs with server-side prices.

**Outcome:** A player can browse what the system sells.

### T-059 · Buy with NC
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `PurchaseModal` with quantity, total and a confirmation, plus the insufficient-funds state.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `POST /shop-items/{sku}/purchases` takes a quantity, charges the server-side price through Spend, and mints the items as a `SHOP_PURCHASE` journal.

**Outcome:** A player can buy fuel, ships, drills and bundles from their balance.

### T-060 · Buy by card or USDC
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `PaymentMethodPicker` in `PurchaseModal` with the Card and Wallet flows.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** The `SHOP_PURCHASE` purpose through `POST /payments` for both rails. Tests show that all three payment methods produce the same ledger result.

**Outcome:** The shop accepts all three payment methods, meeting that acceptance criterion.

---

## Milestone 7 — Gameplay

### T-061 · Loot engine (pure logic)
**Layers:** BE  
**Status:** Todo

**Description:** A pure Go function for loot: uniform rolls with `crypto/rand`, drill multiplier, cargo cap reduction by `rarity_rank`, the Void Shard chance, and duration by ship speed. Table-driven tests.

**Outcome:** Loot rules are correct and easy to tune.

### T-062 · Zones list
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/missions` with a `ZoneCard` list showing duration, fuel, loot ranges, requirements and a locked state.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** A per-player "unlocked" flag on `GET /zones`, computed from the player's inventory.

**Outcome:** A player can see which zones they can run.

### T-063 · Start mission
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `MissionLauncher` to pick ship, drill and zone, showing the fuel cost and expected loot range, with a confirmation.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `missions` table. `POST /missions` validates the requirements and the 3-mission limit, burns fuel (`MISSION_FUEL`), holds the ship and drill, and snapshots the zone config on the row.

**Outcome:** A player can launch a mission, and their ship and drill are locked.

### T-064 · mission_resolver job
**Layers:** BE  
**Status:** Todo

**Description:** A 5-second job using `SKIP LOCKED` that rolls loot for missions past `ends_at` and stores it on the row with status `COMPLETED`.

**Outcome:** Loot is decided at the end time and can't be re-rolled by refreshing.

### T-065 · Collect and abort
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `ActiveMissionList` with `CountdownTimer`, a Collect button with a loot reveal, and Abort with a warning that fuel isn't refunded.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `POST /missions/{id}/collect` mints the loot (`MISSION_LOOT`) and releases the ship and drill holds. `POST /missions/{id}/abort` gives no refund and no loot.

**Outcome:** The full mission loop works end to end.

### T-066 · Mission history
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A `MissionHistory` list on `/missions` and a Missions tab on `/history`.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `GET /missions` with status filters and cursor pagination.

**Outcome:** A player can review past runs and their loot.

### T-067 · Start craft
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/workshop` Crafting tab with `RecipeCard` (inputs owned vs needed, time, fee), a quantity stepper and a Craft button.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `craft_jobs` table. `POST /crafts` burns the inputs and fee immediately (`CRAFT_START`), with batch quantity 1–100 and one active queue.

**Outcome:** A player can queue a craft.

### T-068 · Craft queue and craft_resolver
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `CraftQueue` with a countdown, and a Crafts tab on `/history`.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** A job that delivers the outputs (`CRAFT_OUTPUT`) when the timer ends. `GET /crafts`.

**Outcome:** Crafted components arrive in the inventory automatically.

### T-069 · Upgrade by crafting
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** The Upgrades tab with `UpgradeCard` (current → next tier, materials owned vs needed, fee).

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `POST /upgrades/{id}/craft` burns the current item and materials plus the fee, and mints the next tier (`UPGRADE`). Items on hold are rejected.

**Outcome:** A player can upgrade drills and ships with materials.

### T-070 · Upgrade by buying
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A buy path on `UpgradeCard` with `PaymentMethodPicker` (NC, Card, Wallet).

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** The buy-path upgrade through the shop and payments flow for all three methods.

**Outcome:** The paid upgrade feature works with all three payment methods.

### T-071 · In-app notifications
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `NotificationBell` with an unread count and dropdown, and a `/notifications` page with mark-all-read and links to the related page.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `notifications` table, `GET /me/notifications`, mark-as-read. Wired to payment, mission and craft events, with email where section 13 requires it.

**Outcome:** Players get notified about game and payment events.

### T-072 · WebSocket connection and status
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A shared WS client with reconnect, a `ConnectionStatus` indicator, and a fallback to polling every 3 seconds, tested against a mock socket.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `GET /ws/ticket` (30-second ticket), `GET /ws`, subscribe and unsubscribe, the private `user` channel, and a publisher interface for modules.

**Outcome:** The server can push live updates to the browser.

### T-073 · Live game updates
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** TanStack Query caches and toasts update from `user` channel events (balance change, mission or craft done, notification), driven by mock events.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** Publish those events from the ledger, mission, craft and notify modules.

**Outcome:** Timers finishing and balance changes show up without a refresh.

### T-074 · Hangar v2
**Layers:** UI  
**Status:** Todo

**UI:** Adds running missions with a Collect button, the craft queue, and recent notifications to the dashboard, using the existing APIs.

⏸ **Checkpoint:** you verify the UI.

**Outcome:** The Hangar shows the player's whole game state at a glance.

---

## Milestone 8 — Exchange

### T-075 · Order book data structure
**Layers:** BE  
**Status:** Todo

**Description:** A pure in-memory book: bids and asks as sorted price levels, each a FIFO queue, with O(log n) best price, add and remove. Unit tests.

**Outcome:** A fast, correct book with no dependency on the database.

### T-076 · Matcher (pure logic)
**Layers:** BE  
**Status:** Todo

**Description:** GTC, IOC and market orders with protection price, maker-price execution, self-trade prevention, fee calculation (rounded up), and hold-release amounts. Unit tests, the PRD worked example, and property tests.

**Outcome:** Matching is proven correct before it touches the database.

### T-077 · Markets page
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/exchange` with `MarketsTable` (symbol, last, 24h change, high, low, volume, sparkline, sortable columns), Top gainers, Top losers and Volume tabs, search and a category filter, from mock market data.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `markets` migration, seeded with one market per non-legendary item and its tick size and limits. `GET /markets` and `/markets/{symbol}` (24h stats come in T-085).

**Outcome:** All trading pairs exist and can be browsed.

### T-078 · Order form and place order
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/exchange/[symbol]` layout (chart, book and trades placeholders) with `OrderForm`: buy and sell, limit and market, IOC toggle, price and quantity, %-of-balance buttons, fee and total preview, available balance, and confirmation.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `orders` and `trades` tables. `POST /orders` validates the order (state, market open, tick, limits, 50 open orders), holds funds or items, assigns `seq`, and returns `202`. Order rate limit of 10 per second.

**Outcome:** Orders are accepted from the UI with correct holds.

### T-079 · Engine goroutine and settlement
**Layers:** BE  
**Status:** Todo

**Description:** One goroutine per market with a command channel of 1,000 (`ENGINE_BUSY` when full). Fills are settled in one transaction: trades, order updates, `TRADE_FILL` journals (seller credited to `earned_pending`, fees), and hold adjustments. On a failed write it reloads the book and rejects the order.

**Outcome:** Orders match and settle atomically.

### T-080 · Engine boot and recovery
**Layers:** BE  
**Status:** Todo

**Description:** Rebuild the books from the database at startup in `seq` order and re-queue orders that were accepted but not processed. A crash test kills the server mid-stream and checks the results.

**Outcome:** A restart loses no orders, fills or balances, meeting that acceptance criterion.

### T-081 · Open orders and cancel
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `OpenOrders` with cancel and cancel-all, and `OrderHistory` with status badges, updating live from `user` channel events.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `GET /orders`, `/orders/{id}`, `DELETE /orders/{id}` and `DELETE /orders?market=`. Cancel releases the remaining hold and works even when the market is halted. Order and fill events on the `user` channel.

**Outcome:** A player can track and pull their orders in real time.

### T-082 · My trades
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `MyTrades` on the trading page and a Trades tab on `/history`.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `GET /me/trades` with fees, side and maker/taker role.

**Outcome:** A player can see every fill they took part in.

### T-083 · Public order book and recent trades
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `OrderBook` with depth bars (clicking a price fills the form) and `RecentTrades`, fed by a mock stream of book diffs and trades.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `GET /markets/{symbol}/book` (top 20 levels) and `/trades` (last 50). WS channels `book:{symbol}` (snapshot plus diffs) and `trades:{symbol}`.

**Outcome:** The live book and trade feed update within 1 second of a trade.

### T-084 · Candles
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `CandleChart` using lightweight-charts with an interval switch (1m to 1d) and live updates of the current candle from mock data.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** Current candles built in Redis from trades, a `candle_closer` job that stores closed candles in Postgres, `GET /candles`, and the `candles:{symbol}:{interval}` channel.

**Outcome:** Live candlestick charts for every market.

### T-085 · Tickers and TickerStrip
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A scrolling `TickerStrip` in the layout with price and 24h change (arrow + color), and live values in `MarketsTable`.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** 24h statistics (last price, change %, high, low, volume) cached in Redis, and a `tickers` channel pushing every second.

**Outcome:** Prices move across the whole app.

### T-086 · Circuit breaker and halts
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A halted banner on the trading page with a countdown to reopening, and the order form disabled (cancel still enabled).

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** An automatic 5-minute halt when the price moves more than 25% in 5 minutes, `Halt` and `Resume` engine commands, and `MARKET_HALTED` rejection.

**Outcome:** Markets are protected from runaway prices.

### T-087 · order_expiry job
**Layers:** BE  
**Status:** Todo

**Description:** Cancels GTC orders older than 30 days and notifies the player.

**Outcome:** The book doesn't collect stale orders.

### T-088 · Exchange links in the inventory and item page
**Layers:** UI  
**Status:** Todo

**UI:** A "Sell on exchange" action in the inventory, and a mini price chart and "View market" link on the item detail page.

⏸ **Checkpoint:** you verify the UI.

**Outcome:** Trading can be reached from wherever a player sees an item.

### T-089 · Engine load test
**Layers:** BE + INFRA  
**Status:** Todo

**Description:** A load script (k6 or Go) at 500 orders per second on one market, with profiling and fixes.

**Outcome:** The engine meets the performance target.

---

## Milestone 9 — Auction house

### T-090 · Create and cancel a player auction
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/auctions/new` with an inventory item picker, quantity, start price, reserve, buy-now, duration, fee preview and confirmation.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `auctions` and `bids` tables. `POST /auctions` validates the lot, holds the items, and enforces the 10-live-auction limit. `DELETE /auctions/{id}` only works with no bids.

**Outcome:** A player can list items for auction.

### T-091 · Auction list and detail pages
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/auctions` tabs (Live, Ending soon, Ended, My auctions) with `AuctionCard`, and `/auctions/[id]` with the item panel, current bid, reserve met or not met, buy-now price and seller controls.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `GET /auctions` with filters, `GET /auctions/{id}` and `GET /auctions/{id}/bids`.

**Outcome:** Players can browse and inspect auctions.

### T-092 · Bidding
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `BidForm` with the minimum bid hint, amount and confirmation, plus the `BID_TOO_LOW` and outbid states.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `POST /auctions/{id}/bids` locks the auction row, applies the minimum increment, holds the bid (only the difference when a player raises their own bid), releases the outbid player's hold and notifies them, and ends the auction immediately on buy-now. Bid rate limit of 5 per second.

**Outcome:** Concurrent bids are handled correctly, and only one can win.

### T-093 · Live auction updates and anti-sniping
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A live `BidHistory` list and a countdown that visibly extends when a late bid arrives, driven by mock events.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** A bid in the last 60 seconds extends the end to bid time + 60 seconds, capped at 30 minutes past the original end. An `auction:{id}` WS channel for new bids, new end times and settlement.

**Outcome:** Auction pages update in real time, and last-second sniping doesn't work.

### T-094 · auction_settler job
**Layers:** BE  
**Status:** Todo

**Description:** A job every second that settles ended auctions in one transaction. `SETTLED` captures the winner's bid, pays the seller minus 5% into `earned_pending`, and moves the items. `UNSOLD` releases everything. Notifies both sides and records the auction price.

**Outcome:** Auctions end correctly, including unsold and buy-now cases.

### T-095 · System legendary drops
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A System badge and legendary styling on `AuctionCard` and the auction detail page.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** A drop table, an `auction_dropper` job (3 drops every 2 hours, 30 minutes each), the legendary minted at settlement with its max supply enforced, and unsold drops returned to the table.

**Outcome:** Legendary items enter the economy through auctions.

### T-096 · My bids and auction history
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A My bids tab, leading bids on the Hangar, an Auctions tab on `/history`, and last auction prices on the item page.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** A "my bids" filter on `GET /auctions` and an auction price history per item.

**Outcome:** A player can track their auction activity.

---

## Milestone 10 — On-chain items, withdrawals and deposits

### T-097 · NebulaItems contract
**Layers:** SC  
**Status:** Todo

**Description:** ERC1155 + ERC1155Supply + ERC2981 + AccessControl + Pausable. `setVault` (callable once), `mintBatch` only to the Vault, `burnBatch` with the burner role, max supply that can be set once, URI, and 5% royalty. Full tests.

**Outcome:** The item token contract, which can only be minted into the Vault.

### T-098 · NebulaVault v2: items
**Layers:** SC  
**Status:** Todo

**Description:** ERC1155Holder receiver hooks (only NebulaItems accepted, `ItemsDeposited` emitted when the transfer isn't a mint), `withdrawItems(ref, …)` and `burnBatch` for supply sync. Replay tests.

**Outcome:** The Vault can hold, release and burn items safely.

### T-099 · NebulaVault v2: USDC withdrawals and admin
**Layers:** SC  
**Status:** Todo

**Description:** `withdrawUSDC(ref, …)` with a rolling 24-hour cap, `setDailyUSDCCap`, `adminTransferUSDC` and a `PAUSER_ROLE`. Fuzz tests.

**Outcome:** USDC can leave the Vault only within the cap.

### T-100 · Invariant tests and deploy v2
**Layers:** SC + INFRA  
**Status:** Todo

**Description:** A Foundry invariant test (`totalSupply == vault + withdrawn holders`). The deploy script adds NebulaItems, wires the roles, sets max supplies, and hands admin to the Safe multisig. Redeploy to Base Sepolia and verify on Basescan.

**Outcome:** All contract acceptance criteria are met.

### T-101 · Relayer
**Layers:** BE  
**Status:** Todo

**Description:** A `chain_txs` queue with a single sender, a DB-managed nonce, receipt tracking, a 20% gas bump after 3 minutes, and pausing plus an alert when out of gas. Tests against Anvil.

**Outcome:** The backend sends on-chain transactions reliably.

### T-102 · supply_sync job
**Layers:** BE  
**Status:** Todo

**Description:** Every 5 minutes, sums unsynced mint and burn entries per item, queues `mintBatch` and `burnBatch`, and marks the entries synced once confirmed (through the indexer's `TransferBatch`).

**Outcome:** On-chain supply follows the game's supply.

### T-103 · Supply stats and supply_reconcile
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A supply stats section on the item page (total, in Vault, withdrawn).

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** A supply stats endpoint, and an hourly check that on-chain balance + unsynced delta = sum of player balances, which pauses withdrawals for an item on a mismatch.

**Outcome:** Ownership can be verified, and players can see supply figures.

### T-104 · Item withdrawal
**Layers:** UI → BE + SC  
**Status:** Todo

**UI (step 1):** `WithdrawPanel` items tab: item and quantity picker, fee, preconditions checklist (wallet, 2FA, 48-hour rule), `TwoFactorPrompt`, and status tracking with `TxStatus`.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend + Solidity (step 2):** `withdrawals` table. `POST /withdrawals` checks all preconditions, holds the items and the 0.50 NC fee, and applies review rules. A worker runs an early supply sync if needed, then `withdrawItems` through the relayer. `SUBMITTED`, then `COMPLETED` on the `ItemsWithdrawn` event. Up to 3 retries, then `FAILED` with holds released. Notifications.

**Outcome:** Items reach the player's wallet as ERC-1155 tokens.

### T-105 · USDC withdrawal
**Layers:** UI → BE + SC  
**Status:** Todo

**UI (step 1):** `WithdrawPanel` NC tab: withdrawable amount (card NC shown as not withdrawable), minimum, fee, daily limit, and the `PENDING_LIQUIDITY` state.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend + Solidity (step 2):** Withdrawal from withdrawable buckets only (minimum 5, fee 0.50, maximum 1,000 per day), `PENDING_LIQUIDITY` with an admin alert, and `withdrawUSDC` through the relayer.

**Outcome:** A player can cash out crypto and earned NC. Card NC can never be withdrawn.

### T-106 · Balance breakdown and earned_settle
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `BalanceBreakdown` on `/wallet` with each bucket and the release dates of pending earnings.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** A job every 10 minutes that moves `earned_pending` older than 72 hours to `earned` (`EARNED_SETTLE`), and release dates in `GET /me/balances`.

**Outcome:** Players can see when trading income becomes withdrawable.

### T-107 · Item deposits
**Layers:** UI → BE + SC  
**Status:** Todo

**UI (step 1):** `DepositItemsPanel` listing the tokens in the connected wallet (`balanceOfBatch`), selection, and a deposit button with `TxStatus`.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend + Solidity (step 2):** Wire the panel to `safeBatchTransferFrom` to the Vault. An `ItemsDeposited` handler credits the linked account, or `unclaimed` when the address isn't linked. A job credits unclaimed items when the address is later linked.

**Outcome:** A player can bring tokens back into the game.

### T-108 · Withdrawal and deposit lists
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `WithdrawalList` and `DepositList` on `/wallet` with status and `TxStatus`.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `GET /withdrawals`, `/withdrawals/{id}` and `/deposits`.

**Outcome:** A player can track every on-chain movement.

### T-109 · usdc_reconcile job
**Layers:** BE  
**Status:** Todo

**Description:** Daily comparison of the Vault's USDC against withdrawable NC, with an alert below 120%.

**Outcome:** Admins are warned before the reserve runs low.

---

## Milestone 11 — Admin

### T-110 · Admin access and audit log
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** The `/admin` layout with side navigation and a required-reason field pattern for every admin write.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** An `is_admin` flag with mandatory 2FA, and `admin_audit_log` middleware that records before and after values plus the reason for every admin write.

**Outcome:** A secure admin area where every action is audited.

### T-111 · User management and freeze
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** User search, a user detail view with their ledger, and freeze, unfreeze and ban actions with confirmation.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** Admin user endpoints. Freezing cancels open orders, voids live auctions, and removes the user's leading bids using the walk-down rule.

**Outcome:** Admins can stop a bad actor in one action.

### T-112 · Withdrawal review queue
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A queue of `PENDING_REVIEW` withdrawals with context (history, value, risk flags) and approve and reject actions.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `/admin/withdrawals` list, approve and reject.

**Outcome:** First-time and large withdrawals are reviewed by a person.

### T-113 · Market admin
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A markets list with halt and reopen, a settings editor (tick size, limits), and the optional system liquidity config (floor bid, ceiling ask, daily budget).

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `/admin/markets` endpoints wired to the engine commands, and the system liquidity feature.

**Outcome:** Admins control the markets.

### T-114 · Config editors
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** Editors for items, recipes, upgrades, zones, loot, shop prices, fees and limits, with validation.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `/admin/config/*` CRUD with validation and cache invalidation.

**Outcome:** Admins can tune the economy without a deploy.

### T-115 · Refunds and disputes
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A payments view with a refund action (only unspent card NC within 14 days) and a dispute review screen.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** Stripe refund plus a `REFUND` journal. The `charge.dispute.created` webhook freezes the account and posts a `DISPUTE_DEBIT` (which may go negative).

**Outcome:** Chargebacks are contained automatically.

### T-116 · Admin adjustments with two approvals
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** An adjustment request form and a pending-approvals list for the second admin.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** An `ADMIN_ADJUSTMENT` request that only posts after a different admin approves it.

**Outcome:** No single admin can create or move value alone.

### T-117 · Drop table management
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A drop table editor and a schedule view.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `/admin/drops` CRUD.

**Outcome:** Admins control legendary releases.

### T-118 · Reconciliation and economy dashboard
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A dashboard with reconciliation results and economy charts: minted vs burned per item per day, NC by bucket, fees earned, and the 20% week-over-week supply alert.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** `/admin/reconciliation` and economy stats endpoints, plus the daily `stripe_reconcile` job.

**Outcome:** Admins can see the health of the economy at a glance.

### T-119 · Audit log viewer
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** A searchable, read-only audit log table with before and after diffs.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** An audit log list endpoint with filters.

**Outcome:** Every admin action can be traced.

---

## Milestone 12 — Hardening and launch

### T-120 · Risk flags and anti-abuse
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** Risk flags on the admin user page and a wash-trading report view.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** Multi-account signals (IP, device fingerprint, deposit address) create `risk_flags`, and a wash-trading report query. No automatic bans.

**Outcome:** Suspicious behavior is surfaced for review.

### T-121 · Global rate limits
**Layers:** BE  
**Status:** Todo

**Description:** A general API limit of 100 per minute per user, and a review of all the other limits.

**Outcome:** The API is protected from abuse.

### T-122 · Landing page
**Layers:** UI → BE  
**Status:** Todo

**UI (step 1):** `/` with the hero, how-it-works steps, the public ticker, top movers, featured auctions and CTAs.

⏸ **Checkpoint:** you verify the UI. Step 2 starts after your "go".

**Backend (step 2):** Public, cached endpoints for top movers and featured auctions.

**Outcome:** Guests understand the game and can sign up.

### T-123 · Terms and Privacy pages
**Layers:** UI  
**Status:** Todo

**UI:** Static legal pages, linked from signup and the footer.

⏸ **Checkpoint:** you verify the pages.

**Outcome:** The legal pages needed for launch exist.

### T-124 · Responsive and accessibility pass
**Layers:** UI  
**Status:** Todo

**UI:** Checks every page at 375 px (the trading page stacks vertically), confirms every color signal also has an arrow or sign, and fixes keyboard navigation and contrast.

⏸ **Checkpoint:** you verify on a phone.

**Outcome:** The app is usable on phones and by colorblind players.

### T-125 · End-to-end tests
**Layers:** INFRA  
**Status:** Todo

**Description:** Playwright tests for each MVP acceptance criterion, run against docker-compose + Anvil + the Stripe test webhook.

**Outcome:** The acceptance criteria are checked automatically.

### T-126 · Security review
**Layers:** BE + SC + UI  
**Status:** Todo

**Description:** A review of auth, idempotency, ledger locking, webhook and SIWE checks, contract roles and caps, secret handling and logging. Fix everything found.

**Outcome:** Known vulnerabilities are closed before a public testnet launch.

### T-127 · Staging deployment
**Layers:** INFRA  
**Status:** Todo

**Description:** Host the Go binary, Postgres, Redis and the Next.js app (provider still to be decided). Secrets, migrations on deploy, monitoring and alerts, and a relayer funded on Base Sepolia.

**Outcome:** A public testnet build that players can try.
