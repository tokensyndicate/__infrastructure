package binance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"order-book-service/internal/exchange"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type BinanceClient struct {
	config    exchange.ModuleConfig
	mu        sync.RWMutex
	done      chan struct{}
	handlers  map[string]func([]byte)
	conn      *websocket.Conn
	symbol    string
	isRunning bool
}

func NewBinanceClient(config exchange.ModuleConfig) *BinanceClient {
	log.Printf("[Binance] Initializing new Binance client with config: %+v", config)
	return &BinanceClient{
		config:    config,
		handlers:  make(map[string]func([]byte)),
		isRunning: false,
	}
}

func (c *BinanceClient) Connect(ctx context.Context, symbol string) error {
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

	log.Printf("[Binance] Starting connection for symbol: %s", symbol)
	go c.connectionLoop(ctx)
	return nil
}

func (c *BinanceClient) connectionLoop(ctx context.Context) {
	backoff := time.Second
	maxBackoff := time.Minute

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Binance] Connection loop stopping due to context cancellation")
			return
		case <-c.done:
			log.Printf("[Binance] Connection loop stopping due to client shutdown")
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
				log.Printf("[Binance] Error in connection/read loop: %v, reconnecting in %v", err, backoff)
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

// generateSignature creates HMAC SHA256 signature for Binance authentication
func (c *BinanceClient) generateSignature(timestamp int64) string {
	if c.config.ApiSecret == "" {
		return ""
	}

	mac := hmac.New(sha256.New, []byte(c.config.ApiSecret))
	message := fmt.Sprintf("timestamp=%d", timestamp)
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func (c *BinanceClient) connectAndRead(ctx context.Context, symbol string) error {
	formattedSymbol := strings.ToLower(strings.ReplaceAll(symbol, "/", ""))

	var wsURL string
	if c.config.ApiKey != "" && c.config.ApiSecret != "" {
		// Private WebSocket endpoint
		timestamp := time.Now().UnixMilli()
		signature := c.generateSignature(timestamp)
		wsURL = fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@depth@100ms?timestamp=%d&signature=%s",
			formattedSymbol, timestamp, signature)
	} else {
		// Public WebSocket endpoint
		wsURL = fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@depth@100ms", formattedSymbol)
	}

	// Create custom header with API Key if available
	headers := http.Header{}
	if c.config.ApiKey != "" {
		headers.Add("X-MBX-APIKEY", c.config.ApiKey)
	}

	log.Printf("[Binance] Connecting to WebSocket URL: %s", wsURL)

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, headers)
	if err != nil {
		return fmt.Errorf("websocket dial error: %w", err)
	}
	defer conn.Close()

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	log.Printf("[Binance] WebSocket connection established for %s (private: %v)",
		symbol, c.config.ApiKey != "")

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
				log.Printf("[Binance] Unexpected message type: %d for symbol %s", messageType, symbol)
			}

			var update BinanceOrderBookUpdate
			if err := json.Unmarshal(message, &update); err != nil {
				log.Printf("[Binance] JSON unmarshal error: %v, raw message: %s", err, string(message))
				continue
			}

			// Проверяем наличие данных
			if len(update.Bids) > 0 || len(update.Asks) > 0 {
				log.Printf("[Binance] Received update for %s - Event: %s, Bids: %d, Asks: %d",
					symbol, update.EventType, len(update.Bids), len(update.Asks))

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

func (c *BinanceClient) Close() error {
	log.Printf("[Binance] Initiating client shutdown for symbol: %s", c.symbol)

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

	log.Printf("[Binance] Client shutdown complete")
	return nil
}

func (c *BinanceClient) AddHandler(id string, handler func([]byte)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	log.Printf("[Binance] Adding handler: %s", id)
	c.handlers[id] = handler
}

func (c *BinanceClient) RemoveHandler(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	log.Printf("[Binance] Removing handler: %s", id)
	delete(c.handlers, id)
}
