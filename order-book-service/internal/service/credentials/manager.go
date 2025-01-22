package credentials

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/go-redis/redis/v8"
)

type ExchangeAccess struct {
    ApiKey    string `json:"api_key"`
    ApiSecret string `json:"api_secret"`
    ExchangeID string `json:"exchange_id"`
}

type CredentialsUpdate struct {
    Action      string         `json:"action"`
    ExchangeID  string         `json:"exchange_id"`
    Credentials ExchangeAccess `json:"credentials"`
}

type Manager struct {
    accessManagerURL string
    redisClient     *redis.Client
    instanceID      string
    credentials     map[string]ExchangeAccess
    mutex           sync.RWMutex
    updateCallback  func(string, ExchangeAccess)
}

func NewManager(accessManagerURL string, redisClient *redis.Client, instanceID string, callback func(string, ExchangeAccess)) *Manager {
    return &Manager{
        accessManagerURL: accessManagerURL,
        redisClient:     redisClient,
        instanceID:      instanceID,
        credentials:     make(map[string]ExchangeAccess),
        updateCallback:  callback,
    }
}

func (m *Manager) Start(ctx context.Context) error {
    if err := m.restoreCredentials(ctx); err != nil {
        log.Printf("Failed to restore credentials: %v", err)
    }

    go m.subscribeToUpdates(ctx)
    return nil
}

func (m *Manager) restoreCredentials(ctx context.Context) error {
    url := fmt.Sprintf("%s/api/v1/instances/%s/credentials", m.accessManagerURL, m.instanceID)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    var credentials []ExchangeAccess
    if err := json.NewDecoder(resp.Body).Decode(&credentials); err != nil {
        return err
    }

    m.mutex.Lock()
    defer m.mutex.Unlock()

    for _, cred := range credentials {
        m.credentials[cred.ExchangeID] = cred
        if m.updateCallback != nil {
            m.updateCallback(cred.ExchangeID, cred)
        }
    }

    return nil
}

func (m *Manager) subscribeToUpdates(ctx context.Context) {
    channel := fmt.Sprintf("instance:%s:credentials", m.instanceID)
    pubsub := m.redisClient.Subscribe(ctx, channel)
    defer pubsub.Close()

    for {
        select {
        case <-ctx.Done():
            return
        case msg := <-pubsub.Channel():
            var update CredentialsUpdate
            if err := json.Unmarshal([]byte(msg.Payload), &update); err != nil {
                log.Printf("Failed to unmarshal credentials update: %v", err)
                continue
            }

            m.handleUpdate(update)
        }
    }
}

func (m *Manager) handleUpdate(update CredentialsUpdate) {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    switch update.Action {
    case "add", "update":
        m.credentials[update.ExchangeID] = update.Credentials
        if m.updateCallback != nil {
            m.updateCallback(update.ExchangeID, update.Credentials)
        }
    case "remove":
        delete(m.credentials, update.ExchangeID)
        if m.updateCallback != nil {
            m.updateCallback(update.ExchangeID, ExchangeAccess{})
        }
    }
}

func (m *Manager) GetCredentials(exchangeID string) (ExchangeAccess, bool) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()

    cred, exists := m.credentials[exchangeID]
    return cred, exists
}
