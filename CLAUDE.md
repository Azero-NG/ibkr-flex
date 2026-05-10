# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A read-only Go CLI (`ibkr-flex`) for the IBKR Flex Web Service. Pulls one bundled Activity Flex Query (XML) and slices it into typed sections (trades, positions, cash, dividends, NAV, MTM, daily P&L attribution). Primary consumer is AI tooling (JSON output by default) — `--format=table|csv` is for human use. Read-only is structural, not just policy: the Flex Web Service has no write surface, and `internal/flex/client.go` only calls `SendRequest` + `GetStatement`.

**Latency is T+1.** No intraday data is ever available through this tool — that's an IBKR-side limitation, not something to "fix" in code.

## Common commands

```bash
go build -o ibkr-flex ./cmd/ibkr-flex          # local binary
go install ./cmd/ibkr-flex                     # install to ~/go/bin
go vet ./...                                   # static checks (CI gate)

# Smoke test against real IBKR (requires IBKR_FLEX_TOKEN + IBKR_FLEX_QUERY_ID)
ibkr-flex accounts                             # cheapest sanity check — lists account IDs
```

There are currently no `*_test.go` files. If you add tests, prefer table-driven unit tests against captured XML fixtures (place under `testdata/`) over hitting the live IBKR endpoint.

## Architecture

### Data flow (one line per stage)

```
config.Load()      env vars > ~/.config/ibkr-flex/config (dotenv-style)
cache.Get()        ~/Library/Caches/ibkr-flex/<queryID>-YYYYMMDD.xml — one fetch per day
flex.Client.Fetch  SendRequest → poll GetStatement (3s × up to 60s) → raw XML bytes
flex.Parse         streaming xml.Decoder → typed *Statement (Trades, Positions, …)
subcommand RunE    filter Statement.<section> by --account + --since/--until
output.go          renderJSON | renderTable (tabwriter) | renderCSV
```

The cache key is `queryID + YYYYMMDD` only — **token is not part of the key**, so multiple users sharing a queryID would collide. Not a problem in practice (one user per machine), but worth knowing before adding multi-account logic.

### Layering

- `cmd/ibkr-flex/` — `main` package: cobra commands, global flags, rendering. One file per subcommand (`trades.go`, `positions.go`, …) + `output.go` (all renderers) + `root.go` (shared `loadStatement`, `requireAccount`, `dateInRange`, `exitCodeFor`).
- `internal/flex/` — IBKR client (`client.go`), error mapping (`errors.go`), XML → typed struct decoder (`parse.go`), section types (`types.go`). Self-contained — does not import anything from `cmd/`.
- `internal/config/` — env-first, dotenv-fallback resolution. `Source` field records provenance (`"env"`, `"<path>"`, or `"env+<path>"`).
- `internal/cache/` — per-day XML cache, atomic write via temp-file + rename.

### The `Extra map[string]string` pattern (important)

In `parse.go`, each section uses an `extractor` that pulls **known** XML attributes via `take(...)` (which `delete`s them from the attr map), then drops everything left into `Extra`. This is how the parser stays robust to IBKR adding new attributes — nothing is silently lost. **When you add a typed field, also remove it from `Extra` by adding it to the `take` call.** Don't read from `Extra` in subcommands; promote the field to a typed one if you need it.

### Adding a new section/subcommand

1. `internal/flex/types.go` — define struct (use `Extra map[string]string` last)
2. `internal/flex/parse.go` — add `decodeXxx` + new `[]Xxx` field on `Statement` + new `case` in the parse switch (matched against IBKR's element name like `OpenPosition`, `CashTransaction`, `ChangeInNAV` — see `references/raw-xml.md` in the skill for IBKR's full element catalog)
3. `cmd/ibkr-flex/output.go` — add `renderXxx` covering `json`/`table`/`csv` (json branch is one line; table/csv mirror the field set)
4. `cmd/ibkr-flex/xxx.go` — cobra command that calls `requireAccount`, `loadStatement`, filters by `globals.account` + `dateInRange`, returns `renderXxx`
5. `cmd/ibkr-flex/root.go` — register in `root.AddCommand(...)`

The `pnl` subcommand (`cmd/ibkr-flex/pnl.go`) is the most non-trivial example: it adds a `--summary` flag and aggregates `ChangeInNAV` rows. The canonical "actual profit" formula is documented in its `Long` string — keep that in sync if the section schema ever changes.

### Error → exit code mapping (`root.go:exitCodeFor`)

| Error | Code | When |
|---|---|---|
| `config.ErrMissing` | 1 | Token or query ID not provided |
| `flex.ErrInvalidToken` / `flex.ErrInvalidQueryID` / any `*FlexError` | 2 | IBKR rejected the request |
| `flex.ErrStatementTimeout` | 3 | Generation took >60s |
| anything else | 4 | Network, parse, etc. |

Map new error types here if you add them; the exit codes are part of the CLI contract.

## Constraints to preserve

- **Read-only structural guarantee.** Don't add any HTTP call beyond `SendRequest` / `GetStatement`. Keep `internal/flex/client.go` as the only place that talks to IBKR.
- **Don't shape data away from IBKR's native semantics.** Sells stay as `{buySell:"SELL", quantity:100}` (positive quantity) — don't negate. Dividend accruals appear as ±pairs that net to 0 — don't dedupe at parse time. The skill at `skills/ibkr-flex/SKILL.md` documents these quirks for downstream consumers; if you change parser semantics, update that file too.
- **JSON output is the contract.** Table/CSV are for humans and can change formatting; JSON shape changes are breaking.

## Related repo files

- `README.md` — user-facing install / setup / usage. Update when adding subcommands or changing flags.
- `docs/flex-setup.md` — IBKR portal walkthrough for getting a token + query ID.
- `skills/ibkr-flex/SKILL.md` — Claude Code skill: analysis recipes, IBKR data quirks, common questions and how to answer them. Distributed via `npx skills add Azero-NG/ibkr-flex`.
- `skills/ibkr-flex/references/raw-xml.md` — escape-hatch recipes for sections not yet promoted to typed subcommands.
