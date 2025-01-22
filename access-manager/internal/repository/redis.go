// internal/repository/redis.go
package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"access-manager/internal/domain"

	"github.com/go-redis/redis/v8"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

func (r *RedisRepository) NotifyInstanceUpdate(ctx context.Context, instanceID string, notification domain.CredentialsNotification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	channel := fmt.Sprintf("instance:%s:credentials", instanceID)
	if err := r.client.Publish(ctx, channel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish notification: %w", err)
	}

	return nil
}

func (r *RedisRepository) StoreInstanceMetrics(ctx context.Context, metrics domain.InstanceMetrics) error {
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	key := fmt.Sprintf("metrics:instance:%s", metrics.InstanceID)
	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to store metrics: %w", err)
	}

	return nil
}

func (r *RedisRepository) GetInstanceMetrics(ctx context.Context, instanceID string) (*domain.InstanceMetrics, error) {
	key := fmt.Sprintf("metrics:instance:%s", instanceID)
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("metrics not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	var metrics domain.InstanceMetrics
	if err := json.Unmarshal([]byte(data), &metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	return &metrics, nil
}
