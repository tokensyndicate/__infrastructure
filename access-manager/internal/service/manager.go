// internal/service/manager.go
package service

import (
	"context"
	"fmt"
	"time"

	"access-manager/internal/domain"
	"access-manager/internal/repository"
)

type AccessManager struct {
	vault *repository.VaultRepository
	redis *repository.RedisRepository
}

func NewAccessManager(vault *repository.VaultRepository, redis *repository.RedisRepository) *AccessManager {
	return &AccessManager{
		vault: vault,
		redis: redis,
	}
}

func (m *AccessManager) RegisterInstance(ctx context.Context, instance domain.Instance) error {
	if instance.ID == "" {
		return fmt.Errorf("instance ID is required")
	}
	if instance.ExternalIP == "" {
		return fmt.Errorf("external IP is required")
	}

	instance.CreatedAt = time.Now()
	instance.UpdatedAt = time.Now()

	if instance.Status == "" {
		instance.Status = domain.InstanceStatusActive
	}

	return m.vault.StoreInstance(ctx, instance)
}

func (m *AccessManager) GetInstance(ctx context.Context, id string) (*domain.Instance, error) {
	return m.vault.GetInstance(ctx, id)
}

func (m *AccessManager) selectOptimalInstance(ctx context.Context) (*domain.Instance, error) {
	instances, err := m.vault.ListInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	var selectedInstance *domain.Instance
	minLoad := int(^uint(0) >> 1) // Max int

	for i, instance := range instances {
		if instance.Status != domain.InstanceStatusActive {
			continue
		}

		// Get current metrics
		metrics, err := m.redis.GetInstanceMetrics(ctx, instance.ID)
		if err == nil && metrics != nil {
			instance.Load = metrics.Load
		}

		if instance.Load < minLoad {
			selectedInstance = &instances[i]
			minLoad = instance.Load
		}
	}

	if selectedInstance == nil {
		return nil, fmt.Errorf("no active instances available")
	}

	return selectedInstance, nil
}

func (m *AccessManager) AssignCredentials(ctx context.Context, creds domain.Credentials) (string, error) {
	if creds.ID == "" {
		return "", fmt.Errorf("credentials ID is required")
	}
	if creds.Exchange == "" {
		return "", fmt.Errorf("exchange is required")
	}

	var instance *domain.Instance
	var err error

	if creds.InstanceID == "" {
		instance, err = m.selectOptimalInstance(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to select instance: %w", err)
		}
		creds.InstanceID = instance.ID
	} else {
		instance, err = m.vault.GetInstance(ctx, creds.InstanceID)
		if err != nil {
			return "", fmt.Errorf("failed to get instance: %w", err)
		}
	}

	// Store credentials
	if err := m.vault.StoreCredentials(ctx, creds); err != nil {
		return "", fmt.Errorf("failed to store credentials: %w", err)
	}

	// Create and store binding
	binding := domain.CredentialsBinding{
		CredentialsID: creds.ID,
		InstanceID:    instance.ID,
		ExternalIP:    instance.ExternalIP,
		CreatedAt:     time.Now(),
	}

	if err := m.vault.StoreBinding(ctx, binding); err != nil {
		return "", fmt.Errorf("failed to store binding: %w", err)
	}

	// Notify instance about new credentials
	notification := domain.CredentialsNotification{
		Action:      "add",
		ExchangeID:  creds.Exchange,
		Credentials: creds,
	}

	if err := m.redis.NotifyInstanceUpdate(ctx, instance.ID, notification); err != nil {
		return "", fmt.Errorf("failed to notify instance: %w", err)
	}

	return instance.ExternalIP, nil
}

func (m *AccessManager) UpdateInstanceMetrics(ctx context.Context, metrics domain.InstanceMetrics) error {
	return m.redis.StoreInstanceMetrics(ctx, metrics)
}

func (m *AccessManager) GetInstanceBindings(ctx context.Context, instanceID string) ([]domain.CredentialsBinding, error) {
	return m.vault.GetBindingsForInstance(ctx, instanceID)
}

func (m *AccessManager) GetCredentials(ctx context.Context, credentialsID string) (*domain.Credentials, error) {
	return m.vault.GetCredentials(ctx, credentialsID)
}
