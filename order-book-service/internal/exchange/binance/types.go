package binance

// BinanceOrderBookUpdate представляет формат данных от Binance WebSocket
type BinanceOrderBookUpdate struct {
	EventType    string     `json:"e"`
	EventTime    int64      `json:"E"`
	Symbol       string     `json:"s"`
	FirstUpdate  int64      `json:"U"`
	FinalUpdate  int64      `json:"u"`
	Bids         [][]string `json:"b"`
	Asks         [][]string `json:"a"`
}
