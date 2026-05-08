package flex

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
)

func Parse(body []byte) (*Statement, error) {
	dec := xml.NewDecoder(bytes.NewReader(body))
	stmt := &Statement{}
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("flex: parse: %w", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch start.Name.Local {
		case "Trade":
			stmt.Trades = append(stmt.Trades, decodeTrade(newExtractor(start)))
		case "OpenPosition":
			stmt.Positions = append(stmt.Positions, decodePosition(newExtractor(start)))
		case "CashTransaction":
			stmt.CashTxs = append(stmt.CashTxs, decodeCashTx(newExtractor(start)))
		case "ChangeInDividendAccrual":
			stmt.Dividends = append(stmt.Dividends, decodeDividend(newExtractor(start)))
		case "EquitySummaryByReportDateInBase":
			stmt.NAVEntries = append(stmt.NAVEntries, decodeNAV(newExtractor(start)))
		case "MTMPerformanceSummaryUnderlying":
			stmt.MTMEntries = append(stmt.MTMEntries, decodeMTM(newExtractor(start)))
		}
	}
	return stmt, nil
}

// AccountIDs scans a parsed statement and returns the union of accountId values across all sections.
func (s *Statement) AccountIDs() []string {
	seen := map[string]struct{}{}
	add := func(id string) {
		if id != "" {
			seen[id] = struct{}{}
		}
	}
	for _, t := range s.Trades {
		add(t.Account)
	}
	for _, p := range s.Positions {
		add(p.Account)
	}
	for _, c := range s.CashTxs {
		add(c.Account)
	}
	for _, d := range s.Dividends {
		add(d.Account)
	}
	for _, n := range s.NAVEntries {
		add(n.Account)
	}
	for _, m := range s.MTMEntries {
		add(m.Account)
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	return out
}

type extractor struct {
	attrs map[string]string
}

func newExtractor(start xml.StartElement) *extractor {
	m := make(map[string]string, len(start.Attr))
	for _, a := range start.Attr {
		m[a.Name.Local] = a.Value
	}
	return &extractor{attrs: m}
}

func (e *extractor) take(keys ...string) string {
	var result string
	for _, k := range keys {
		v, ok := e.attrs[k]
		if !ok {
			continue
		}
		delete(e.attrs, k)
		if result == "" {
			result = v
		}
	}
	return result
}

func (e *extractor) takeFloat(keys ...string) float64 {
	return parseFloat(e.take(keys...))
}

func (e *extractor) takeDate(keys ...string) string {
	return formatDate(e.take(keys...))
}

func (e *extractor) extra() map[string]string {
	if len(e.attrs) == 0 {
		return nil
	}
	return e.attrs
}

func decodeTrade(e *extractor) Trade {
	return Trade{
		TradeID:    e.take("tradeID"),
		TradeDate:  e.takeDate("tradeDate"),
		Symbol:     e.take("symbol"),
		SecType:    e.take("assetCategory", "secType"),
		Currency:   e.take("currency"),
		Exchange:   e.take("exchange"),
		BuySell:    e.take("buySell"),
		Quantity:   e.takeFloat("quantity"),
		TradePrice: e.takeFloat("tradePrice"),
		Commission: e.takeFloat("ibCommission"),
		NetCash:    e.takeFloat("netCash"),
		Account:    e.take("accountId"),
		Extra:      e.extra(),
	}
}

func decodePosition(e *extractor) Position {
	return Position{
		Symbol:        e.take("symbol"),
		SecType:       e.take("assetCategory", "secType"),
		Currency:      e.take("currency"),
		Position:      e.takeFloat("position"),
		MarkPrice:     e.takeFloat("markPrice"),
		PositionValue: e.takeFloat("positionValue"),
		CostBasis:     e.takeFloat("costBasisMoney", "costBasisPrice"),
		UnrealizedPnL: e.takeFloat("fifoPnlUnrealized", "unrealizedPnL"),
		Account:       e.take("accountId"),
		ReportDate:    e.takeDate("reportDate"),
		Extra:         e.extra(),
	}
}

func decodeCashTx(e *extractor) CashTx {
	return CashTx{
		Date:        e.takeDate("dateTime", "settleDate", "reportDate"),
		Type:        e.take("type"),
		Amount:      e.takeFloat("amount"),
		Currency:    e.take("currency"),
		Description: e.take("description"),
		Account:     e.take("accountId"),
		Extra:       e.extra(),
	}
}

func decodeDividend(e *extractor) Dividend {
	return Dividend{
		Date:     e.takeDate("payDate", "exDate", "reportDate"),
		Symbol:   e.take("symbol"),
		Amount:   e.takeFloat("grossAmount", "netAmount"),
		Currency: e.take("currency"),
		Account:  e.take("accountId"),
		Extra:    e.extra(),
	}
}

func decodeNAV(e *extractor) NAVEntry {
	return NAVEntry{
		Date:     e.takeDate("reportDate"),
		Total:    e.takeFloat("total"),
		Currency: e.take("currency"),
		Account:  e.take("accountId"),
		Extra:    e.extra(),
	}
}

func decodeMTM(e *extractor) MTMEntry {
	return MTMEntry{
		Date:     e.takeDate("reportDate"),
		Symbol:   e.take("symbol"),
		MTM:      e.takeFloat("mtm", "total"),
		Currency: e.take("currency"),
		Account:  e.take("accountId"),
		Extra:    e.extra(),
	}
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func formatDate(s string) string {
	if len(s) == 8 && allDigits(s) {
		return s[0:4] + "-" + s[4:6] + "-" + s[6:8]
	}
	// IBKR sometimes emits "YYYYMMDD;HHMMSS" for dateTime fields
	if len(s) == 15 && s[8] == ';' && allDigits(s[:8]) && allDigits(s[9:]) {
		return s[0:4] + "-" + s[4:6] + "-" + s[6:8] + "T" + s[9:11] + ":" + s[11:13] + ":" + s[13:15]
	}
	return s
}

func allDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
