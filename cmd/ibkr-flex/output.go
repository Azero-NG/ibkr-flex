package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/azero/ibkr-flex/internal/flex"
)

func renderJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func newTableWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
}

func newCSVWriter() *csv.Writer {
	return csv.NewWriter(os.Stdout)
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func unknownFormatErr() error {
	return fmt.Errorf("unknown --format %q (expected json|table|csv)", globals.format)
}

func renderAccounts(ids []string) error {
	switch globals.format {
	case "json", "":
		return renderJSON(ids)
	case "table", "csv":
		for _, id := range ids {
			fmt.Println(id)
		}
		return nil
	default:
		return unknownFormatErr()
	}
}

func renderTrades(items []flex.Trade) error {
	switch globals.format {
	case "json", "":
		return renderJSON(items)
	case "table":
		w := newTableWriter()
		fmt.Fprintln(w, "TradeID\tDate\tSymbol\tType\tCcy\tSide\tQty\tPrice\tNetCash\tComm")
		for _, t := range items {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				t.TradeID, t.TradeDate, t.Symbol, t.SecType, t.Currency,
				t.BuySell, formatFloat(t.Quantity), formatFloat(t.TradePrice),
				formatFloat(t.NetCash), formatFloat(t.Commission))
		}
		return w.Flush()
	case "csv":
		w := newCSVWriter()
		w.Write([]string{"tradeId", "tradeDate", "symbol", "secType", "currency", "exchange", "buySell", "quantity", "tradePrice", "commission", "netCash", "account"})
		for _, t := range items {
			w.Write([]string{
				t.TradeID, t.TradeDate, t.Symbol, t.SecType, t.Currency, t.Exchange,
				t.BuySell, formatFloat(t.Quantity), formatFloat(t.TradePrice),
				formatFloat(t.Commission), formatFloat(t.NetCash), t.Account,
			})
		}
		w.Flush()
		return w.Error()
	default:
		return unknownFormatErr()
	}
}

func renderPositions(items []flex.Position) error {
	switch globals.format {
	case "json", "":
		return renderJSON(items)
	case "table":
		w := newTableWriter()
		fmt.Fprintln(w, "Symbol\tType\tCcy\tPos\tMarkPx\tMktVal\tCostBasis\tUnrealPnL\tDate")
		for _, p := range items {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				p.Symbol, p.SecType, p.Currency,
				formatFloat(p.Position), formatFloat(p.MarkPrice),
				formatFloat(p.PositionValue), formatFloat(p.CostBasis),
				formatFloat(p.UnrealizedPnL), p.ReportDate)
		}
		return w.Flush()
	case "csv":
		w := newCSVWriter()
		w.Write([]string{"symbol", "secType", "currency", "position", "markPrice", "positionValue", "costBasis", "unrealizedPnl", "account", "reportDate"})
		for _, p := range items {
			w.Write([]string{
				p.Symbol, p.SecType, p.Currency,
				formatFloat(p.Position), formatFloat(p.MarkPrice),
				formatFloat(p.PositionValue), formatFloat(p.CostBasis),
				formatFloat(p.UnrealizedPnL), p.Account, p.ReportDate,
			})
		}
		w.Flush()
		return w.Error()
	default:
		return unknownFormatErr()
	}
}

func renderCashTxs(items []flex.CashTx) error {
	switch globals.format {
	case "json", "":
		return renderJSON(items)
	case "table":
		w := newTableWriter()
		fmt.Fprintln(w, "Date\tType\tCcy\tAmount\tDescription")
		for _, c := range items {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				c.Date, c.Type, c.Currency, formatFloat(c.Amount), c.Description)
		}
		return w.Flush()
	case "csv":
		w := newCSVWriter()
		w.Write([]string{"date", "type", "currency", "amount", "description", "account"})
		for _, c := range items {
			w.Write([]string{c.Date, c.Type, c.Currency, formatFloat(c.Amount), c.Description, c.Account})
		}
		w.Flush()
		return w.Error()
	default:
		return unknownFormatErr()
	}
}

func renderDividends(items []flex.Dividend) error {
	switch globals.format {
	case "json", "":
		return renderJSON(items)
	case "table":
		w := newTableWriter()
		fmt.Fprintln(w, "Date\tSymbol\tCcy\tAmount")
		for _, d := range items {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", d.Date, d.Symbol, d.Currency, formatFloat(d.Amount))
		}
		return w.Flush()
	case "csv":
		w := newCSVWriter()
		w.Write([]string{"date", "symbol", "currency", "amount", "account"})
		for _, d := range items {
			w.Write([]string{d.Date, d.Symbol, d.Currency, formatFloat(d.Amount), d.Account})
		}
		w.Flush()
		return w.Error()
	default:
		return unknownFormatErr()
	}
}

func renderNAV(items []flex.NAVEntry) error {
	switch globals.format {
	case "json", "":
		return renderJSON(items)
	case "table":
		w := newTableWriter()
		fmt.Fprintln(w, "Date\tCcy\tTotal")
		for _, n := range items {
			fmt.Fprintf(w, "%s\t%s\t%s\n", n.Date, n.Currency, formatFloat(n.Total))
		}
		return w.Flush()
	case "csv":
		w := newCSVWriter()
		w.Write([]string{"date", "currency", "total", "account"})
		for _, n := range items {
			w.Write([]string{n.Date, n.Currency, formatFloat(n.Total), n.Account})
		}
		w.Flush()
		return w.Error()
	default:
		return unknownFormatErr()
	}
}

func renderMTM(items []flex.MTMEntry) error {
	switch globals.format {
	case "json", "":
		return renderJSON(items)
	case "table":
		w := newTableWriter()
		fmt.Fprintln(w, "Date\tSymbol\tCcy\tMTM")
		for _, m := range items {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.Date, m.Symbol, m.Currency, formatFloat(m.MTM))
		}
		return w.Flush()
	case "csv":
		w := newCSVWriter()
		w.Write([]string{"date", "symbol", "currency", "mtm", "account"})
		for _, m := range items {
			w.Write([]string{m.Date, m.Symbol, m.Currency, formatFloat(m.MTM), m.Account})
		}
		w.Flush()
		return w.Error()
	default:
		return unknownFormatErr()
	}
}
