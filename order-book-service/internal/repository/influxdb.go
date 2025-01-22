package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"order-book-service/internal/config"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type InfluxDB struct {
	client influxdb2.Client
	org    string
	bucket string
}

func NewInfluxDB(cfg config.InfluxDBSettings) (*InfluxDB, error) {
	client := influxdb2.NewClient(cfg.URL, cfg.Token)

	// Verify connection
	_, err := client.Ping(context.Background())
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to InfluxDB: %w", err)
	}

	return &InfluxDB{
		client: client,
		org:    cfg.Organization,
		bucket: cfg.Bucket,
	}, nil
}

func (db *InfluxDB) SaveOrderBook(ctx context.Context, data OrderBookData) error {
	writeAPI := db.client.WriteAPIBlocking(db.org, db.bucket)

	// Create points for bids
	for _, bid := range data.Bids {
		point := influxdb2.NewPoint(
			"order_book_bids",
			map[string]string{
				"exchange":   data.Exchange,
				"symbol":     data.Symbol,
				"is_private": fmt.Sprintf("%v", data.IsPrivate),
			},
			map[string]interface{}{
				"price":  bid[0],
				"amount": bid[1],
			},
			data.Timestamp,
		)

		if err := writeAPI.WritePoint(ctx, point); err != nil {
			return fmt.Errorf("failed to write bid point: %w", err)
		}
	}

	// Create points for asks
	for _, ask := range data.Asks {
		point := influxdb2.NewPoint(
			"order_book_asks",
			map[string]string{
				"exchange":   data.Exchange,
				"symbol":     data.Symbol,
				"is_private": fmt.Sprintf("%v", data.IsPrivate),
			},
			map[string]interface{}{
				"price":  ask[0],
				"amount": ask[1],
			},
			data.Timestamp,
		)

		if err := writeAPI.WritePoint(ctx, point); err != nil {
			return fmt.Errorf("failed to write ask point: %w", err)
		}
	}

	return nil
}

func (db *InfluxDB) Close() error {
	db.client.Close()
	return nil
}

func (db *InfluxDB) GetPairs(ctx context.Context, exchange string) ([]string, error) {
	query := fmt.Sprintf(`from(bucket:"%s")
		|> range(start: 0)
		|> filter(fn: (r) => r["_measurement"] == "exchange_pairs" and r["exchange"] == "%s")
		|> last()`, db.bucket, exchange)

	result, err := db.client.QueryAPI(db.org).Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	var pairs []string
	for result.Next() {
		record := result.Record()
		if field := record.Field(); field == "pairs" {
			err := json.Unmarshal([]byte(record.Value().(string)), &pairs)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal pairs: %w", err)
			}
			break
		}
	}

	return pairs, nil
}

func (db *InfluxDB) SavePairs(ctx context.Context, exchange string, pairs []string) error {
	pairsJSON, err := json.Marshal(pairs)
	if err != nil {
		return fmt.Errorf("failed to marshal pairs: %w", err)
	}

	point := influxdb2.NewPoint(
		"exchange_pairs",
		map[string]string{
			"exchange": exchange,
		},
		map[string]interface{}{
			"pairs": string(pairsJSON),
		},
		time.Now(),
	)

	writeAPI := db.client.WriteAPIBlocking(db.org, db.bucket)
	if err := writeAPI.WritePoint(ctx, point); err != nil {
		return fmt.Errorf("failed to write pairs: %w", err)
	}

	return nil
}
