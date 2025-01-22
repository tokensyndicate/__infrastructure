package service

import "order-book-service/internal/exchange"

type ExchangeAccess struct {
	ApiKey     string `json:"api_key"`
	ApiSecret  string `json:"api_secret"`
	ExchangeID string `json:"exchange_id"`
}

type CredentialsUpdate struct {
	Action      string         `json:"action"` // "add", "remove", "update"
	InstanceID  string         `json:"instance_id"`
	ExchangeID  string         `json:"exchange_id"`
	Credentials ExchangeAccess `json:"credentials"`
}

type ModuleInstance struct {
	Module      exchange.Module
	InstanceID  string
	ExchangeID  string
	Credentials ExchangeAccess
}
