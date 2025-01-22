package bybit

// BybitOrderBookUpdate represents the format of Bybit WebSocket updates
type BybitOrderBookUpdate struct {
	Topic     string          `json:"topic"`
	Type      string          `json:"type"`
	Data      []OrderBookData `json:"data"`
	Timestamp int64           `json:"ts"`
}

type OrderBookData struct {
	Symbol string      `json:"s"`
	Bids   [][2]string `json:"b"`
	Asks   [][2]string `json:"a"`
	Update int64       `json:"u"`   // Update ID
	Seq    int64       `json:"seq"` // Cross sequence
}

// BybitAuthRequest represents authentication request for WebSocket
type BybitAuthRequest struct {
	Op   string `json:"op"`
	Args []any  `json:"args"`
}

// BybitSubscribeRequest represents subscription request for WebSocket
type BybitSubscribeRequest struct {
	Op   string   `json:"op"`
	Args []string `json:"args"`
}
