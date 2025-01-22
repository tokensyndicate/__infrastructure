package exchange

import "context"

type ModuleConfig struct {
	Exchange  string            `json:"exchange"`
	ApiKey    string            `json:"api_key,omitempty"`
	ApiSecret string            `json:"api_secret,omitempty"`
	Extra     map[string]string `json:"extra,omitempty"`
}

type Module interface {
	Start(ctx context.Context) error
	Stop() error
	IsPrivate() bool
	GetPairs() []string
}
