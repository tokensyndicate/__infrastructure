package bybit

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"order-book-service/internal/exchange"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	publicWebSocketURL  = "wss://stream.bybit.com/v5/public"
	privateWebSocketURL = "wss://stream.bybit.com/v5/private"
)

type BybitClient struct {
	config    exchange.ModuleConfig
	mu        sync.RWMutex
	done      chan struct{}
	handlers  map[string]func([]byte)
	conn      *websocket.Conn
	symbol    string
	isRunning bool
}

func NewBybitClient(config exchange.ModuleConfig) *BybitClient {
	log.Printf("[Bybit] Initializing new Bybit client with config: %+v", config)
	return &BybitClient{
		config:    config,
		handlers:  make(map[string]func([]byte)),
		isRunning: false,
	}
}

func (c *BybitClient) Connect(ctx context.Context, symbol string) error {
	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	if c.done != nil {
		close(c.done)
	}
	c.done = make(chan struct{})
	c.symbol = symbol
	c.isRunning = true
	c.mu.Unlock()

	log.Printf("[Bybit] Starting connection for symbol: %s", symbol)
	go c.connectionLoop(ctx)
	return nil
}

func (c *BybitClient) generateAuthParams() (string, string, string) {
	timestamp := time.Now().UnixMilli()
	expires := timestamp + 10000 // 10 seconds expiry

	// Concatenate timestamp and API key
	signStr := fmt.Sprintf("%d%s%d", timestamp, c.config.ApiKey, expires)

	// Generate HMAC signature
	h := hmac.New(sha256.New, []byte(c.config.ApiSecret))
	h.Write([]byte(signStr))
	signature := hex.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("%d", timestamp), fmt.Sprintf("%d", expires), signature
}

func (c *BybitClient) connectionLoop(ctx context.Context) {
	backoff := time.Second
	maxBackoff := time.Minute

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Bybit] Connection loop stopping due to context cancellation")
			return
		case <-c.done:
			log.Printf("[Bybit] Connection loop stopping due to client shutdown")
			return
		default:
			c.mu.RLock()
			if !c.isRunning {
				c.mu.RUnlock()
				return
			}
			symbol := c.symbol
			c.mu.RUnlock()

			err := c.connectAndRead(ctx, symbol)
			if err != nil {
				log.Printf("[Bybit] Error in connection/read loop: %v, reconnecting in %v", err, backoff)
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			}
		}
	}
}

func (c *BybitClient) connectAndRead(ctx context.Context, symbol string) error {
	var wsURL string
	if c.config.ApiKey != "" && c.config.ApiSecret != "" {
		wsURL = privateWebSocketURL
	} else {
		wsURL = publicWebSocketURL
	}

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return fmt.Errorf("websocket dial error: %w", err)
	}
	defer conn.Close()

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	// Authenticate if using private endpoint
	if c.config.ApiKey != "" && c.config.ApiSecret != "" {
		timestamp, expires, signature := c.generateAuthParams()
		authReq := BybitAuthRequest{
			Op: "auth",
			Args: []any{
				c.config.ApiKey,
				timestamp,
				expires,
				signature,
			},
		}

		if err := conn.WriteJSON(authReq); err != nil {
			return fmt.Errorf("authentication error: %w", err)
		}

		// Wait for auth response
		_, message, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed to read auth response: %w", err)
		}
		log.Printf("[Bybit] Authentication response: %s", string(message))
	}

	// Subscribe to orderbook
	formattedSymbol := strings.ToUpper(strings.ReplaceAll(symbol, "/", ""))
	topic := fmt.Sprintf("orderbook.50.%s", formattedSymbol)
	subscribeReq := BybitSubscribeRequest{
		Op:   "subscribe",
		Args: []string{topic},
	}

	if err := conn.WriteJSON(subscribeReq); err != nil {
		return fmt.Errorf("subscription error: %w", err)
	}

	log.Printf("[Bybit] WebSocket connection established for %s", symbol)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		case <-c.done:
			return fmt.Errorf("client shutdown")
		default:
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				return fmt.Errorf("websocket read error: %w", err)
			}

			if messageType != websocket.TextMessage {
				log.Printf("[Bybit] Unexpected message type: %d for symbol %s", messageType, symbol)
				continue
			}

			// Check if it's a pong message
			if strings.Contains(string(message), "pong") {
				continue
			}

			var update BybitOrderBookUpdate
			if err := json.Unmarshal(message, &update); err != nil {
				log.Printf("[Bybit] JSON unmarshal error: %v, raw message: %s", err, string(message))
				continue
			}

			// Process only if we have data
			if len(update.Data) > 0 {
				c.mu.RLock()
				handlers := make([]func([]byte), 0, len(c.handlers))
				for _, handler := range c.handlers {
					handlers = append(handlers, handler)
				}
				c.mu.RUnlock()

				for _, handler := range handlers {
					handler(message)
				}
			}
		}
	}
}

func (c *BybitClient) Close() error {
	log.Printf("[Bybit] Initiating client shutdown for symbol: %s", c.symbol)

	c.mu.Lock()
	c.isRunning = false
	if c.done != nil {
		close(c.done)
		c.done = nil
	}
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.mu.Unlock()

		if err != nil {
			return fmt.Errorf("websocket close error: %w", err)
		}
	} else {
		c.mu.Unlock()
	}

	log.Printf("[Bybit] Client shutdown complete")
	return nil
}

func (c *BybitClient) AddHandler(id string, handler func([]byte)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	log.Printf("[Bybit] Adding handler: %s", id)
	c.handlers[id] = handler
}

func (c *BybitClient) RemoveHandler(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	log.Printf("[Bybit] Removing handler: %s", id)
	delete(c.handlers, id)
}
