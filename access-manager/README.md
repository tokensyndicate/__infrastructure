# Access Manager Service

Access Manager is a microservice designed to manage and distribute API credentials across multiple instances of services that collect order book data from various exchanges. It provides centralized management of API keys, secrets, and instance assignments while ensuring proper IP restrictions and load balancing.

## Features

- Secure credential storage using HashiCorp Vault
- Real-time instance notification using Redis Pub/Sub
- Dynamic load balancing across instances
- IP restriction management for exchange API access
- Instance health monitoring
- RESTful API for service management

## Architecture

```plaintext
[API Gateway]
     │
     ├─► [Access Manager] ◄─► [Vault]
     │         │
     │         ▼
     │    [Redis Pub/Sub]
     │         │
     ▼         ▼
[Order Book Service Instances]
```

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- HashiCorp Vault
- Redis

## Configuration

Create a `config.yaml` file in the `config` directory:

```yaml
server:
  port: 8080
  host: "0.0.0.0"

vault:
  address: "http://vault:8200"
  token: "your-vault-token"

redis:
  address: "redis:6379"
  password: ""
  db: 0

logging:
  level: "info"
```

## API Endpoints

### Instance Management

#### Register a new instance

```http
POST /api/v1/instances
Content-Type: application/json

{
  "id": "instance-1",
  "external_ip": "1.2.3.4",
  "region": "us-east",
  "status": "active"
}
```

#### Get instance details

```http
GET /api/v1/instances/{id}
```

### Credentials Management

#### Assign credentials to instance

```http
POST /api/v1/credentials
Content-Type: application/json

{
  "id": "cred-1",
  "exchange": "binance",
  "api_key": "your-api-key",
  "api_secret": "your-api-secret",
  "ip_restriction": true
}
```

#### Get instance bindings

```http
GET /api/v1/instances/{id}/bindings
```

### Instance Metrics

#### Update instance metrics

```http
POST /api/v1/instances/{id}/metrics
Content-Type: application/json

{
  "instance_id": "instance-1",
  "load": 42,
  "connection_count": 100
}
```

## Project Structure

```plaintext
├── cmd/
│   └── server/
│       └── main.go               # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration handling
│   ├── domain/
│   │   ├── credentials.go       # Domain models for credentials
│   │   └── instance.go          # Domain models for instances
│   ├── ports/
│   │   └── http/
│   │       ├── handlers.go      # HTTP request handlers
│   │       └── server.go        # HTTP server setup
│   ├── repository/
│   │   ├── vault.go            # Vault repository implementation
│   │   └── redis.go            # Redis repository implementation
│   └── service/
│       └── manager.go          # Business logic implementation
├── go.mod
└── go.sum
```

## Running with Docker Compose

1. Clone the repository

```bash
git clone https://github.com/yourusername/access-manager.git
cd access-manager
```

2. Build and start the services

```bash
docker-compose up -d
```

3. Initialize Vault (first time only)

```bash
# Access Vault
docker exec -it vault sh

# Initialize Vault and note down the tokens
vault operator init

# Unseal Vault using the generated keys
vault operator unseal <key1>
vault operator unseal <key2>
vault operator unseal <key3>
```

## Development

1. Install dependencies

```bash
go mod tidy
```

2. Run tests

```bash
go test ./...
```

3. Run locally

```bash
go run cmd/server/main.go -config config.yaml
```

## Service Integration

Order Book Service instances should subscribe to Redis channels for credential updates:

```bash
instance:{instance-id}:credentials
```

Example notification message:

```json
{
  "action": "add",
  "exchange_id": "binance",
  "credentials": {
    "api_key": "key",
    "api_secret": "secret"
  }
}
```

## Security Considerations

- Vault is used for secure credential storage
- All sensitive data is encrypted at rest
- API keys are only transmitted through secure channels
- Instance authentication is required for accessing credentials
- IP restrictions are enforced for exchange API access

## Docker Environment

```yaml
version: "3.8"

services:
  access-manager:
    container_name: access-manager
    build:
      context: ./access-manager
      dockerfile: Dockerfile
    environment:
      - CONFIG_FILE=/app/config/config.yaml
    volumes:
      - ./access-manager/config:/app/config
    depends_on:
      - vault
      - redis
    networks:
      - backend-network

  vault:
    image: vault:latest
    container_name: vault
    ports:
      - "8200:8200"
    environment:
      - VAULT_ADDR=http://0.0.0.0:8200
      - VAULT_DEV_ROOT_TOKEN_ID=root
    cap_add:
      - IPC_LOCK
    volumes:
      - ./vault/config:/vault/config
      - ./vault/data:/vault/data
      - ./vault/logs:/vault/logs
    command: server
    networks:
      - backend-network

  redis:
    image: redis:alpine
    container_name: redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - backend-network
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 3

volumes:
  redis_data:
    driver: local

networks:
  backend-network:
    name: backend-network
    driver: bridge
```

## Monitoring

The service exposes metrics for monitoring:

- Instance health
- Load distribution
- Credential assignment statistics
- Redis pub/sub performance

## Error Handling

The service implements comprehensive error handling:

- Vault connection errors
- Redis pub/sub failures
- Invalid request validation
- Instance availability issues
- Configuration errors

Error responses follow a standard format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": {}
  }
}
```
