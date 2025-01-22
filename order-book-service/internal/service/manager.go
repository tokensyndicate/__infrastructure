package service

import (
	"context"
	"fmt"
	"log"
	"order-book-service/internal/config"
	"order-book-service/internal/exchange"
	"order-book-service/internal/repository"
	"sync"
)

type Manager struct {
	config    config.Config
	repo      repository.Repository
	instances map[string]*ModuleInstance // InstanceID -> ModuleInstance
	mutex     sync.RWMutex
}

func NewManager(cfg config.Config, repo repository.Repository) *Manager {
	return &Manager{
		config:    cfg,
		repo:      repo,
		instances: make(map[string]*ModuleInstance),
	}
}

func (m *Manager) StartInstance(ctx context.Context, instanceID string, credentials ExchangeAccess) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	log.Printf("[Manager] Starting new instance %s for exchange %s", instanceID, credentials.ExchangeID)

	// Проверяем, не запущен ли уже инстанс с таким ID
	if _, exists := m.instances[instanceID]; exists {
		return fmt.Errorf("instance %s already exists", instanceID)
	}

	// Создаем конфигурацию модуля
	moduleConfig := exchange.ModuleConfig{
		Exchange:  credentials.ExchangeID,
		ApiKey:    credentials.ApiKey,
		ApiSecret: credentials.ApiSecret,
	}

	// Создаем модуль напрямую
	module, err := exchange.CreateModule(credentials.ExchangeID, moduleConfig, m.repo)
	if err != nil {
		return fmt.Errorf("failed to create new module: %w", err)
	}

	// Запускаем модуль
	if err := module.Start(ctx); err != nil {
		return fmt.Errorf("failed to start module: %w", err)
	}

	// Сохраняем инстанс
	m.instances[instanceID] = &ModuleInstance{
		Module:      module,
		InstanceID:  instanceID,
		ExchangeID:  credentials.ExchangeID,
		Credentials: credentials,
	}

	log.Printf("[Manager] Successfully started instance %s", instanceID)
	return nil
}

func (m *Manager) StopInstance(instanceID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	log.Printf("[Manager] Stopping instance %s", instanceID)

	instance, exists := m.instances[instanceID]
	if !exists {
		return fmt.Errorf("instance %s not found", instanceID)
	}

	// Останавливаем модуль
	if err := instance.Module.Stop(); err != nil {
		log.Printf("[Manager] Error stopping instance %s: %v", instanceID, err)
	}

	// Удаляем из списка инстансов
	delete(m.instances, instanceID)

	log.Printf("[Manager] Successfully stopped instance %s", instanceID)
	return nil
}

func (m *Manager) UpdateInstanceCredentials(ctx context.Context, instanceID string, credentials ExchangeAccess) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	log.Printf("[Manager] Updating credentials for instance %s", instanceID)

	// Получаем старый инстанс
	oldInstance, exists := m.instances[instanceID]
	if !exists {
		return fmt.Errorf("instance %s not found", instanceID)
	}

	// Останавливаем старый инстанс
	if err := oldInstance.Module.Stop(); err != nil {
		log.Printf("[Manager] Error stopping old instance %s: %v", instanceID, err)
	}

	// Создаем новую конфигурацию
	moduleConfig := exchange.ModuleConfig{
		Exchange:  oldInstance.ExchangeID,
		ApiKey:    credentials.ApiKey,
		ApiSecret: credentials.ApiSecret,
	}

	newModule, err := exchange.CreateModule(credentials.ExchangeID, moduleConfig, m.repo)
	if err != nil {
		return fmt.Errorf("failed to create new module: %w", err)
	}

	// Запускаем новый модуль
	if err := newModule.Start(ctx); err != nil {
		return fmt.Errorf("failed to start new module: %w", err)
	}

	// Обновляем инстанс
	m.instances[instanceID] = &ModuleInstance{
		Module:      newModule,
		InstanceID:  instanceID,
		ExchangeID:  oldInstance.ExchangeID,
		Credentials: credentials,
	}

	log.Printf("[Manager] Successfully updated instance %s", instanceID)
	return nil
}

// Обработчик обновлений от Access Manager
func (m *Manager) HandleCredentialsUpdate(ctx context.Context, update CredentialsUpdate) error {
	log.Printf("[Manager] Handling credentials update: action=%s, instanceID=%s",
		update.Action, update.InstanceID)

	switch update.Action {
	case "add":
		return m.StartInstance(ctx, update.InstanceID, update.Credentials)
	case "remove":
		return m.StopInstance(update.InstanceID)
	case "update":
		return m.UpdateInstanceCredentials(ctx, update.InstanceID, update.Credentials)
	default:
		return fmt.Errorf("unknown action: %s", update.Action)
	}
}

// Для совместимости с существующим кодом
func (m *Manager) Start(ctx context.Context) error {
	log.Printf("[Manager] Starting manager")

	if !m.config.AccessManager.Enabled {
		// Если Access Manager отключен, запускаем инстансы из конфига
		for exchangeID, exchangeCfg := range m.config.Exchanges {
			if !exchangeCfg.Enabled {
				continue
			}

			instanceID := fmt.Sprintf("%s-default", exchangeID)
			credentials := ExchangeAccess{
				ApiKey:    exchangeCfg.APIKey,
				ApiSecret: exchangeCfg.APISecret,
			}

			if err := m.StartInstance(ctx, instanceID, credentials); err != nil {
				log.Printf("[Manager] Error starting instance %s: %v", instanceID, err)
			}
		}
	}

	return nil
}

func (m *Manager) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var lastErr error
	for instanceID := range m.instances {
		if err := m.StopInstance(instanceID); err != nil {
			lastErr = err
			log.Printf("[Manager] Error stopping instance %s during shutdown: %v",
				instanceID, err)
		}
	}

	return lastErr
}
