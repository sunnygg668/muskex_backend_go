package mproto

type data struct {
	EventType             string  `json:"e"`
	EventTime             int64   `json:"E"`
	Symbol                string  `json:"s"`
	PriceChange           string  `json:"p"`
	PriceChangePercent    string  `json:"P"`
	WeightedAvgPrice      string  `json:"w"`
	FirstTradePrice       string  `json:"x"`
	LastPrice             string  `json:"c"`
	LastQty               string  `json:"Q"`
	BestBidPrice          string  `json:"b"`
	BestBidQty            string  `json:"B"`
	BestAskPrice          string  `json:"a"`
	BestAskQty            string  `json:"A"`
	OpenPrice             string  `json:"o"`
	HighPrice             string  `json:"h"`
	LowPrice              string  `json:"l"`
	TotalTradedBaseAsset  string  `json:"v"`
	TotalTradedQuoteAsset string  `json:"q"`
	StatisticsOpenTime    int64   `json:"O"`
	StatisticsCloseTime   int64   `json:"C"`
	FirstTradeID          int64   `json:"F"`
	LastTradeID           int64   `json:"L"`
	TradeCount            int64   `json:"n"`
	Usd                   float64 `json:"usd"`
}

type ticker struct {
	Stream string `json:"stream"`
	Data   data   `json:"data"`
}
