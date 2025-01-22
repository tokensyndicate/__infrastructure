// internal/repository/vault.go
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"access-manager/internal/domain"

	vault "github.com/hashicorp/vault/api"
)

type VaultRepository struct {
	client *vault.Client
}

func NewVaultRepository(client *vault.Client) *VaultRepository {
	return &VaultRepository{client: client}
}

// Instances
func (r *VaultRepository) StoreInstance(ctx context.Context, instance domain.Instance) error {
	instance.UpdatedAt = time.Now()
	if instance.CreatedAt.IsZero() {
		instance.CreatedAt = instance.UpdatedAt
	}

	data, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("failed to marshal instance: %w", err)
	}

	_, err = r.client.Logical().Write(fmt.Sprintf("instances/%s", instance.ID), map[string]interface{}{
		"data": string(data),
	})
	if err != nil {
		return fmt.Errorf("failed to store instance in vault: %w", err)
	}

	return nil
}

func (r *VaultRepository) GetInstance(ctx context.Context, id string) (*domain.Instance, error) {
	secret, err := r.client.Logical().Read(fmt.Sprintf("instances/%s", id))
	if err != nil {
		return nil, fmt.Errorf("failed to read instance from vault: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("instance not found")
	}

	data, ok := secret.Data["data"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	var instance domain.Instance
	if err := json.Unmarshal([]byte(data), &instance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal instance: %w", err)
	}

	return &instance, nil
}

func (r *VaultRepository) ListInstances(ctx context.Context) ([]domain.Instance, error) {
	secret, err := r.client.Logical().List("instances")
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return []domain.Instance{}, nil
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid keys format")
	}

	instances := make([]domain.Instance, 0, len(keys))
	for _, key := range keys {
		instanceID, ok := key.(string)
		if !ok {
			continue
		}

		instance, err := r.GetInstance(ctx, instanceID)
		if err != nil {
			continue
		}
		instances = append(instances, *instance)
	}

	return instances, nil
}

// Credentials
func (r *VaultRepository) StoreCredentials(ctx context.Context, creds domain.Credentials) error {
	creds.UpdatedAt = time.Now()
	if creds.CreatedAt.IsZero() {
		creds.CreatedAt = creds.UpdatedAt
	}

	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	_, err = r.client.Logical().Write(fmt.Sprintf("credentials/%s", creds.ID), map[string]interface{}{
		"data": string(data),
	})
	if err != nil {
		return fmt.Errorf("failed to store credentials in vault: %w", err)
	}

	return nil
}

func (r *VaultRepository) GetCredentials(ctx context.Context, id string) (*domain.Credentials, error) {
	secret, err := r.client.Logical().Read(fmt.Sprintf("credentials/%s", id))
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials from vault: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("credentials not found")
	}

	data, ok := secret.Data["data"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	var creds domain.Credentials
	if err := json.Unmarshal([]byte(data), &creds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return &creds, nil
}

// Bindings
func (r *VaultRepository) StoreBinding(ctx context.Context, binding domain.CredentialsBinding) error {
	if binding.CreatedAt.IsZero() {
		binding.CreatedAt = time.Now()
	}

	data, err := json.Marshal(binding)
	if err != nil {
		return fmt.Errorf("failed to marshal binding: %w", err)
	}

	_, err = r.client.Logical().Write(fmt.Sprintf("bindings/%s", binding.CredentialsID), map[string]interface{}{
		"data": string(data),
	})
	if err != nil {
		return fmt.Errorf("failed to store binding in vault: %w", err)
	}

	return nil
}

func (r *VaultRepository) GetBindingsForInstance(ctx context.Context, instanceID string) ([]domain.CredentialsBinding, error) {
	secret, err := r.client.Logical().List("bindings")
	if err != nil {
		return nil, fmt.Errorf("failed to list bindings: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return []domain.CredentialsBinding{}, nil
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid keys format")
	}

	bindings := make([]domain.CredentialsBinding, 0)
	for _, key := range keys {
		bindingID, ok := key.(string)
		if !ok {
			continue
		}

		secret, err := r.client.Logical().Read(fmt.Sprintf("bindings/%s", bindingID))
		if err != nil {
			continue
		}
		if secret == nil || secret.Data == nil {
			continue
		}

		data, ok := secret.Data["data"].(string)
		if !ok {
			continue
		}

		var binding domain.CredentialsBinding
		if err := json.Unmarshal([]byte(data), &binding); err != nil {
			continue
		}

		if binding.InstanceID == instanceID {
			bindings = append(bindings, binding)
		}
	}

	return bindings, nil
}
