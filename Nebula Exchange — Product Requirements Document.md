# Nebula Exchange — Product Requirements Document

Sep 29, 2026 · @Md Nazmul Hasan

## 1. Overview

Nebula Exchange is a browser-based space trading game where players mine resources on timed missions, craft and upgrade gear, and trade everything with other players on a live order-book exchange and an auction house. Every item in the game is backed 1:1 by an ERC-1155 token held in an on-chain vault, and players can withdraw items to their own wallet at any time.

**Product goals**

1. A playable game loop (missions → resources → crafting → upgrades) that creates real supply and demand.
2. A real trading experience: order book, live prices, candlestick charts, 24h price change, and timed auctions.
3. Two payment rails, Stripe (card) and crypto (USDC), that converge into one internal balance.
4. Verifiable ownership: total item supply on-chain always equals total item balances in the game.

**Non-goals for MVP**

- Real-time action gameplay (combat, movement). Gameplay is timer and menu based.
- Mainnet launch. MVP runs on Base Sepolia testnet with Stripe test mode.
- Mobile apps. The web app must be responsive, but no native apps.
- Player-to-player chat, guilds, or direct gifting.

**Tech stack**

| Layer | Choice |
| --- | --- |
| Frontend | Next.js (App Router), TypeScript, Tailwind CSS, wagmi + viem for wallet, lightweight-charts for candles |
| Backend | Go monolith with Gin, single deployable binary |
| Database | PostgreSQL (source of truth), Redis (rate limits, price cache, sessions) |
| Realtime | WebSocket hub inside the monolith |
| Smart contracts | Solidity, Foundry, OpenZeppelin |
| Chain | Base Sepolia (MVP), Base mainnet later |
| Payments | Stripe Checkout + webhooks; USDC on Base |
| Email | Any transactional provider (Resend, Postmark, SES) |

**Glossary**

| Term | Meaning |
| --- | --- |
| NC (Nebula Credits) | The single in-game currency. 1 NC = 1 USD. Stored as int64 with 6 decimals (1 NC = 1,000,000 units), same as USDC. |
| Item | Anything a player owns except NC: resources, components, tools, ships, legendary items. Each has an `item_id` equal to its ERC-1155 token ID. |
| Vault | The smart contract that holds all in-game items and USDC on behalf of players. |
| In-game balance | What the database says a player owns. Always backed by tokens in the Vault. |
| Withdraw | Move NC (as USDC) or items from the Vault to the player's own wallet. |
| Deposit | Move USDC or items from the player's wallet into the Vault. |
| Hold | Funds or items reserved for an open order or bid. Not spendable until released. |
| Market | A trading pair on the exchange, e.g. `IRON/NC`. |
| Sink | Anything that removes NC or items from the economy (fees, crafting, fuel). |

## 2. Roles and player permissions

There are four roles. Permissions are enforced in backend middleware, never only in the UI.

| Role | Who | Access |
| --- | --- | --- |
| Guest | Not logged in | Landing page, public market prices, public auction list (read-only) |
| Pending player | Signed up, email verified, entry fee not paid | Everything a guest sees + pay entry fee + profile settings |
| Player | Entry fee paid, account `ACTIVE` | Full game, shop, exchange, auctions, wallet |
| Admin | Internal staff | Admin panel (section 13) |

**A player CAN**

- Buy items and upgrades from the shop, paying from NC balance, by card, or with USDC.
- Top up NC by card (Stripe) or by depositing USDC.
- Start missions, collect loot, craft components, and upgrade tools and ships.
- Place limit and market orders on any exchange market, and cancel open orders.
- Bid on auctions, and create auctions for items they own.
- Connect one wallet, withdraw withdrawable NC as USDC, withdraw items as ERC-1155 tokens, and deposit tokens back.
- See full history of payments, trades, bids, missions, and ledger movements.

**A player CANNOT**

- Trade before paying the entry fee or before verifying email.
- Withdraw NC that came from card top-ups (card NC is spend-only; see section 5).
- Withdraw anything without a connected wallet and 2FA enabled.
- Spend NC or items that are on hold for open orders, bids, or pending withdrawals.
- Bid on their own auction, match against their own exchange orders, or cancel an auction after the first bid.
- Send items or NC directly to another player. All transfers between players go through the exchange or auction house.
- Connect more than one wallet at a time, or connect a wallet already linked to another account.
- Have more than 50 open exchange orders, 10 live auctions, or 3 missions running at once.
- Edit a completed trade, auction result, or mission result.

## 3. Signup, login, and account flow

Signup is email + password. Wallet connection is optional at signup and required only for crypto payments, deposits, and withdrawals.

**Signup steps**

1. Player enters email, username (3–20 chars, letters/digits/underscore, unique, case-insensitive), password (min 10 chars), and ticks "I am 18+ and accept the Terms".
2. Backend creates the user with status `UNVERIFIED`, hashes the password with argon2id, and emails a 6-digit code (valid 15 minutes, max 5 attempts, resend allowed after 60 seconds).
3. Player enters the code. Status becomes `PENDING_PAYMENT`. Player is logged in.
4. Player pays the entry fee (5 NC = $5) by card or USDC (section 5). On confirmation, status becomes `ACTIVE`.
5. The backend grants the starter pack in the same database transaction: 1 Scout ship (tier 1), 1 Basic Drill (tier 1), 10 Fuel Cells, and 2 NC bonus (card-origin, spend-only).
6. Player lands on the Hangar dashboard with an onboarding checklist.

**Login and sessions**

- Login with email + password. Returns a short-lived access JWT (15 minutes) in memory and a refresh token (30 days) in an httpOnly, Secure, SameSite=Strict cookie.
- Refresh tokens are stored hashed in the database and rotated on every use. Reusing an old refresh token revokes all sessions for that user.
- 5 failed logins in 15 minutes locks login for 15 minutes for that email.
- Forgot password: emailed reset link, valid 30 minutes, single use. Resetting revokes all sessions.
- 2FA (TOTP, e.g. Google Authenticator) is optional for login and mandatory before any withdrawal or wallet change.

**Connecting a wallet (later, from the Wallet page)**

We use Sign-In With Ethereum (EIP-4361) to prove the player owns the address.

1. Player clicks "Connect wallet" and chooses MetaMask, Coinbase Wallet, or WalletConnect.
2. Frontend calls `POST /wallet/nonce`. Backend returns a random nonce valid for 5 minutes.
3. Frontend builds a SIWE message (domain, address, chain ID, nonce, issued-at) and asks the wallet to sign it.
4. Frontend sends the message + signature to `POST /wallet/link`. Backend verifies the signature, nonce, domain, and chain ID, then checks the address is not linked to any other account.
5. Address is stored lowercase with `linked_at`. An email confirms the link.

Changing a wallet requires 2FA, sends an email, and blocks withdrawals for 48 hours after the change. Unlinking is allowed only when there are no pending withdrawals.

**Account states**

| State | Meaning | Can do |
| --- | --- | --- |
| `UNVERIFIED` | Signed up, email not confirmed | Enter code, resend code |
| `PENDING_PAYMENT` | Email confirmed, entry fee unpaid | Pay entry fee, view markets |
| `ACTIVE` | Entry fee paid | Everything |
| `FROZEN` | Admin or automated risk freeze (e.g. chargeback) | Log in and view only; no trading, withdrawals, or payments |
| `BANNED` | Permanent | Nothing; open orders cancelled, holds released |

Unverified accounts older than 7 days are deleted by a daily job.

## 4. Game economy and gameplay

The economy has one currency (NC) and fungible items identified by `item_id`. All game rules run on the server; the browser only displays state and sends actions. All numbers below are starting values stored in config tables so admins can tune them without a deploy.

**Item catalog and ID ranges**

The `item_id` is the ERC-1155 token ID. Every item is fungible within its ID. Upgrades convert one ID to another (burn old, mint new), so no item has per-instance state.

| ID range | Category | Items (MVP) | Tradeable on |
| --- | --- | --- | --- |
| 1–99 | Resources | 1 Iron Ore, 2 Copper Ore, 3 Crystal, 4 Plasma, 5 Rare Earth, 6 Void Shard | Exchange, auctions |
| 100–199 | Components | 101 Alloy Plate, 102 Circuit, 103 Power Core, 104 Void Engine | Exchange, auctions |
| 200–299 | Drills | 201–205 Drill T1–T5 | Exchange, auctions |
| 300–399 | Ships | 301 Scout, 302 Hauler, 303 Freighter, 304 Interceptor | Exchange, auctions |
| 400–499 | Consumables | 401 Fuel Cell | Exchange, auctions |
| 10000+ | Legendary | Unique items, max supply 1 each (e.g. 10001 "Relic Drill of Orion") | Auctions only |

**Missions (the main supply source)**

A mission sends one ship with one drill to a zone. The player chooses ship, drill, and zone, and pays fuel up front. The ship and drill are locked (held) until the mission is collected.

| Zone | Duration | Fuel | Loot (base units per run) | Requirement |
| --- | --- | --- | --- | --- |
| Asteroid Belt | 15 min | 2 | Iron 8–14, Copper 3–6 | None |
| Crystal Moon | 30 min | 4 | Crystal 4–8, Copper 4–8 | Drill T2+ |
| Plasma Nebula | 60 min | 6 | Plasma 3–6, Crystal 2–5 | Drill T3+, Hauler+ |
| Deep Void | 4 h | 15 | Rare Earth 2–4, 5% chance of 1 Void Shard | Drill T4+, Freighter or Interceptor |

Loot rules:

1. When the timer ends, the server rolls each resource range uniformly using Go `crypto/rand`, then multiplies by the drill multiplier (T1 1.0, T2 1.25, T3 1.5, T4 2.0, T5 2.5), rounding down.
2. Total units are capped by ship cargo (Scout 20, Hauler 60, Freighter 150, Interceptor 40). If over cap, resources are reduced proportionally, rarest resource last.
3. Ship speed changes duration: Interceptor ×0.6, Freighter ×1.2, others ×1.0.
4. The roll result is saved on the mission row when the timer ends (by a background job), not when the player clicks, so refreshing cannot re-roll.
5. Player clicks "Collect" to move loot into inventory and unlock the ship and drill. Uncollected loot stays forever.
6. Max 3 running missions per player. A mission can be aborted before it ends: fuel is not refunded and no loot is given.

**Crafting (converts resources into components)**

| Output | Inputs | Craft time | NC fee |
| --- | --- | --- | --- |
| 1 Alloy Plate | 5 Iron Ore + 2 Copper Ore | 1 min | 0.02 |
| 1 Circuit | 3 Copper Ore + 1 Crystal | 2 min | 0.05 |
| 1 Power Core | 2 Crystal + 2 Plasma | 5 min | 0.10 |
| 1 Void Engine | 1 Void Shard + 2 Power Core + 1 Rare Earth | 30 min | 1.00 |

Crafting takes inputs and fee immediately (burned), queues the job, and delivers the output when the timer ends. Batch crafting (quantity 1–100) is allowed; time multiplies by quantity. Max 1 active craft queue per player.

**Upgrades (drills and ships)**

Every upgrade has two paths. Both burn the current item and mint the next tier.

| Upgrade | Craft path (materials + NC fee) | Buy path (shop price) |
| --- | --- | --- |
| Drill T1 → T2 | 5 Alloy Plate + 2 Circuit + 0.50 NC | 3.00 NC |
| Drill T2 → T3 | 10 Alloy Plate + 5 Circuit + 1 Power Core + 1.00 NC | 7.00 NC |
| Drill T3 → T4 | 20 Alloy Plate + 10 Circuit + 5 Power Core + 2.50 NC | 15.00 NC |
| Drill T4 → T5 | 2 Void Engine + 10 Power Core + 5.00 NC | Not sold |
| Scout → Hauler | 15 Alloy Plate + 5 Circuit + 1.00 NC | 6.00 NC |
| Hauler → Freighter | 40 Alloy Plate + 10 Power Core + 3.00 NC | 18.00 NC |
| Scout → Interceptor | 10 Circuit + 5 Power Core + 2.00 NC | 12.00 NC |

Buy-path upgrades are the paid "upgrade" feature and can be paid from NC, by card, or with USDC (section 5).

**Shop (fixed-price sales by the system)**

The shop sells: Fuel Cell 0.20 NC, Scout 2.00 NC, Drill T1 1.00 NC, Starter Bundle (1 Scout + 1 Drill T1 + 20 Fuel) 3.50 NC, and buy-path upgrades above. Shop items are minted new on purchase. The shop never buys items back.

**Sources and sinks**

| Sources (create value) | Sinks (remove value) |
| --- | --- |
| Mission loot (resources) | Fuel burned on missions |
| Shop purchases (new items minted) | Crafting inputs and fees burned |
| Starter pack | Upgrade inputs burned |
| System auction drops (legendaries) | Exchange fees (0.25% maker, 0.50% taker) |
| NC top-ups (card, USDC) | Auction fee (5% of final price) |
|  | Withdrawal fee (flat 0.50 NC) |

An admin dashboard tracks daily minted vs burned units per item. If an item's supply grows more than 20% week over week, admins can lower loot or raise crafting demand.

## 5. Payments

Every payment, card or crypto, ends the same way: it credits NC to the player's ledger, then the purpose (entry fee, top-up, or shop purchase) is executed as an NC debit in the same database transaction. After that point, the rest of the system only knows about NC.

&#91;embedded content: payment convergence · 2 rails, 1 ledger\]

Stripe and the chain indexer are the only two entry points for money; both credit the ledger and then run the same purpose code.

**Where each payment method is allowed**

| Use case | Card (Stripe) | Crypto (USDC) | From NC balance |
| --- | --- | --- | --- |
| Entry fee (5 NC) | Yes | Yes | No (balance is empty before activation) |
| NC top-up | Yes | Yes | — |
| Shop items and buy-path upgrades | Yes | Yes | Yes |
| Crafting fees, craft-path upgrade fees | No | No | Yes |
| Exchange orders | No | No | Yes |
| Auction bids | No | No | Yes |
| Withdraw NC | No | Yes (as USDC to linked wallet) | — |
| Withdraw or deposit items | No | Yes (ERC-1155 to/from linked wallet) | — |

The exchange and auctions use NC only because trades must settle instantly, holds must lock funds before matching, and per-trade card or gas fees would make small trades impossible.

**Three NC buckets**

Each player's NC is split by origin. The UI shows one total plus a breakdown.

| Bucket | Comes from | Spendable | Withdrawable |
| --- | --- | --- | --- |
| `card` | Stripe top-ups and card purchases, starter bonus | Yes | Never |
| `crypto` | USDC payments and deposits | Yes | Yes |
| `earned` | Proceeds from exchange sales and auctions | Yes | Yes, 72 hours after the sale |

Spending order is always `card` → `earned` → `crypto`, so the player keeps as much withdrawable NC as possible. Card NC is never withdrawable because a card payment can be charged back for up to 120 days, and paying it out as USDC would make the platform a money-transfer service. The 72-hour delay on `earned` NC gives the risk system time to catch chargebacks from buyers.

**Card payment flow (Stripe Checkout)**

1. Frontend calls `POST /payments` with `{purpose, method: "card", amount_nc | sku}` and an `Idempotency-Key` header.
2. Backend validates (account state, limits, SKU price), inserts a `payments` row with status `PENDING`, and creates a Stripe Checkout Session with `metadata.payment_id`. Returns the Checkout URL.
3. Player pays on Stripe's hosted page and is redirected to `/payment/result?id=…`, which polls `GET /payments/{id}` every 2 seconds for up to 60 seconds.
4. Stripe sends `checkout.session.completed` to `POST /webhooks/stripe`. Backend verifies the signature, stores the event ID (duplicate events are ignored), and in one transaction: marks the payment `SUCCEEDED`, credits NC to `card`, and runs the purpose.
5. If the purpose fails (e.g. the account is frozen), the NC stays in the balance and the player is notified.

Card limits: top-ups of 5, 10, 25, 50, or 100 NC; max 500 NC per day and 2,000 NC per 30 days per player. Stripe fees are absorbed by the platform.

**Crypto payment flow (USDC on Base)**

Requires a linked wallet.

1. Frontend calls `POST /payments` with `method: "crypto"`. Backend creates the row with a `payment_ref` (bytes32 = keccak256 of the payment UUID), the exact USDC amount, and a 30-minute expiry.
2. Frontend asks the wallet for two transactions: `USDC.approve(Vault, amount)` (skipped if allowance is enough), then `Vault.pay(paymentRef, amount)`. Gas is paid by the player (a few cents on Base).
3. The chain indexer sees `PaymentReceived(paymentRef, payer, amount)`, waits 5 confirmations, and in one transaction marks the payment `SUCCEEDED`, credits NC to `crypto`, and runs the purpose.
4. Frontend shows "Confirming (n/5)" using the tx hash, then success.

Crypto edge cases:

- Amount less than expected → credit it as a top-up, do not run the purpose, notify the player.
- Amount more than expected → run the purpose; the extra stays as `crypto` NC.
- Paid after expiry → credit as a top-up.
- Payer address differs from the linked wallet → still credit the account that owns the `payment_ref`, and flag for review.
- USDC sent straight to the Vault address without `pay()` cannot be matched automatically. The UI warns against this. Recovery is a manual admin action after the player proves the sending address.

**Refunds and chargebacks**

- Unspent card NC can be refunded within 14 days through support. Admin triggers a Stripe refund and debits the `card` bucket.
- Spent NC is not refundable. Crypto payments are not refundable; players can withdraw instead.
- On `charge.dispute.created`, the account is set to `FROZEN`, open orders are cancelled, and the disputed amount is debited from `card`. If the bucket is short, it goes negative and future income repays it first. An admin reviews every dispute.

**Payment statuses**

`PENDING` → `SUCCEEDED` | `FAILED` | `EXPIRED` | `REFUNDED` | `DISPUTED`. Pending card payments expire after 24 hours (Stripe session expiry); pending crypto payments after 30 minutes.

## 6. Custody, deposits, and withdrawals

Gameplay and trading run off-chain on the database ledger, while every item is backed by a real ERC-1155 token held in the Vault. This gives instant trades and gas-free gameplay while keeping supply verifiable on-chain.

**How tokens are generated (supply sync)**

1. Whenever an item is created in the game (mission loot, crafting output, shop purchase, upgrade, starter pack, legendary drop), the ledger records a mint entry immediately. The player sees it in inventory at once.
2. Whenever an item is destroyed (fuel, crafting inputs, upgrade inputs), the ledger records a burn entry.
3. A `supply_sync` job runs every 5 minutes. It sums unsynced mint and burn entries per `item_id`, then the relayer calls `NebulaItems.mintBatch(Vault, ids, amounts)` and `Vault.burnBatch(ids, amounts)`. Entries are marked synced with the tx hash once confirmed.
4. Invariant, checked hourly: `Vault balance on-chain + unsynced delta = sum of all player balances` for every `item_id`. A mismatch pages an admin and pauses withdrawals for that item.

**Withdrawing items to a wallet**

Preconditions: account `ACTIVE`, linked wallet, 2FA enabled, no wallet change in the last 48 hours, enough available (unheld) quantity, and 0.50 NC for the fee.

1. Player picks item and quantity on the Wallet page and confirms with a TOTP code.
2. Backend creates a `withdrawals` row (`REQUESTED`), holds the items and fee.
3. If rules require review (first-ever withdrawal, or value over 250 NC at last trade price), status is `PENDING_REVIEW` until an admin approves or rejects.
4. A `withdrawal_worker` ensures the item is synced on-chain (runs an early supply sync if needed), then the relayer calls `Vault.withdrawItems(withdrawalRef, to, ids, amounts)`. Status `SUBMITTED` with tx hash.
5. After 5 confirmations, status `COMPLETED`: held items move to the `withdrawn` system account and the fee to `fees`.
6. On failure, the worker retries up to 3 times with backoff. After that, status `FAILED`, holds are released, fee refunded, admin alerted.

Once withdrawn, the token is fully the player's. They can hold it, sell it on external marketplaces (5% ERC-2981 royalty to the platform), or deposit it back.

**Withdrawing NC as USDC**

Same flow as items, using `Vault.withdrawUSDC(withdrawalRef, to, amount)`. Only `crypto` and settled `earned` NC can be withdrawn. Minimum 5 NC, fee 0.50 NC, max 1,000 NC per day. The Vault must hold enough USDC; if not, the withdrawal waits in `PENDING_LIQUIDITY` and admins are alerted (card revenue sits in Stripe, not in the Vault, so admins top up the Vault's USDC reserve).

**Depositing items from a wallet**

1. Player clicks "Deposit", selects items held in their wallet (read via the frontend with viem `balanceOfBatch`).
2. Wallet sends `NebulaItems.safeBatchTransferFrom(player, Vault, ids, amounts, "")`.
3. The Vault's `onERC1155BatchReceived` accepts only NebulaItems tokens and emits `ItemsDeposited(from, ids, amounts)`.
4. Indexer waits 5 confirmations, finds the account by linked wallet address, and credits the items.
5. If the sender address is not linked to any account, items go to an `unclaimed` account. When that address is later linked, a job credits them automatically.

**Depositing USDC**

There is no separate USDC deposit. Adding USDC is a crypto top-up (section 5), which always carries a `payment_ref` so it can be matched to an account.

## 7. Exchange (order-book trading)

The exchange is a continuous limit order book per item, priced in NC, with price-time priority matching, like a stock or crypto exchange. It is used for every fungible item; legendary items trade only in auctions.

**Markets**

- One market per non-legendary `item_id`, symbol `<ITEM>/NC`, e.g. `IRON/NC`, `ALLOY/NC`, `DRILL3/NC`, `FUEL/NC`. About 21 markets at launch.
- Each market row stores: `tick_size` (resources and fuel 0.0001 NC; components 0.001 NC; drills and ships 0.01 NC), `min_qty` 1, `max_qty` 1,000,000, `min_notional` 0.01 NC, `status` (`OPEN`, `HALTED`, `CLOSED`).
- Quantities are whole units. Prices are int64 micro-NC and must be a multiple of `tick_size`.

**Order types**

| Type | Player enters | Behavior |
| --- | --- | --- |
| Limit GTC | Side, price, qty | Matches what it can, the rest rests in the book until filled, cancelled, or 30 days pass |
| Limit IOC | Side, price, qty | Matches what it can immediately, cancels the rest |
| Market | Side, qty | Matches immediately against the best prices up to a protection price (best opposite price ±10%); any rest is cancelled |

**Placing an order (validation and holds)**

1. Check: account `ACTIVE`, market `OPEN`, price on tick, qty and notional within limits, under 50 open orders.
2. Hold funds or items, in the same DB transaction as inserting the order:
   - Buy limit: hold `price × qty + taker fee`.
   - Buy market: hold `protection price × qty + taker fee`.
   - Sell: hold `qty` units of the item.
3. If the hold fails (insufficient available balance), reject with `INSUFFICIENT_FUNDS` and insert nothing.
4. Send the order to the market's matching engine (section 11) and return `202 Accepted` with the order ID. The final result arrives over WebSocket and via `GET /orders/{id}`.

**Matching rules**

1. Best price first: highest bid, lowest ask. At the same price, earlier order first (sequence number, not wall clock).
2. A trade happens when the incoming (taker) order's price crosses the best resting (maker) order. The trade price is always the maker's price.
3. The taker keeps matching level by level until it is filled, no longer crosses, or hits its protection price.
4. Self-trade prevention: if the next maker belongs to the same player, the taker's remaining quantity is cancelled (not matched).
5. Every fill is settled in the same DB transaction as the fill record: buyer gets items, seller gets NC into `earned` (withdrawable after 72 hours), fees go to the `fees` system account, and holds shrink accordingly.
6. If a buy fills below its limit price, the unused part of the hold (price improvement and unused fee) is released at once.

**Fees**

Maker 0.25%, taker 0.50% of the trade value, charged in NC to both sides. The buyer pays the fee on top of the price; the seller's fee is taken from proceeds. Fees are rounded up to 1 micro-NC.

**Worked example**

The `IRON/NC` book has asks: 20 @ 0.0100 (Ali, placed first), 30 @ 0.0100 (Sam), 50 @ 0.0110 (Ria). Priya sends a buy limit of 40 @ 0.0105. She fills 20 from Ali and 20 from Sam, both at 0.0100. Sam's order stays open with 10 left. Priya's order is filled, and she gets back the unused hold: 40 × (0.0105 − 0.0100) plus the fee difference. The new last price is 0.0100.

**Cancelling**

The owner can cancel an `OPEN` or `PARTIALLY_FILLED` order at any time, even when the market is halted. Cancel releases the remaining hold. Filled parts are final.

**Order statuses**

`OPEN` → `PARTIALLY_FILLED` → `FILLED`; or → `CANCELLED` (by player, IOC/market remainder, self-trade, 30-day expiry, admin, or account freeze); or `REJECTED` at validation.

**Price data (the "price hike" view)**

- Last price, 24h change % = (last − price 24h ago) ÷ price 24h ago, 24h high, 24h low, 24h volume in units and NC.
- Candles: 1m, 5m, 15m, 1h, 4h, 1d, built from trades and kept in Redis for the current candle and Postgres for closed candles.
- Order book depth aggregated by price level (top 20 each side), and the last 50 trades.
- Markets page ranks top gainers, top losers, and highest volume in 24h.

**Protections**

- Circuit breaker: if the last price moves more than 25% versus 5 minutes earlier, the market is `HALTED` for 5 minutes. New orders are rejected; cancels still work. Admins can also halt or reopen any market.
- Rate limit: 10 order actions per second per player.
- Optional system liquidity (off by default): admins can configure a floor bid and ceiling ask per item from a treasury account with a daily NC budget, to keep new markets from being empty at launch.

## 8. Auction house

Auctions are English (ascending-price) auctions with a fixed end time and anti-sniping extensions. They are the only way to trade legendary items and an optional way to sell any item as a lot.

**Auction sources**

| Source | Items | Schedule | Duration | Proceeds go to |
| --- | --- | --- | --- | --- |
| System drop | Legendary items (minted on settlement) | 3 auctions every 2 hours, from an admin-managed drop table | 30 min | `treasury` system account |
| Player auction | Any item the player owns, as a lot (item + qty) | Any time | 1 h, 6 h, or 24 h | Seller's `earned` bucket, minus 5% fee |

**Creating a player auction**

1. Player picks item, quantity, starting price (min 0.10 NC), optional reserve price (hidden minimum), optional buy-now price (must be ≥ 2 × starting price), and duration.
2. Backend checks: account `ACTIVE`, under 10 live auctions, item available. It holds the items and creates the auction as `LIVE` with `ends_at`.
3. The seller can cancel only while there are no bids. Cancel releases the items.

**Bidding rules**

1. Bidder must be `ACTIVE` and cannot be the seller.
2. Minimum next bid = current highest bid + max(5% of it, 0.10 NC). The first bid must be ≥ starting price.
3. The full bid amount is held from the bidder's NC. If they are already the highest bidder and raise their bid, only the difference is added to the hold.
4. When someone is outbid, their hold is released immediately and they get a notification.
5. Bids are processed one at a time per auction: the bid handler locks the auction row (`SELECT … FOR UPDATE`), validates, updates the highest bid, adjusts holds, and commits. Two bids at the same moment can never both win.
6. Buy-now: any bid ≥ buy-now price ends the auction immediately at that price.
7. Bids cannot be withdrawn.

**Anti-sniping**

If a valid bid arrives in the last 60 seconds, `ends_at` becomes bid time + 60 seconds. Total extension is capped at 30 minutes past the original end time.

**Settlement (runs when `ends_at` passes)**

A scheduler goroutine checks every second for `LIVE` auctions with `ends_at ≤ now` and settles each in one transaction:

- Highest bid ≥ reserve (or no reserve): winner's held NC is captured; seller gets proceeds minus 5% fee into `earned`; items move to the winner (system drops mint the legendary at this point). Status `SETTLED`.
- No bids, or reserve not met: holds released, items return to the seller. Status `UNSOLD`. System drops that go unsold return to the drop table.
- Winner and seller are notified. The final price is recorded as an auction trade and shown on the item page.

**Auction statuses**

`LIVE` → `SETTLED` | `UNSOLD` | `CANCELLED` (seller, before any bid) | `VOIDED` (admin, e.g. fraud; all holds released).

**Freezes during an auction**

If the seller is frozen, their live auctions are voided. If the highest bidder is frozen, their bid is removed and its hold released. The handler then walks down earlier bids, newest first, and makes the first bidder who can still cover their bid the new leader, placing a fresh hold. If nobody can, the auction has no bids.

## 9. Ledger and balances

All NC and item movements go through one double-entry ledger in Postgres. No code may change a balance except through the ledger package. This is the rule that keeps money and items from being created or lost by bugs.

**Core model**

- **Account**: one per (owner, asset, bucket). Owner is a user or a system account. Asset is `NC` or an `item_id`. Bucket applies to NC only (`card`, `crypto`, `earned_pending`, `earned`).
- **Journal**: one business event (a trade fill, a payment, a craft). Has `type`, `ref_type`, `ref_id`, and a unique key on (`type`, `ref_type`, `ref_id`) so the same event can never be posted twice.
- **Entry**: a signed amount on one account inside a journal. For every journal, entries sum to zero per asset.
- **Balance**: `available` and `held` per account, updated in the same transaction as the entries. A CHECK constraint keeps `available ≥ 0` and `held ≥ 0`; only the `DISPUTE_DEBIT` journal type may push `card` below zero.
- **Hold**: a reservation row (`amount`, `ref_type`, `ref_id`, `status`: `ACTIVE`, `CAPTURED`, `RELEASED`) that moves value from `available` to `held`.

**Operations the ledger package exposes**

| Function | Effect |
| --- | --- |
| `Post(journal)` | Writes entries and updates available balances. Fails if any balance goes negative. |
| `Hold(account, amount, ref)` | available − amount, held + amount, creates hold row. |
| `Release(hold, amount)` | held − amount, available + amount. Partial release allowed. |
| `Capture(hold, amount, to)` | Takes from held and posts a journal to the target account(s). Partial capture allowed. |
| `Spend(user, amount, ref)` | Helper that debits NC across buckets in order `card` → `earned``_pending → earned → crypto (not-yet-withdrawable earnings are spent before withdrawable ones)`. |

**System accounts**

`stripe_clearing` (card money in), `crypto_clearing` (USDC in), `fees`, `treasury` (system auction proceeds, liquidity budget), `mint` (source of new items), `burn` (destroyed items), `withdrawn` (items and NC that left to wallets), `unclaimed` (deposits from unlinked addresses). `mint` and `stripe_clearing` go negative by design; they represent value entering the system.

**Journal types**

`ENTRY_FEE`, `TOPUP_CARD`, `TOPUP_CRYPTO`, `SHOP_PURCHASE`, `STARTER_PACK`, `MISSION_FUEL`, `MISSION_LOOT`, `CRAFT_START`, `CRAFT_OUTPUT`, `UPGRADE`, `TRADE_FILL`, `AUCTION_SETTLE`, `WITHDRAWAL`, `DEPOSIT`, `EARNED_SETTLE`, `REFUND`, `DISPUTE_DEBIT`, `ADMIN_ADJUSTMENT` (requires a reason and two admins).

**Concurrency rules**

1. Every ledger call runs inside the caller's DB transaction, so a trade fill and its ledger entries commit or fail together.
2. Balance rows are locked with `SELECT … FOR UPDATE`, always in ascending account ID order, to prevent deadlocks.
3. Amounts are int64 micro-NC for NC and int64 units for items. No floats anywhere in the backend. The frontend formats with a decimal library.

**Idempotency**

- Every state-changing API accepts an `Idempotency-Key` header. The key, request hash, and response are stored for 24 hours; a retry with the same key returns the stored response.
- Webhooks and chain events are stored by their external ID (Stripe event ID, tx hash + log index) before processing.
- Journals are unique per business event, as above.

**Scheduled ledger jobs**

- `earned_settle` (every 10 min): moves `earned_pending` entries older than 72 hours into `earned`.
- `ledger_check` (hourly): verifies every journal sums to zero, every balance equals the sum of its entries, and held equals the sum of active holds. Any mismatch pages an admin.
- `supply_reconcile` (hourly): the on-chain vs off-chain item check from section 6.
- `stripe_reconcile` (daily): Stripe balance transactions vs `stripe_clearing`.
- `usdc_reconcile` (daily): Vault USDC balance vs total withdrawable NC; alert if reserve falls below 120% of withdrawable NC.

## 10. Smart contracts

Three contracts, built with Foundry and OpenZeppelin v5, deployed to Base Sepolia. They are non-upgradeable for MVP; a fix means redeploy and migrate. The contracts only handle ownership and money in/out. All game rules stay in the backend.

**NebulaItems (ERC-1155)**

Inherits `ERC1155`, `ERC1155Supply`, `ERC2981`, `AccessControl`, `Pausable`.

| Function | Access | Behavior |
| --- | --- | --- |
| `mintBatch(ids, amounts)` | `MINTER_ROLE` (relayer) | Mints only to the Vault address (hard-coded at deploy), never to players directly |
| `burnBatch(from, ids, amounts)` | `BURNER_ROLE` (Vault) | Burns tokens the Vault holds |
| `setMaxSupply(id, max)` | `DEFAULT_ADMIN_ROLE` | Set once per ID; legendary IDs get max 1. Mint reverts above max |
| `setURI(newuri)` | `DEFAULT_ADMIN_ROLE` | Base metadata URI, e.g. `https://api.<domain>/metadata/{id}.json` |
| `setDefaultRoyalty(receiver, bps)` | `DEFAULT_ADMIN_ROLE` | 500 bps (5%) to treasury |
| `pause()` / `unpause()` | `DEFAULT_ADMIN_ROLE` | Blocks all transfers in an emergency |

Metadata JSON per ID follows the OpenSea standard: `name`, `description`, `image`, and `attributes` (category, tier, cargo, drill multiplier). The backend serves it; images are static files.

**NebulaVault (custody)**

Holds all in-game items and the USDC reserve. Inherits `AccessControl`, `Pausable`, `ReentrancyGuard`, `ERC1155Holder`.

| Function | Access | Behavior |
| --- | --- | --- |
| `pay(bytes32 paymentRef, uint256 amount)` | Anyone | Pulls USDC from caller with `transferFrom`. Reverts if `paymentRef` was used. Emits `PaymentReceived(paymentRef, payer, amount)` |
| `withdrawItems(bytes32 ref, address to, ids, amounts)` | `OPERATOR_ROLE` | Reverts if `ref` processed. Transfers items to `to`. Emits `ItemsWithdrawn(ref, to, ids, amounts)` |
| `withdrawUSDC(bytes32 ref, address to, uint256 amount)` | `OPERATOR_ROLE` | Reverts if `ref` processed or daily cap exceeded. Emits `USDCWithdrawn(ref, to, amount)` |
| `burnBatch(ids, amounts)` | `OPERATOR_ROLE` | Calls `NebulaItems.burnBatch(address(this), …)` for supply sync |
| `onERC1155Received` / `onERC1155BatchReceived` | Token contract | Reverts unless `msg.sender` is NebulaItems. If `from != address(0)` (not a mint), emits `ItemsDeposited(from, ids, amounts)` |
| `setDailyUSDCCap(amount)` | `DEFAULT_ADMIN_ROLE` | On-chain limit on USDC leaving per 24 h (start: 5,000 USDC), a backstop if the relayer key leaks |
| `adminTransferUSDC(to, amount)` | `DEFAULT_ADMIN_ROLE` | Moves reserve (e.g. to treasury). Admin should be a multisig |
| `pause()` / `unpause()` | `PAUSER_ROLE` | Stops pay, deposits, and withdrawals |

**MockUSDC (testnet only)**

ERC-20 with 6 decimals, ERC-2612 permit, and a public `faucet()` that mints 1,000 test USDC per address per day, so testers can pay without real money.

**Keys and roles**

- `DEFAULT_ADMIN_ROLE`: a Safe multisig (2 of 3), even on testnet, so the process is practiced.
- Relayer: one backend hot wallet holding `MINTER_ROLE` on NebulaItems and `OPERATOR_ROLE` on the Vault. Funded with test ETH for gas. Its key is loaded from an environment secret; production must use a KMS.
- The relayer sends transactions from a single worker with a DB-managed nonce, so two transactions never use the same nonce. Stuck transactions are replaced with a 20% higher gas price after 3 minutes.

**Events the backend indexes**

`PaymentReceived`, `ItemsDeposited`, `ItemsWithdrawn`, `USDCWithdrawn`, and NebulaItems `TransferBatch` / `TransferSingle` (for supply sync confirmation). The indexer polls `eth_getLogs` every 3 seconds from the last processed block, treats a log as final after 5 confirmations, and stores (tx hash, log index) to avoid double processing. On a reorg, unconfirmed logs are simply re-read.

**Required tests (Foundry)**

- Unit tests for every function and every revert path, including role checks.
- Fuzz tests on `pay`, `withdrawItems`, and `withdrawUSDC` amounts and refs.
- Invariant test: for every ID, `totalSupply(id) == balanceOf(Vault, id) + sum of balances of withdrawn holders`.
- Replay test: the same `ref` can never be processed twice.
- Deployment script (`script/Deploy.s.sol`) that deploys all three, wires roles, sets max supplies, and writes addresses to a JSON file the backend and frontend read. Contracts are verified on Basescan.

## 11. Backend (Go + Gin monolith)

One Go binary serves the REST API, WebSockets, background jobs, the matching engines, and the chain worker. Modules talk through Go interfaces and share one Postgres connection pool. MVP runs a single instance; a Postgres advisory lock makes sure only one process ever runs the engines and jobs.

&#91;embedded content: system architecture · 12 modules in one binary\]

The browser talks only to the monolith, except for wallet transactions it signs itself; the monolith alone writes to Postgres and sends relayer transactions.

**Folder layout**

```
cmd/server/main.go            wiring, config, graceful shutdown
internal/
  platform/  config, db (pgx + sqlc), redis, logger, http (gin, middleware), ws hub, scheduler
  auth/      signup, email codes, login, JWT, refresh, TOTP
  users/     profile, account state
  wallet/    SIWE nonce + link, wallet change
  payments/  intents, stripe (checkout, webhooks), crypto (intent, event handling)
  ledger/    accounts, journals, holds, balances  (the ONLY writer of balances)
  catalog/   items, recipes, upgrades, zones, loot tables, shop SKUs (config tables)
  game/      missions, crafting, upgrades, shop
  exchange/  markets, order service, engine (book + matcher), settlement
  marketdata/ tickers, candles, depth snapshots
  auction/   auctions, bids, settlement, system drops
  chain/     relayer, indexer, supply sync, withdrawals, deposits
  notify/    in-app notifications, email
  risk/      limits, flags, freezes
  admin/     admin APIs, audit log
migrations/                   SQL files (golang-migrate)
```

**Matching engine inside the monolith**

1. At boot, for each `OPEN` market, load its `OPEN` and `PARTIALLY_FILLED` orders ordered by `seq` and rebuild the in-memory book (bids and asks as price levels, each a FIFO queue).
2. Start one goroutine per market with a buffered command channel (size 1,000). Commands: `Place`, `Cancel`, `Halt`, `Resume`.
3. The order API handler inserts the order and its hold in a DB transaction, gets `seq` from a Postgres sequence, then sends `Place` to the channel. If the channel is full, the order is cancelled with `ENGINE_BUSY` and the hold released.
4. The goroutine matches in memory, then writes all fills, order updates, and ledger journals in one DB transaction. If that transaction fails, it reloads the book for that market from the DB and marks the order `REJECTED`.
5. After commit, it publishes book, trade, and ticker updates to the WebSocket hub and the private order/balance update to the involved users.
6. Because the DB is written before anything is published, a crash loses nothing: the book is rebuilt from the DB on restart.

**Data model (main tables)**

| Table | Key columns | Notes |
| --- | --- | --- |
| `users` | id, email, username, password\_hash, status, totp\_secret\_enc, created\_at | status enum from section 3 |
| `email_codes` | user\_id, code\_hash, purpose, expires\_at, attempts | purpose: verify, reset |
| `refresh_tokens` | id, user\_id, token\_hash, expires\_at, revoked\_at, replaced\_by | rotation chain |
| `wallets` | user\_id (unique), address (unique, lowercase), linked\_at | one per user |
| `siwe_nonces` | nonce, user\_id, expires\_at, used\_at |  |
| `idempotency_keys` | user\_id, key, request\_hash, response, created\_at | 24 h TTL |
| `payments` | id, user\_id, purpose, method, sku, amount\_nc, status, stripe\_session\_id, payment\_ref, tx\_hash, expires\_at |  |
| `external_events` | source, external\_id (unique), payload, processed\_at | Stripe events and chain logs |
| `ledger_accounts` | id, owner\_type, owner\_id, asset, bucket | unique (owner, asset, bucket) |
| `ledger_journals` | id, type, ref\_type, ref\_id, created\_at | unique (type, ref\_type, ref\_id) |
| `ledger_entries` | id, journal\_id, account\_id, amount, synced\_onchain | `synced_onchain` for item mint/burn entries |
| `ledger_balances` | account\_id, available, held | CHECKs from section 9 |
| `ledger_holds` | id, account\_id, amount, remaining, ref\_type, ref\_id, status |  |
| `items` | id, name, category, tier, tradeable, auction\_only, max\_supply, image\_url, attrs jsonb | catalog |
| `recipes`, `upgrades`, `zones`, `loot_tables`, `shop_skus` | config rows from section 4 | editable in admin |
| `missions` | id, user\_id, zone\_id, ship\_item\_id, drill\_item\_id, status, started\_at, ends\_at, loot jsonb, collected\_at | status: RUNNING, COMPLETED, COLLECTED, ABORTED |
| `craft_jobs` | id, user\_id, recipe\_id, qty, status, ends\_at |  |
| `markets` | id, symbol, item\_id, tick\_size, min\_qty, max\_qty, min\_notional, status |  |
| `orders` | id, seq, user\_id, market\_id, side, type, tif, price, qty, filled\_qty, status, hold\_id, created\_at | index (market\_id, status) |
| `trades` | id, market\_id, price, qty, maker\_order\_id, taker\_order\_id, maker\_fee, taker\_fee, created\_at |  |
| `candles` | market\_id, interval, open\_time, o, h, l, c, volume | closed candles |
| `auctions` | id, source, seller\_id, item\_id, qty, start\_price, reserve\_price, buy\_now\_price, highest\_bid\_id, starts\_at, ends\_at, original\_ends\_at, status |  |
| `bids` | id, auction\_id, user\_id, amount, hold\_id, created\_at |  |
| `withdrawals` | id, user\_id, kind, item\_id, qty, amount\_nc, to\_address, withdrawal\_ref, status, tx\_hash, attempts |  |
| `deposits` | id, user\_id, from\_address, item\_ids, amounts, tx\_hash, log\_index, status |  |
| `chain_txs` | id, nonce, kind, payload, tx\_hash, status, gas\_price, sent\_at | relayer queue |
| `notifications` | id, user\_id, type, data jsonb, read\_at |  |
| `admin_audit_log` | id, admin\_id, action, target, before, after, reason, created\_at | append-only |
| `risk_flags` | id, user\_id, type, details, status |  |

**API endpoints (all under `/api/v1`, JSON)**

| Area | Endpoints |
| --- | --- |
| Auth | `POST /auth/signup`, `POST /auth/verify-email`, `POST /auth/resend-code`, `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`, `POST /auth/forgot-password`, `POST /auth/reset-password`, `POST /auth/2fa/setup`, `POST /auth/2fa/enable`, `POST /auth/2fa/disable` |
| Me | `GET /me`, `PATCH /me`, `GET /me/balances`, `GET /me/inventory`, `GET /me/ledger?cursor=`, `GET /me/notifications`, `POST /me/notifications/read` |
| Wallet | `POST /wallet/nonce`, `POST /wallet/link`, `DELETE /wallet` |
| Payments | `POST /payments`, `GET /payments/{id}`, `GET /payments`, `POST /webhooks/stripe` |
| Catalog | `GET /items`, `GET /items/{id}`, `GET /recipes`, `GET /upgrades`, `GET /zones`, `GET /shop` |
| Game | `POST /shop/buy` (from NC), `POST /missions`, `GET /missions`, `POST /missions/{id}/collect`, `POST /missions/{id}/abort`, `POST /crafts`, `GET /crafts`, `POST /upgrades/{id}/craft` |
| Exchange | `GET /markets`, `GET /markets/{symbol}`, `GET /markets/{symbol}/book`, `GET /markets/{symbol}/trades`, `GET /markets/{symbol}/candles?interval=&from=&to=`, `POST /orders`, `GET /orders?status=&market=`, `GET /orders/{id}`, `DELETE /orders/{id}`, `DELETE /orders?market=` (cancel all), `GET /me/trades` |
| Auctions | `GET /auctions?status=&source=&item=`, `GET /auctions/{id}`, `GET /auctions/{id}/bids`, `POST /auctions`, `DELETE /auctions/{id}`, `POST /auctions/{id}/bids` |
| Withdrawals | `POST /withdrawals`, `GET /withdrawals`, `GET /withdrawals/{id}`, `GET /deposits` |
| Metadata | `GET /metadata/{id}.json` (public, used by NebulaItems URI) |
| Realtime | `GET /ws/ticket` (short-lived ticket), `GET /ws?ticket=` |
| Admin | `/admin/users`, `/admin/users/{id}/freeze`, `/admin/markets/{id}/halt`, `/admin/config/*`, `/admin/withdrawals/{id}/approve`, `/admin/payments/{id}/refund`, `/admin/adjustments`, `/admin/drops`, `/admin/reconciliation` |

Errors always use `{"error": {"code": "INSUFFICIENT_FUNDS", "message": "…", "details": {…}}}` with a fixed list of codes (e.g. `VALIDATION_FAILED`, `UNAUTHORIZED`, `ACCOUNT_NOT_ACTIVE`, `INSUFFICIENT_FUNDS`, `INSUFFICIENT_ITEMS`, `MARKET_HALTED`, `BID_TOO_LOW`, `LIMIT_EXCEEDED`, `WALLET_REQUIRED`, `TWO_FA_REQUIRED`, `RATE_LIMITED`, `ENGINE_BUSY`). Lists use cursor pagination.

**WebSocket channels**

One connection per browser tab, authenticated with a 30-second ticket from `GET /ws/ticket`. The client sends `{"op":"subscribe","channel":"…"}`.

| Channel | Public/private | Payload |
| --- | --- | --- |
| `tickers` | Public | Last price, 24h change, volume for all markets, every 1 s |
| `book:{symbol}` | Public | Top-20 depth snapshot, then diffs |
| `trades:{symbol}` | Public | Each new trade |
| `candles:{symbol}:{interval}` | Public | Current candle updates |
| `auction:{id}` | Public | New bid, new end time, settled |
| `user` | Private (auto) | Order updates, fills, balance changes, mission and craft completion, outbid, notifications |

**Background jobs**

Jobs run on an in-process scheduler. Row-based work uses `SELECT … FOR UPDATE SKIP LOCKED` so it is safe if we later run more than one instance.

| Job | Frequency | Does |
| --- | --- | --- |
| `mission_resolver` | 5 s | Rolls loot for missions past `ends_at`, sets `COMPLETED`, notifies |
| `craft_resolver` | 5 s | Delivers craft outputs, notifies |
| `auction_settler` | 1 s | Settles ended auctions |
| `auction_dropper` | 2 h | Creates system drop auctions |
| `order_expiry` | 1 h | Cancels GTC orders older than 30 days |
| `payment_expiry` | 1 min | Expires stale pending payments |
| `earned_settle` | 10 min | `earned_pending` → `earned` after 72 h |
| `supply_sync` | 5 min | Mint/burn deltas on-chain |
| `chain_indexer` | 3 s | Reads contract events |
| `relayer` | continuous | Sends queued `chain_txs` with nonce management |
| `withdrawal_worker` | 10 s | Processes approved withdrawals |
| `candle_closer` | 1 min | Persists closed candles |
| `ledger_check`, `supply_reconcile` | 1 h | Integrity checks |
| `stripe_reconcile`, `usdc_reconcile` | daily | Payment reconciliation |
| `cleanup` | daily | Deletes unverified accounts older than 7 days, old nonces and idempotency keys |

## 12. UI pages and components

The frontend is a Next.js App Router app with a dark space theme in Tailwind. Price up is green, price down is red, and every colored signal also has an arrow or sign for colorblind users. Server data is fetched with TanStack Query; live data comes from one shared WebSocket client. All amounts are formatted from integer micro-units with a decimal library, never JavaScript floats.

**Global layout (every logged-in page)**

- `TopNav`: logo, links (Hangar, Missions, Workshop, Shop, Exchange, Auctions, Wallet), `BalanceChip` (total NC, click shows bucket breakdown), `NotificationBell` with unread count, `ProfileMenu`.
- `TickerStrip`: scrolling row of markets with last price and 24h change, from the `tickers` channel.
- `AccountStateBanner`: shows for `PENDING_PAYMENT` ("Pay entry fee to start") and `FROZEN`.
- `ToastProvider`, `ConfirmDialog`, `TwoFactorPrompt` (modal asking for TOTP), `ConnectionStatus` (WebSocket reconnecting indicator).

**Shared components**

`ItemIcon`, `ItemCard` (icon, name, tier, qty), `NcAmount` (formatted NC with bucket tooltip), `PriceChange` (arrow + %), `CountdownTimer`, `StatusBadge`, `PaymentMethodPicker` (NC balance / Card / Wallet, disabling options that are not allowed), `WalletConnectButton` (wagmi), `TxStatus` (hash link to Basescan + confirmations), `DataTable` (sortable, paginated), `EmptyState`, `Skeleton`, `ErrorState`.

**Pages**

| Page | Route | Components and behavior |
| --- | --- | --- |
| Landing | `/` | Hero, how-it-works steps, live `TickerStrip` (public), top movers, featured live auctions, Sign up / Log in CTAs |
| Sign up | `/signup` | Email, username, password with strength meter, 18+ and Terms checkbox, inline field errors |
| Verify email | `/verify` | 6-digit code input, resend button with 60 s countdown, attempts left |
| Log in | `/login` | Email, password, TOTP step if 2FA on, lockout message, forgot-password link |
| Forgot / reset password | `/forgot`, `/reset?token=` | Email form; new password form |
| Entry fee | `/onboarding/pay` | What you get (starter pack preview), price 5 NC, `PaymentMethodPicker` (Card or Wallet only), `WalletConnectButton` if Wallet chosen |
| Payment result | `/payment/result?id=` | Polling spinner, success (what was credited, CTA), failure with retry, `TxStatus` for crypto |
| Hangar (dashboard) | `/hangar` | Onboarding checklist, balance summary, running missions with `CountdownTimer` and Collect button, craft queue, open orders count, your live auctions and leading bids, recent notifications |
| Missions | `/missions` | `ZoneCard` list (duration, fuel, loot ranges, requirements, locked state), `MissionLauncher` (pick ship, drill, zone; shows fuel cost and expected loot range), `ActiveMissionList`, `MissionHistory` with loot results |
| Workshop | `/workshop` | Tabs: Crafting and Upgrades. `RecipeCard` (inputs owned/needed, time, fee, qty stepper, Craft button), `CraftQueue`, `UpgradeCard` (current → next tier, craft path vs buy path, `PaymentMethodPicker` for buy path) |
| Shop | `/shop` | `ShopItemCard` grid, `PurchaseModal` with quantity, total, `PaymentMethodPicker` (NC, Card, Wallet) |
| Inventory | `/inventory` | Filter by category, `ItemCard` grid with available vs held qty, item actions: Sell on exchange, Auction, Withdraw, View market |
| Item detail | `/items/[id]` | Image, attributes, supply stats (total, held in Vault, withdrawn), mini price chart, last auction prices, where it is used (recipes, upgrades) |
| Markets | `/exchange` | `MarketsTable` (symbol, last, 24h change, high, low, volume, sparkline), tabs Top gainers / Top losers / Volume, search, category filter |
| Trading | `/exchange/[symbol]` | `CandleChart` (lightweight-charts, interval switch), `OrderBook` (depth, click price to fill form), `RecentTrades`, `OrderForm` (Buy/Sell tabs; Limit/Market; price, qty, total, fee, available balance, % buttons, IOC toggle), `OpenOrders` (cancel, cancel all), `OrderHistory`, `MyTrades`, halted banner |
| Auctions | `/auctions` | Tabs Live / Ending soon / Ended / My auctions / My bids, `AuctionCard` (item, current bid, `CountdownTimer`, bid count, System badge), filters |
| Auction detail | `/auctions/[id]` | Item panel, current bid, reserve met/not met, buy-now price, `CountdownTimer` that updates on extension, `BidForm` (min bid hint, amount, confirm), `BidHistory` live list, seller controls (cancel if no bids) |
| Create auction | `/auctions/new` | Item picker from inventory, qty, start price, reserve, buy-now, duration, fee preview, confirm |
| Wallet | `/wallet` | `BalanceBreakdown` (card, crypto, earned pending with release dates, earned, held), `TopUpPanel` (amount presets, Card or Wallet), `WalletLinkCard` (connect via SIWE, linked address, change/unlink), `WithdrawPanel` (NC or items, limits, fee, 2FA), `DepositItemsPanel` (tokens in wallet, select, deposit), `TestUsdcFaucet` (testnet only), `WithdrawalList`, `DepositList` |
| History | `/history` | Tabs: Ledger (every journal with type and amounts), Payments, Trades, Missions, Crafts, Auctions |
| Settings | `/settings` | Profile (username), password change, 2FA setup (QR + verify), active sessions with revoke, email preferences |
| Notifications | `/notifications` | Full list, mark all read, links to the related page |
| Admin | `/admin/*` | Users (search, view, freeze), markets (halt, config), config editors (items, recipes, zones, loot, shop, fees, limits), withdrawal review queue, payments and refunds, auction drop table, reconciliation dashboard, economy dashboard (minted vs burned per item), audit log |

**UX rules developers must follow**

- Every money or item action shows a confirmation with exact amounts and fees before sending.
- Every POST sends a fresh `Idempotency-Key` (UUID) and disables the button until the response arrives.
- Order and bid results are shown from WebSocket events; if the socket is down, the page polls every 3 seconds.
- Wallet actions check the chain ID and ask the wallet to switch to Base Sepolia if wrong.
- Pages must work at 375 px wide. The trading page stacks chart, form, and book vertically on mobile.

## 13. Notifications, admin, security, and compliance

**Notifications**

| Event | In-app | Email |
| --- | --- | --- |
| Email code, password reset, wallet linked or changed, 2FA changed | — | Yes |
| Payment succeeded or failed | Yes | Yes |
| Mission completed, craft completed | Yes | No |
| Order filled (full or partial), order cancelled by system | Yes | No |
| Outbid, auction won, auction sold or unsold | Yes | Won/sold only |
| Withdrawal completed, failed, or needs review | Yes | Yes |
| Account frozen | Yes | Yes |

**Admin panel**

Admins log in with the same auth plus mandatory 2FA and an `is_admin` flag. Every admin write is stored in `admin_audit_log` with before/after values and a reason. Admins can:

- Search users, view their ledger, freeze or unfreeze (freeze cancels open orders and voids live auctions).
- Halt or reopen markets and edit market settings.
- Edit catalog config: items, recipes, upgrades, zones, loot tables, shop prices, fees, limits.
- Approve or reject withdrawals in `PENDING_REVIEW`.
- Refund unspent card payments and review disputes.
- Manage the system auction drop table.
- Post `ADMIN_ADJUSTMENT` journals (needs a second admin's approval).
- View reconciliation results and economy stats (daily minted vs burned per item, NC in circulation by bucket, fees earned).

**Security**

- Passwords: argon2id. TOTP secrets and any stored keys encrypted with a key from the environment (KMS in production).
- Auth: access JWT in memory only, refresh token in httpOnly cookie, CSRF protection via SameSite=Strict plus a custom header check on state-changing requests.
- Rate limits in Redis: signup 5/hour/IP, login 10/15 min/IP, orders 10/s/user, bids 5/s/user, general API 100/min/user.
- Stripe webhooks verified with the signing secret; SIWE checks domain, chain ID, nonce, and expiry.
- All amounts validated server-side; the client never sends prices for shop items, only SKUs.
- CORS limited to the frontend origin. Security headers (CSP, HSTS) on the Next.js app.
- Structured logs with request IDs; no secrets, passwords, or full card data ever logged (Stripe hosts card entry).

**Anti-abuse**

- Self-trade prevention in the engine; wash-trading report for admins (pairs of accounts trading mostly with each other).
- Multi-account signals: same IP, device fingerprint, or deposit address across accounts create a `risk_flag`, not an automatic ban.
- Withdrawal velocity rules: first withdrawal and anything over 250 NC go to manual review.
- Circuit breakers on markets (section 7) and on-chain daily USDC cap (section 10).

**Compliance (must be resolved before any mainnet launch)**

MVP runs only on testnet with Stripe test mode and MockUSDC, so no real money moves. Before real money, get legal advice on: whether an entry fee plus random mission loot counts as gambling in target countries; money-transmitter and virtual-asset rules for USDC withdrawals; KYC requirements for withdrawals; consumer protection and refund rules for card purchases; age verification (18+); and geo-blocking of restricted countries. Terms of Service and a Privacy Policy pages are required at launch.

## 14. Edge cases, build plan, and acceptance criteria

**Edge cases and expected behavior**

| Situation | Expected behavior |
| --- | --- |
| Player double-clicks Buy / Place order | Same `Idempotency-Key` → one action, same response returned |
| Stripe webhook arrives twice or out of order | Event ID dedupe; payment already `SUCCEEDED` → ignore |
| Stripe webhook arrives before the redirect | Result page shows success on first poll |
| Player closes tab during crypto payment | Indexer still credits when the event confirms; notification sent |
| Server restarts mid-match | DB transaction either committed (book rebuilt with fills) or not (order re-read as `OPEN` and re-queued in `seq` order) |
| Engine DB write fails | Market book reloaded from DB, order `REJECTED`, hold released |
| Market halted while player has open orders | Orders stay; new orders rejected with `MARKET_HALTED`; cancels allowed |
| Two bids at the same millisecond | Row lock serializes them; the second must beat the first's new minimum or gets `BID_TOO_LOW` |
| Bid at 0:59 remaining | End time extends to bid time + 60 s (cap 30 min total) |
| Player's items are held by an order and they try to start a mission / auction / withdraw them | Rejected with `INSUFFICIENT_ITEMS`; UI shows held quantity |
| Ship or drill on a mission | Held; cannot be sold, upgraded, or withdrawn until collected |
| Withdrawal tx stuck | Relayer bumps gas after 3 min; after 3 failures, `FAILED` and holds released |
| Relayer out of gas | Chain jobs pause, admin alerted; game and trading keep working |
| Chain RPC down | Indexer retries with backoff; deposits and crypto payments are delayed, not lost |
| Vault USDC reserve too low | Withdrawal waits in `PENDING_LIQUIDITY`; admin alerted |
| Chargeback after buyer spent card NC on a trade | Buyer frozen and `card` goes negative; seller's `earned_pending` is not clawed back automatically; admin decides |
| Player loses 2FA device | Support flow: identity check by email + admin reset; withdrawals blocked 7 days after reset |
| Deposit from an unlinked address | Items go to `unclaimed`; credited when the address is linked |
| Item config changed while missions run | Running missions use the zone and loot values saved on the mission row at start |

**Build plan (MVP phases)**

1. **Foundation (week 1–2):** repo setup, Postgres migrations, config, Gin skeleton, auth (signup, verify, login, refresh, reset), Next.js layout and auth pages.
2. **Ledger and catalog (week 3):** ledger package with holds and full tests, catalog tables seeded with section 4 values, inventory and balances APIs.
3. **Contracts (week 3–4, in parallel):** NebulaItems, NebulaVault, MockUSDC with Foundry tests and deploy script on Base Sepolia.
4. **Payments (week 4–5):** Stripe Checkout + webhooks, wallet link (SIWE), crypto payments + indexer, entry fee and starter pack, shop.
5. **Gameplay (week 6):** missions, crafting, upgrades, Hangar, Missions, Workshop, Inventory pages.
6. **Exchange (week 7–8):** engine with unit and property tests, order APIs, WebSocket hub, market data, Markets and Trading pages.
7. **Auctions (week 9):** player auctions, bidding, anti-sniping, settlement, system drops, auction pages.
8. **Chain sync and withdrawals (week 10):** supply sync, relayer, withdrawals, deposits, Wallet page, reconciliation jobs.
9. **Admin and hardening (week 11–12):** admin panel, rate limits, risk flags, load test (500 orders/s on one market), security review, bug fixing.

**Acceptance criteria for MVP**

- [ ] A new player can sign up with email, verify, pay the entry fee by card or USDC, and receive the starter pack.
- [ ] A player can link a wallet with SIWE; linking the same address to a second account is rejected.
- [ ] Shop items and buy-path upgrades can be paid from NC, by card, and with USDC, and all three produce the same ledger result.
- [ ] Missions, crafting, and upgrades work with loot, timers, and holds as specified in section 4.
- [ ] Limit, market, and IOC orders match with price-time priority, correct fees, and correct hold releases, verified by engine unit tests.
- [ ] Order book, trades, candles, and tickers update live in the browser within 1 second of a trade.
- [ ] Auctions accept bids, extend on late bids, and settle correctly, including unsold and buy-now cases.
- [ ] Items and USDC can be withdrawn to and items deposited from the linked wallet; card NC can never be withdrawn.
- [ ] Hourly `ledger_check` and `supply_reconcile` report zero mismatches after a full test run.
- [ ] Restarting the server during active trading loses no orders, fills, or balances.
- [ ] All contract tests, including the invariant test, pass; contracts are verified on Basescan.

**Open questions**

- Final game name, art style, and item images: who produces them?
- Should card-paid entry fees and card NC be priced differently to cover Stripe fees (about 3% + 30¢)?
- Is the 72-hour `earned` hold acceptable for players, or should it be 24 hours for small amounts?
- Should withdrawn items keep counting toward in-game supply stats on the item page?
- Which email provider and which hosting (e.g. Fly.io, Railway, AWS) for the Go binary and Postgres?
