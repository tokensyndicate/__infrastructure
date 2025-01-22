package bybit

import (
	"context"
	"encoding/json"
	"log"
	"order-book-service/internal/exchange"
	"order-book-service/internal/repository"
	"strconv"
	"sync"
	"time"
)

type Module struct {
	config  exchange.ModuleConfig
	repo    repository.Repository
	client  *BybitClient
	symbols []string
	mutex   sync.RWMutex
}

func NewModule(config exchange.ModuleConfig, repo repository.Repository) *Module {
	log.Printf("[Bybit Module] Initializing new module with config: %+v", config)
	module := &Module{
		config:  config,
		repo:    repo,
		client:  NewBybitClient(config),
		symbols: []string{},
	}
	log.Printf("[Bybit Module] Module initialized successfully")
	return module
}

func (m *Module) Start(ctx context.Context) error {
	log.Printf("[Bybit Module] Starting module")

	pairs, err := m.repo.GetPairs(ctx, "bybit")
	if err != nil {
		log.Printf("[Bybit Module] Failed to get pairs from repository: %v", err)
		return err
	}
	log.Printf("[Bybit Module] Retrieved %d pairs from repository", len(pairs))

	m.mutex.Lock()
	m.symbols = pairs
	m.mutex.Unlock()

	successfulConnections := 0
	for _, symbol := range m.symbols {
		log.Printf("[Bybit Module] Connecting to symbol: %s", symbol)

		if err := m.client.Connect(ctx, symbol); err != nil {
			log.Printf("[Bybit Module] Failed to connect to symbol %s: %v", symbol, err)
			return err
		}
		successfulConnections++
		log.Printf("[Bybit Module] Successfully connected to symbol: %s", symbol)

		m.client.AddHandler(symbol, func(data []byte) {
			startTime := time.Now()

			var bybitUpdate BybitOrderBookUpdate
			if err := json.Unmarshal(data, &bybitUpdate); err != nil {
				log.Printf("[Bybit Module] Failed to unmarshal order book data for %s: %v", symbol, err)
				return
			}

			// Convert and save each update in the message
			for _, update := range bybitUpdate.Data {
				orderBook, err := m.convertToOrderBookData(update, symbol)
				if err != nil {
					log.Printf("[Bybit Module] Failed to convert order book data for %s: %v", symbol, err)
					continue
				}

				log.Printf("[Bybit Module] Saving order book data for %s with %d bids and %d asks",
					symbol, len(orderBook.Bids), len(orderBook.Asks))

				if err := m.repo.SaveOrderBook(ctx, orderBook); err != nil {
					log.Printf("[Bybit Module] Failed to save order book for %s: %v", symbol, err)
					return
				}
			}

			processingTime := time.Since(startTime)
			if processingTime > 100*time.Millisecond {
				log.Printf("[Bybit Module] Warning: %s order book processing took %v",
					symbol, processingTime)
			} else {
				log.Printf("[Bybit Module] Successfully processed and saved order book for %s in %v",
					symbol, processingTime)
			}
		})

		log.Printf("[Bybit Module] Handler registered for symbol: %s", symbol)
	}

	log.Printf("[Bybit Module] Module started successfully. Connected to %d/%d symbols",
		successfulConnections, len(m.symbols))
	return nil
}

func (m *Module) convertToOrderBookData(update OrderBookData, symbol string) (repository.OrderBookData, error) {
	bids := make([][2]float64, len(update.Bids))
	asks := make([][2]float64, len(update.Asks))
	var conversionErrors int

	for i, bid := range update.Bids {
		if len(bid) != 2 {
			conversionErrors++
			continue
		}
		price, err := strconv.ParseFloat(bid[0], 64)
		if err != nil {
			conversionErrors++
			continue
		}
		quantity, err := strconv.ParseFloat(bid[1], 64)
		if err != nil {
			conversionErrors++
			continue
		}
		bids[i] = [2]float64{price, quantity}
	}

	for i, ask := range update.Asks {
		if len(ask) != 2 {
			conversionErrors++
			continue
		}
		price, err := strconv.ParseFloat(ask[0], 64)
		if err != nil {
			conversionErrors++
			continue
		}
		quantity, err := strconv.ParseFloat(ask[1], 64)
		if err != nil {
			conversionErrors++
			continue
		}
		asks[i] = [2]float64{price, quantity}
	}

	if conversionErrors > 0 {
		log.Printf("[Bybit Module] %s: Encountered %d conversion errors while processing order book",
			symbol, conversionErrors)
	}

	// Filter out zero quantities (deleted orders)
	filteredBids := make([][2]float64, 0, len(bids))
	for _, bid := range bids {
		if bid[1] > 0 {
			filteredBids = append(filteredBids, bid)
		}
	}

	filteredAsks := make([][2]float64, 0, len(asks))
	for _, ask := range asks {
		if ask[1] > 0 {
			filteredAsks = append(filteredAsks, ask)
		}
	}

	return repository.OrderBookData{
		Exchange:  "bybit",
		Symbol:    symbol,
		Timestamp: time.Now(),
		Bids:      filteredBids,
		Asks:      filteredAsks,
		IsPrivate: m.IsPrivate(),
	}, nil
}

func (m *Module) Stop() error {
	log.Printf("[Bybit Module] Stopping module")
	err := m.client.Close()
	if err != nil {
		log.Printf("[Bybit Module] Error stopping module: %v", err)
		return err
	}
	log.Printf("[Bybit Module] Module stopped successfully")
	return nil
}

func (m *Module) IsPrivate() bool {
	return m.config.ApiKey != "" && m.config.ApiSecret != ""
}

func (m *Module) GetPairs() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	pairs := make([]string, len(m.symbols))
	copy(pairs, m.symbols)
	return pairs
}
