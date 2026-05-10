---
name: ibkr-flex
description: Query the user's Interactive Brokers (IBKR) portfolio via the local `ibkr-flex` Go CLI installed at `~/go/bin/ibkr-flex`. Read-only access to positions, trades, cash flows, dividends, account NAV (净值), MTM performance, and full daily P&L attribution — sourced from IBKR Flex Web Service with T+1 latency. Use this skill whenever the user asks about their IBKR / Interactive Brokers account state, even if they don't say "IBKR" or "ibkr-flex" explicitly. Triggers on phrases like "我的持仓" / "我的盈利" / "我赚了多少" / "最近交易" / "账户净值" / "分红" / "实际收益" / "投资表现" / "PnL 拆解" / "my positions" / "portfolio P&L" / "how much have I made" / "what's in my account". Encodes the user's account ID, query ID, common analysis recipes, and known data quirks so future sessions don't have to rediscover them.
---

# ibkr-flex

A Go CLI for read-only IBKR Flex Web Service queries against the user's account. Repo lives at `<repo root>/`.

## User context (固定信息)

- **Binary**: `ibkr-flex` (in `~/go/bin/`, on PATH)
- **Account ID**: `U1234567`
- **Account base currency**: SGD
- **Flex Query ID**: `1234567` (named "claude" in the IBKR portal — bundled Activity Flex Query)
- **Account opened**: ~2025-05-07
- **Data freshness**: **T+1** — never has today's intraday data. If the user wants intraday/realtime, this skill cannot help.

## Configuration

The CLI reads `IBKR_FLEX_TOKEN` and `IBKR_FLEX_QUERY_ID` from either:
- Environment variables (override), or
- `~/.config/ibkr-flex/config` (dotenv-style `KEY=VALUE`, chmod 600)

Default config file should already be populated for this user. If `ibkr-flex accounts` returns `error: config: required value not set`, the file is missing or empty — tell the user, ask them to paste fresh credentials. Never invent a token.

## Subcommands

All data subcommands except `accounts` require `--account U1234567`. Default output is JSON.

| Command | Returns |
|---|---|
| `ibkr-flex accounts` | List of account IDs visible to the token (sanity check) |
| `ibkr-flex positions --account U1234567` | Daily position snapshots (per-day, per-symbol, T+1) |
| `ibkr-flex trades --account U1234567` | All trade executions in the query period |
| `ibkr-flex cash --account U1234567` | Cash transactions: deposits, withdrawals, fees, interest, paid dividends |
| `ibkr-flex dividends --account U1234567` | Dividend accruals (declared + reversed pairs — see Quirks) |
| `ibkr-flex nav --account U1234567` | Daily account NAV time series |
| `ibkr-flex mtm --account U1234567` | Per-symbol mark-to-market performance entries |
| `ibkr-flex pnl --account U1234567` | Daily P&L attribution: mtm, fx, dividends, commissions, withholding, deposits/withdrawals, grants, TWR |
| `ibkr-flex pnl --summary --account U1234567` | Same components, summed across the period — single cumulative record |

Universal flags:
- `--account ID` — required for data commands
- `--format json|table|csv` — default `json`. Use `table` for human view, `json` for jq pipelines
- `--since YYYY-MM-DD` / `--until YYYY-MM-DD` — client-side date filter (inclusive)
- `--refresh` — skip cache and re-fetch from IBKR (~5-30s)
- `-v` — print fetch progress to stderr

## Cache

First call of the day fetches the bundled query (5-30s server-side generation) and writes XML to `~/Library/Caches/ibkr-flex/1234567-YYYYMMDD.xml`. Subsequent same-day calls reuse it (instant). Auto-refetches across days. Cache survives across Claude sessions, so repeated questions in the same day don't re-hit IBKR.

## Recipes for common questions

### "我的持仓" / "What's in my portfolio?"

Position records are daily history. Get the latest snapshot:

```bash
LATEST=$(ibkr-flex positions --account U1234567 --format=json | jq -r '[.[].reportDate]|max')
ibkr-flex positions --account U1234567 --since "$LATEST" --format=table
```

Then group/summarize by currency or asset type:

```bash
ibkr-flex positions --account U1234567 --since "$LATEST" --format=json | \
  jq 'group_by(.currency) | map({ccy: .[0].currency, total: (map(.positionValue)|add), positions: .})'
```

### "我的实际盈利" / "How much have I actually made?"

Use `pnl --summary` — that's literally what it's for. Returns the canonical attribution:

```bash
ibkr-flex pnl --summary --account U1234567 --format=table
```

Result columns: starting/ending NAV, MTM (trading P&L), FX impact (translation + trading), dividends (received + accrual changes), interest, commissions, withholding tax, deposits/withdrawals, grants, time-weighted return.

For a clean "total profit" number:

```bash
ibkr-flex pnl --summary --account U1234567 --format=json | \
  jq '.[0] | (.endingValue - .startingValue - .depositsWithdrawals)'
```

For a date-bounded view (e.g., YTD, last quarter):

```bash
ibkr-flex pnl --summary --account U1234567 --since 2026-01-01 --format=table
```

### "最近交易" / "Recent trades"

```bash
ibkr-flex trades --account U1234567 --since 2026-04-01 --format=table
```

Note: side is preserved as IBKR-native `BUY`/`SELL` strings, quantity is always positive. To compute net flow per symbol with signed quantity:

```bash
ibkr-flex trades --account U1234567 --since 2026-04-01 --format=json | \
  jq 'map({symbol, signed: (if .buySell=="SELL" then -.quantity else .quantity end), netCash})'
```

### "分红收了多少" / "How much in dividends did I actually receive?"

The `dividends` section returns IBKR's accrual ledger — each dividend appears as `+amount` (declaration) and `-amount` (reversal once paid), netting to 0. For **received cash**, use `cash` filtered to dividend-related types:

```bash
ibkr-flex cash --account U1234567 --format=json | \
  jq '[.[] | select(.type | test("Dividend|Payment In Lieu"))]'
```

Or just look at `pnl --summary`'s `dividends` field (post-tax cumulative).

### "现金流 / 入金" / "Cash deposits"

```bash
ibkr-flex cash --account U1234567 --format=json | \
  jq '[.[] | select(.type | startswith("Deposits"))] | map({date, amount, currency})'
```

### "账户净值曲线" / "NAV over time"

NAV has duplicate rows per date (different model views). Dedupe by date:

```bash
ibkr-flex nav --account U1234567 --format=json | \
  jq 'group_by(.date) | map(.[0]) | sort_by(.date)'
```

For TWR (time-weighted return) curve, prefer `pnl`'s `twr` field which IBKR pre-computes per day.

## Known quirks (don't try to "fix" — they're how IBKR works)

1. **`positions` cost basis & unrealized PnL are 0.** The current Flex Query's OpenPositions section does not have those fields enabled. Account-level PnL still works exactly via `pnl --summary` (NAV-based math). Per-position cost basis requires editing the IBKR query to add the fields — tell the user but don't try to compute around it from trades, since FIFO matching across years is error-prone.

2. **Dividends look duplicated and net to 0.** Each dividend has an accrual entry (positive) + reversal entry (negative once paid in cash). This is IBKR's accounting, not a bug. Use `cash` for actually-received cash, or `pnl`'s `dividends` field.

3. **NAV has 2 rows per date.** Two records per date (`model=Independent` + aggregated). Either is fine; dedupe on date.

4. **MTM rows with empty `symbol`.** These are `description="Total P/L"` account-level rollups, not bugs. Filter out for per-symbol MTM views.

5. **`pnl` records are per-day even when nothing happened.** ChangeInNAV emits a record for every report date in range, even weekends and holidays (typically zeroes). Sum still works correctly.

6. **Multi-currency.** Account base is SGD. Positions span SGD/USD/HKD. The typed output shows native currency; the raw XML has `*Base` fields that pre-convert to SGD. For cross-currency totals, prefer NAV (already in SGD).

## Read-only guarantee

The Flex Web Service exposes only report retrieval endpoints — there is no order-placement surface. The CLI calls only `SendRequest` and `GetStatement`. No code path can modify the user's account.

## When to add a new typed subcommand

If the user repeatedly asks for data the existing subcommands don't expose, add a typed subcommand instead of telling them to grep the XML. The cached XML at `~/Library/Caches/ibkr-flex/1234567-YYYYMMDD.xml` has 70+ section types; the most commonly-requested ones to promote next are listed in `references/raw-xml.md` along with the recipe for adding a new section. Use that file as an escape hatch only, not as the primary path.

## Building / dev workflow

If the binary is missing or stale:

```bash
cd <repo root>
go install ./cmd/ibkr-flex   # rebuilds and installs to ~/go/bin
```

Tests and vet:

```bash
go vet ./... && go test ./...
```
