# Vault Setup and Management Guide

## 1. Initial Setup

### Directory Structure

```
vault/
├── config/
│   ├── vault.hcl
│   └── policies/
│       ├── microservice.hcl
│       └── admin.hcl
├── data/
└── logs/
```

### Base Configuration (vault.hcl)

```hcl
storage "file" {
  path = "/vault/data"
}

listener "tcp" {
  address     = "0.0.0.0:8200"
  tls_disable = 1  # Enable TLS in production!
}

api_addr = "http://0.0.0.0:8200"

ui = true
```

### Initialize Vault

```bash
# Initialize vault
vault operator init -key-shares=5 -key-threshold=3 -format=json > init-keys.json

# Unseal vault (needs to be done after each restart)
vault operator unseal <key1>
vault operator unseal <key2>
vault operator unseal <key3>
```

## 2. Manage Permissions

### Create Policies (policies/microservice.hcl)

```hcl
# Read access to secrets
path "secret/data/users/*/keys" {
  capabilities = ["read"]
}

# Deny all other paths
path "*" {
  capabilities = ["deny"]
}
```

### Apply Policies

```bash
# Write policy
vault policy write microservice-policy ./config/policies/microservice.hcl

# Update policy (same command)
vault policy write microservice-policy ./config/policies/microservice.hcl

# List policies
vault policy list

# View policy
vault policy read microservice-policy
```

## 3. Update Configuration

### Update Vault Configuration

1. Modify configuration files in `./config/`
2. Reload Vault configuration:

```bash
vault operator reload
```

### Update Policies

1. Modify policy files in `./config/policies/`
2. Apply updated policies:

```bash
vault policy write policy-name ./config/policies/policy-file.hcl
```

## 4. Token Management

### Create Service Token

```bash
# Create token with policy
vault token create \
    -policy="microservice-policy" \
    -period="768h" \
    -format=json

# Create token with specific TTL
vault token create \
    -policy="microservice-policy" \
    -ttl="1h" \
    -explicit-max-ttl="4h"
```

### Token Rotation Setup

```bash
# Enable approle auth method
vault auth enable approle

# Create role with token rotation
vault write auth/approle/role/microservice \
    token_policies="microservice-policy" \
    token_ttl="1h" \
    token_max_ttl="4h" \
    token_num_uses=100

# Get credentials
vault read auth/approle/role/microservice/role-id
vault write -f auth/approle/role/microservice/secret-id \
    -format=json
```

### Periodic Token Rotation (cronjob example)

```bash
0 */4 * * * /usr/local/bin/vault write -f auth/approle/role/microservice/secret-id -format=json > /path/to/new/secret-id.json
```

## 5. Backup

### Raft Snapshot

```bash
# Create snapshot
vault operator raft snapshot save snapshot.gz

# Restore snapshot
vault operator raft snapshot restore snapshot.gz
```

### Automated Backup Script

```bash
#!/bin/bash
BACKUP_DIR="/path/to/backup"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup
vault operator raft snapshot save "${BACKUP_DIR}/vault_${DATE}.gz"

# Keep last 7 days of backups
find ${BACKUP_DIR} -name "vault_*.gz" -mtime +7 -delete
```

## 6. Golang Implementation Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    vault "github.com/hashicorp/vault/api"
)

type VaultClient struct {
    client *vault.Client
    ctx    context.Context
}

func NewVaultClient(ctx context.Context) (*VaultClient, error) {
    config := vault.DefaultConfig()

    // Configure from environment
    config.Address = os.Getenv("VAULT_ADDR")

    client, err := vault.NewClient(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create vault client: %w", err)
    }

    // AppRole authentication
    roleID := os.Getenv("VAULT_ROLE_ID")
    secretID := os.Getenv("VAULT_SECRET_ID")

    data := map[string]interface{}{
        "role_id":   roleID,
        "secret_id": secretID,
    }

    resp, err := client.Logical().Write("auth/approle/login", data)
    if err != nil {
        return nil, fmt.Errorf("failed to login: %w", err)
    }

    // Set the token
    client.SetToken(resp.Auth.ClientToken)

    return &VaultClient{
        client: client,
        ctx:    ctx,
    }, nil
}

func (v *VaultClient) GetUserKeys(userID string) (map[string]interface{}, error) {
    path := fmt.Sprintf("secret/data/users/%s/keys", userID)

    secret, err := v.client.Logical().Read(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read secret: %w", err)
    }

    if secret == nil || secret.Data == nil {
        return nil, fmt.Errorf("no secret found")
    }

    // KV v2 stores data in a nested map
    data, ok := secret.Data["data"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid secret format")
    }

    return data, nil
}

// Usage example
func main() {
    ctx := context.Background()

    client, err := NewVaultClient(ctx)
    if err != nil {
        log.Fatalf("Failed to create vault client: %v", err)
    }

    keys, err := client.GetUserKeys("user123")
    if err != nil {
        log.Fatalf("Failed to get keys: %v", err)
    }

    fmt.Printf("Keys: %v\n", keys)
}
```

### Environment Setup for Go Service

```bash
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_ROLE_ID='role-id-here'
export VAULT_SECRET_ID='secret-id-here'
```

## Important Notes

1. Security:

   - Enable TLS in production
   - Use proper access control
   - Implement rate limiting
   - Monitor audit logs

2. High Availability:

   - Use raft storage in production
   - Set up multiple nodes
   - Implement proper backup strategy

3. Maintenance:
   - Regularly rotate root tokens
   - Monitor token expiration
   - Keep policies updated
   - Regular security audits
