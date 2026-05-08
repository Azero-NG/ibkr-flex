# Flex Web Service Setup

`ibkr-flex` reads a single bundled Activity Flex Query through the IBKR Flex Web Service. Configure once on the IBKR side, set two env vars locally, and you're done.

## 1. Enable Flex Web Service & generate Token

1. Log into [Client Portal](https://www.interactivebrokers.com/sso/Login).
2. Go to **Performance & Reports → Flex Queries** (some legacy UIs put it under **Reports → Settings**).
3. Find **Flex Web Service** → click **Configure**.
4. Toggle to **Enabled**, set an IP allowlist if you want, then **Save**.
5. Click **Generate Token**.
6. **Copy the token immediately** — IBKR shows it exactly once. Token is a 16-digit numeric string.

## 2. Create the bundled Activity Flex Query

1. Same page: **Performance & Reports → Flex Queries → Activity Flex Query → Create**.
2. **Query Name**: anything, e.g. `ibkr-flex-bundle`.
3. **Format**: `XML`.
4. **Period**: `Last 365 Calendar Days` (or whatever range you want — ibkr-flex caches per day so wider is fine).
5. **Sections** — enable all six (these map to the six subcommands):
   - **Trades**
   - **Open Positions**
   - **Cash Transactions**
   - **Change in Dividend Accruals**
   - **Net Asset Value (Equity Summary in Base)** — sometimes labeled **NAV**
   - **Mark-to-Market Performance Summary in Base**
6. For each section, click **All** under fields (default is a subset).
7. **Save** the query.
8. Note the **Query ID** column — a 7-digit number like `1234567`.

## 3. Set environment variables

```bash
export IBKR_FLEX_TOKEN=1111222233334444
export IBKR_FLEX_QUERY_ID=1234567
```

Persist them in `~/.zshrc` / `~/.bashrc`, or use a `.env` loader of your choice (the binary itself doesn't read `.env`).

## 4. Verify

```bash
ibkr-flex accounts
```

If the token and query ID are valid, you'll see a JSON array of account IDs visible to the token. Pick the one you want and use it as `--account` for all data commands.

```bash
ibkr-flex trades --account U1234567
```

## Troubleshooting

| Symptom | Likely cause |
|---|---|
| `flex: token invalid or expired` (exit 2) | Wrong `IBKR_FLEX_TOKEN`, or it expired (regenerate) |
| `flex: invalid query id` (exit 2) | Wrong `IBKR_FLEX_QUERY_ID`, or query was deleted |
| `flex: statement generation timed out` (exit 3) | IBKR backend slow; retry, or increase `--verbose` to see progress |
| Empty results despite valid trades | Check the query's **Period** and **Sections** — wider period and "All" fields |
| HTTP 4xx / network errors (exit 4) | Flex Web Service IP allowlist may be blocking you |

## Notes on data freshness

Flex data is **T+1**: today's trades and end-of-day positions usually appear after IBKR's overnight reconciliation. Realtime monitoring is not the use case for this tool.
