package domain

import "time"

type InstanceStatus string

const (
	InstanceStatusActive   InstanceStatus = "active"
	InstanceStatusInactive InstanceStatus = "inactive"
	InstanceStatusError    InstanceStatus = "error"
)

type Instance struct {
	ID         string            `json:"id"`
	ExternalIP string            `json:"external_ip"`
	Region     string            `json:"region"`
	Status     InstanceStatus    `json:"status"`
	Load       int               `json:"load"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type InstanceMetrics struct {
	InstanceID      string    `json:"instance_id"`
	Load            int       `json:"load"`
	ConnectionCount int       `json:"connection_count"`
	UpdatedAt       time.Time `json:"updated_at"`
}
