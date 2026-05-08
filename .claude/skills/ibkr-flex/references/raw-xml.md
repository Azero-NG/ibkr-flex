# Raw XML — escape hatch and section catalog

This file is for cases where the user asks for data not exposed by any subcommand, AND the data lives in a Flex section we don't currently parse. The default flow is **promote the section to a typed subcommand**, not script around the XML.

## When to drop to XML at all

- One-off, exploratory question where adding code is overkill
- The section truly is rare and won't be asked again
- Debugging a parser issue (verifying which fields exist)

If the user is going to ask the same question more than once, **add a subcommand** (instructions at the bottom of this file).

## Cache file

```bash
XML="$HOME/Library/Caches/ibkr-flex/1234567-$(date +%Y%m%d).xml"
```

If today's cache is missing (no command run yet today), trigger a fetch first:

```bash
ibkr-flex accounts -v   # generates today's cache as a side effect
```

## Sections we currently parse

If the question involves any of these, use the typed subcommand instead:

| Section | Subcommand |
|---|---|
| `Trade` | `trades` |
| `OpenPosition` | `positions` |
| `CashTransaction` | `cash` |
| `ChangeInDividendAccrual` | `dividends` |
| `EquitySummaryByReportDateInBase` | `nav` |
| `MTMPerformanceSummaryUnderlying` | `mtm` |
| `ChangeInNAV` | `pnl` |

## Sections worth promoting (in priority order)

| Section | Why useful | Notes |
|---|---|---|
| `StatementOfFundsLine` | Full chronological ledger of every account event with running balance | Best for "what happened on day X" forensic |
| `OpenDividendAccrual` | Pending dividends declared but not yet paid | Useful for "what dividends are coming up" |
| `FxPositions` | Open foreign exchange positions | Useful when user holds multi-currency cash |
| `CorporateActions` | Splits, mergers, spin-offs, name changes | Forensic / position reconciliation |
| `Order` | Original orders that produced trades | Order-level (not execution-level) view |

## Generic extraction template

All Flex sections encode data as XML attributes on leaf elements. Stdlib regex works:

```python
import re

def extract(xml_path, leaf_element):
    """Returns list of dict, one per <leaf_element .../> instance."""
    with open(xml_path) as f:
        content = f.read()
    pattern = rf'<{leaf_element}\s+[^>]+/>'
    return [
        dict(re.findall(r'(\w+)="([^"]*)"', m.group()))
        for m in re.finditer(pattern, content)
    ]

# Examples:
ledger = extract(xml, 'StatementOfFundsLine')
fx_positions = extract(xml, 'FxPosition')
corporate_actions = extract(xml, 'CorporateAction')
```

To discover what fields exist in a section:

```python
records = extract(xml, 'StatementOfFundsLine')
all_fields = set()
for r in records:
    all_fields.update(r.keys())
print(sorted(all_fields))
```

## Date format normalization

IBKR uses two formats:

- `YYYYMMDD` (8 digits, e.g., `20260507`)
- `YYYYMMDD;HHMMSS` (15 chars with semicolon, e.g., `20260909;114417`)

```python
def to_iso(s):
    if len(s) == 8 and s.isdigit():
        return f'{s[:4]}-{s[4:6]}-{s[6:8]}'
    if len(s) == 15 and s[8] == ';':
        return f'{s[:4]}-{s[4:6]}-{s[6:8]}T{s[9:11]}:{s[11:13]}:{s[13:15]}'
    return s
```

## Adding a new typed subcommand

For each section pattern (`internal/flex` and `cmd/ibkr-flex`):

1. **`internal/flex/types.go`** — define a struct (core fields typed + `Extra map[string]string`):
   ```go
   type StatementLine struct {
       Date     string  `json:"date"`
       Type     string  `json:"type"`
       // ... core fields
       Extra    map[string]string `json:"extra,omitempty"`
   }
   ```
   Add the slice to `Statement` struct.

2. **`internal/flex/parse.go`** — add a `case "ElementName":` in `Parse()`'s switch, and a `decodeXxx` function using the `extractor` helper:
   ```go
   case "StatementOfFundsLine":
       stmt.Lines = append(stmt.Lines, decodeStatementLine(newExtractor(start)))
   ```

3. **`cmd/ibkr-flex/<section>.go`** — cobra command following the pattern of `trades.go` (apply `--account` and `--since/--until` filtering).

4. **`cmd/ibkr-flex/output.go`** — add `renderXxx()` covering `json` (always works via `renderJSON`), `table` (with `tabwriter`), `csv`.

5. **`cmd/ibkr-flex/root.go`** — wire the new command into the `AddCommand` block.

6. **Update `SKILL.md`** — add the new command to the subcommand table and a recipe section if the question is common.

Each new section is ~80 lines of code mirroring the existing seven. Discover field names by running `extract()` above on a real fetch, then pick the 8-15 most useful ones for typed access.
