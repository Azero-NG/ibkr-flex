# ibkr-flex

Read-only Go CLI for the IBKR Flex Web Service. Pulls one bundled Activity Flex Query and slices it by data dimension (trades, positions, cash flows, dividends, NAV, mark-to-market). Designed primarily as a JSON data source for AI tooling; `--format=table` and `--format=csv` are available for human consumption.

**Latency**: T+1. Today's executions appear after IBKR's overnight reconciliation. Use a TWS-API-based tool if you need intraday data.

## Install

```bash
go install github.com/azero/ibkr-flex/cmd/ibkr-flex@latest
```

Or build from source:

```bash
git clone <this repo>
cd ibkr
go build -o ibkr-flex ./cmd/ibkr-flex
```

## Setup

1. Generate a Flex Token and configure a bundled Activity Flex Query in the IBKR client portal — full walkthrough in [docs/flex-setup.md](docs/flex-setup.md).
2. Provide the credentials via either env vars or a config file (env wins when both are set):

   **Option A — env vars** (transient):
   ```bash
   export IBKR_FLEX_TOKEN=...
   export IBKR_FLEX_QUERY_ID=...
   ```

   **Option B — config file** (persistent across sessions):
   ```bash
   mkdir -p ~/.config/ibkr-flex
   cat > ~/.config/ibkr-flex/config <<EOF
   IBKR_FLEX_TOKEN=...
   IBKR_FLEX_QUERY_ID=...
   EOF
   chmod 600 ~/.config/ibkr-flex/config
   ```

   The default path is `${XDG_CONFIG_HOME:-$HOME/.config}/ibkr-flex/config`; override with `IBKR_FLEX_CONFIG=/some/other/path`.

3. Verify:

   ```bash
   ibkr-flex accounts
   ```

## Usage

```bash
# List account IDs visible to the token
ibkr-flex accounts

# Data slices (each requires --account)
ibkr-flex trades     --account U1234567
ibkr-flex positions  --account U1234567
ibkr-flex cash       --account U1234567
ibkr-flex dividends  --account U1234567
ibkr-flex nav        --account U1234567
ibkr-flex mtm        --account U1234567

# Daily P&L attribution (mtm / fx / dividends / commissions / withholding / ...)
ibkr-flex pnl        --account U1234567
ibkr-flex pnl --summary --account U1234567   # cumulative components across the period

# Date filter (client-side, applies to the section's primary date field)
ibkr-flex trades --account U1234567 --since 2026-04-01 --until 2026-05-01

# Alternate output formats
ibkr-flex positions --account U1234567 --format=table
ibkr-flex trades    --account U1234567 --format=csv > trades.csv

# Skip cache and re-fetch from IBKR (otherwise cached for the day)
ibkr-flex trades --account U1234567 --refresh

# Show fetch progress on stderr
ibkr-flex trades --account U1234567 -v
```

## Caching

The first command of the day fetches the full Activity Flex Query (one HTTP round-trip with up to 60s server-side report generation) and writes it under Go's `os.UserCacheDir()`:

- macOS: `~/Library/Caches/ibkr-flex/<query-id>-YYYYMMDD.xml`
- Linux: `~/.cache/ibkr-flex/<query-id>-YYYYMMDD.xml`

Subsequent commands the same day reuse this file. Cache auto-expires when the date changes; use `--refresh` to force a re-fetch.

## Output schema

All JSON output preserves IBKR's native semantics — for example, a sell trade is reported as `{"buySell": "SELL", "quantity": 100}` (positive number) rather than negated. Core fields are typed; everything else from the XML lands in `extra` (string-valued map) so no data is lost when IBKR adds fields.

Field reference per section (subject to IBKR's ongoing schema additions):

- `trades`: tradeId, tradeDate, symbol, secType, currency, exchange, buySell, quantity, tradePrice, commission, netCash, account, extra
- `positions`: symbol, secType, currency, position, markPrice, positionValue, costBasis, unrealizedPnl, account, reportDate, extra
- `cash`: date, type, amount, currency, description, account, extra
- `dividends`: date, symbol, amount, currency, account, extra
- `nav`: date, total, currency, account, extra
- `mtm`: date, symbol, mtm, currency, account, extra
- `pnl`: date, currency, startingValue, endingValue, mtm, realized, changeInUnrealized, fxTranslation, netFxTrading, dividends, changeInDividendAccruals, interest, commissions, otherFees, withholdingTax, depositsWithdrawals, grantActivity, twr, account, extra

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | Required env var missing (`IBKR_FLEX_TOKEN` or `IBKR_FLEX_QUERY_ID`) |
| 2 | Auth or Flex error (invalid token, invalid query ID, etc.) |
| 3 | Flex report generation timed out (>60s) |
| 4 | Other (parse error, network, unknown) |

## Claude Code skill

This repo ships a [Claude Code](https://claude.com/claude-code) skill at `.claude/skills/ibkr-flex/` that teaches Claude how to drive this CLI — common analysis recipes, IBKR data quirks, the canonical "how much have I made" answer, and the read-only guarantee — so you can ask portfolio questions in natural language (`我的持仓`, `how much have I made`, `recent trades`) without re-explaining the schema every session.

### Project-level (auto-loaded)

Claude Code auto-loads skills under `.claude/skills/` of the current project. When you launch Claude Code inside this repo, the skill is already active — no install step.

```bash
git clone https://github.com/Azero-NG/ibkr-flex.git
cd ibkr-flex
claude   # the ibkr-flex skill is now available in this session
```

### User-level (available in any directory)

Symlink the skill into your global skills directory:

```bash
mkdir -p ~/.claude/skills
ln -s "$(pwd)/.claude/skills/ibkr-flex" ~/.claude/skills/ibkr-flex
```

After installing (either path), edit `SKILL.md` to replace the example values with yours:

- **Account ID** (e.g., `U1234567`) — your IBKR account
- **Flex Query ID** (e.g., `1234567`) — the bundled Activity Flex Query you configured in [docs/flex-setup.md](docs/flex-setup.md)
- **Account base currency** — for multi-currency totals

The skill assumes the binary is on `PATH` as `ibkr-flex` (default location: `~/go/bin/ibkr-flex` after `go install`).

## Read-only guarantee

The Flex Web Service exposes only report retrieval endpoints — there is no write surface. This tool calls only `SendRequest` and `GetStatement`; no order placement, modification, or cancellation is possible through it.
