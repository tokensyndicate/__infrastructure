package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"order-book-service/internal/config"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type TimescaleDB struct {
	pool *pgxpool.Pool
}

func NewTimescaleDB(cfg config.DatabaseSettings) (*TimescaleDB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	return &TimescaleDB{pool: pool}, nil
}

func (db *TimescaleDB) SaveOrderBook(ctx context.Context, data OrderBookData) error {
	log.Println("Saving order book data...")

	// Implementation for saving order book data
	return nil
}

func (db *TimescaleDB) Close() error {
	db.pool.Close()
	return nil
}

func (db *TimescaleDB) initTables(ctx context.Context) error {
	_, err := db.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS exchange_pairs (
			exchange TEXT PRIMARY KEY,
			pairs JSONB NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`)
	return err
}

func (db *TimescaleDB) GetPairs(ctx context.Context, exchange string) ([]string, error) {
	var pairsJSON []byte
	err := db.pool.QueryRow(ctx,
		"SELECT pairs FROM exchange_pairs WHERE exchange = $1",
		exchange,
	).Scan(&pairsJSON)

	if err == pgx.ErrNoRows {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pairs: %w", err)
	}

	var pairs []string
	if err := json.Unmarshal(pairsJSON, &pairs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pairs: %w", err)
	}

	return pairs, nil
}

func (db *TimescaleDB) SavePairs(ctx context.Context, exchange string, pairs []string) error {
	pairsJSON, err := json.Marshal(pairs)
	if err != nil {
		return fmt.Errorf("failed to marshal pairs: %w", err)
	}

	_, err = db.pool.Exec(ctx,
		`INSERT INTO exchange_pairs (exchange, pairs)
		VALUES ($1, $2)
		ON CONFLICT (exchange) DO UPDATE
		SET pairs = EXCLUDED.pairs,
			updated_at = NOW()`,
		exchange, pairsJSON)

	if err != nil {
		return fmt.Errorf("failed to save pairs: %w", err)
	}

	return nil
}
