package exchange

import (
	"fmt"
	"order-book-service/internal/repository"
)

type ModuleFactory func(config ModuleConfig, repo repository.Repository) Module

var moduleFactories = make(map[string]ModuleFactory)

func RegisterModuleFactory(exchangeID string, factory ModuleFactory) {
    moduleFactories[exchangeID] = factory
}

func CreateModule(exchangeID string, config ModuleConfig, repo repository.Repository) (Module, error) {
    factory, exists := moduleFactories[exchangeID]
    if !exists {
        return nil, fmt.Errorf("no module factory registered for exchange: %s", exchangeID)
    }
    return factory(config, repo), nil
}
