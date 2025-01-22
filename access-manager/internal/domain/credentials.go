package domain

import (
	"time"
)

type Credentials struct {
    ID            string            `json:"id"`
    Exchange      string            `json:"exchange"`
    ApiKey        string            `json:"api_key"`
    ApiSecret     string            `json:"api_secret"`
    IPRestriction bool              `json:"ip_restriction"`
    InstanceID    string            `json:"instance_id,omitempty"`
    Metadata      map[string]string `json:"metadata,omitempty"`
    CreatedAt     time.Time         `json:"created_at"`
    UpdatedAt     time.Time         `json:"updated_at"`
}

type CredentialsBinding struct {
    CredentialsID string            `json:"credentials_id"`
    InstanceID    string            `json:"instance_id"`
    ExternalIP    string            `json:"external_ip"`
    CreatedAt     time.Time         `json:"created_at"`
    Metadata      map[string]string `json:"metadata,omitempty"`
}

// For Redis notifications
type CredentialsNotification struct {
    Action      string     `json:"action"`
    ExchangeID  string     `json:"exchange_id"`
    Credentials Credentials `json:"credentials"`
}
