package binance

import (
	"order-book-service/internal/exchange"
	"order-book-service/internal/repository"
)

func init() {
    exchange.RegisterModuleFactory("binance", func(config exchange.ModuleConfig, repo repository.Repository) exchange.Module {
        return NewModule(config, repo)
    })
}
