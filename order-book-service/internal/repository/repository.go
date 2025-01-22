package repository

import (
	"context"
	"time"
)

type OrderBookData struct {
	Exchange  string
	Symbol    string
	Timestamp time.Time
	Bids      [][2]float64
	Asks      [][2]float64
	IsPrivate bool
}

type ExchangeConfig struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Pairs     []string `json:"pairs"`
	IsEnabled bool     `json:"is_enabled"`
}

// Add these methods to your Repository interface
type Repository interface {
	SaveOrderBook(ctx context.Context, data OrderBookData) error

	// Settings management
	GetPairs(ctx context.Context, exchange string) ([]string, error)
	SavePairs(ctx context.Context, exchange string, pairs []string) error
	Close() error
}
