package flex

type Trade struct {
	TradeID    string            `json:"tradeId"`
	TradeDate  string            `json:"tradeDate"`
	Symbol     string            `json:"symbol"`
	SecType    string            `json:"secType"`
	Currency   string            `json:"currency"`
	Exchange   string            `json:"exchange"`
	BuySell    string            `json:"buySell"`
	Quantity   float64           `json:"quantity"`
	TradePrice float64           `json:"tradePrice"`
	Commission float64           `json:"commission"`
	NetCash    float64           `json:"netCash"`
	Account    string            `json:"account"`
	Extra      map[string]string `json:"extra,omitempty"`
}

type Position struct {
	Symbol        string            `json:"symbol"`
	SecType       string            `json:"secType"`
	Currency      string            `json:"currency"`
	Position      float64           `json:"position"`
	MarkPrice     float64           `json:"markPrice"`
	PositionValue float64           `json:"positionValue"`
	CostBasis     float64           `json:"costBasis"`
	UnrealizedPnL float64           `json:"unrealizedPnl"`
	Account       string            `json:"account"`
	ReportDate    string            `json:"reportDate"`
	Extra         map[string]string `json:"extra,omitempty"`
}

type CashTx struct {
	Date        string            `json:"date"`
	Type        string            `json:"type"`
	Amount      float64           `json:"amount"`
	Currency    string            `json:"currency"`
	Description string            `json:"description"`
	Account     string            `json:"account"`
	Extra       map[string]string `json:"extra,omitempty"`
}

type Dividend struct {
	Date     string            `json:"date"`
	Symbol   string            `json:"symbol"`
	Amount   float64           `json:"amount"`
	Currency string            `json:"currency"`
	Account  string            `json:"account"`
	Extra    map[string]string `json:"extra,omitempty"`
}

type NAVEntry struct {
	Date     string            `json:"date"`
	Total    float64           `json:"total"`
	Currency string            `json:"currency"`
	Account  string            `json:"account"`
	Extra    map[string]string `json:"extra,omitempty"`
}

type MTMEntry struct {
	Date     string            `json:"date"`
	Symbol   string            `json:"symbol"`
	MTM      float64           `json:"mtm"`
	Currency string            `json:"currency"`
	Account  string            `json:"account"`
	Extra    map[string]string `json:"extra,omitempty"`
}

type Statement struct {
	Trades     []Trade
	Positions  []Position
	CashTxs    []CashTx
	Dividends  []Dividend
	NAVEntries []NAVEntry
	MTMEntries []MTMEntry
}
