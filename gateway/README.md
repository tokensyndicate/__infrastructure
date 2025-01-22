# Nginx API Gateway with JWT Authentication

This project implements an API Gateway using Nginx (OpenResty) with JWT authentication and rate limiting capabilities. It provides a secure way to route traffic to your microservices while protecting private endpoints with JWT tokens.

## Features

- JWT authentication for private endpoints
- Rate limiting
- Request routing
- Health check endpoint
- Modular configuration
- Easy service scaling
- Docker support

## Project Structure

```
/
├── Dockerfile
├── docker-compose.yaml
├── nginx.conf
├── jwt_secret.key
└── conf.d/
    ├── lua_settings.conf
    ├── rate_limits.conf
    ├── upstreams.conf
    ├── proxy_settings.conf
    ├── jwt_auth.conf
    └── locations/
        ├── public_endpoints.conf
        ├── private_endpoints.conf
        └── error_handlers.conf
```

## Prerequisites

- Docker
- Docker Compose

## Installation

1. Clone the repository:

```bash
git clone <repository-url>
cd <project-directory>
```

2. Create a JWT secret key:

```bash
echo "your-secret-key" > jwt_secret.key
```

3. Build and run the containers:

```bash
docker-compose up --build
```

## Configuration

### Adding New Services

1. Add your service to `docker-compose.yaml`:

```yaml
services:
  your-service:
    image: your-service-image
    networks:
      - backend-network
```

2. Add an upstream in `conf.d/upstreams.conf`:

```nginx
upstream your-service {
    server your-service:80;
}
```

3. Add a location in either `conf.d/locations/public_endpoints.conf` or `conf.d/locations/private_endpoints.conf`:

```nginx
location /private/your-service/ {
    limit_req zone=private_api burst=100 nodelay;
    limit_req zone=per_ip burst=50;
    limit_req_status 429;
    error_page 429 = @too_many_requests;

    include /etc/nginx/conf.d/jwt_auth.conf;

    proxy_pass http://your-service/;
    include /etc/nginx/conf.d/proxy_settings.conf;
}
```

### Rate Limiting Configuration

Rate limits can be adjusted in `conf.d/rate_limits.conf`:

```nginx
limit_req_zone $binary_remote_addr zone=public_api:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=private_api:10m rate=50r/s;
limit_req_zone $binary_remote_addr zone=per_ip:10m rate=100r/s;
```

## Usage

### Public Endpoints

Access public endpoints without authentication:

```bash
curl http://localhost/public/your-endpoint
```

### Private Endpoints

Access private endpoints with JWT token:

```bash
curl http://localhost/private/your-endpoint \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Health Check

Check the gateway health:

```bash
curl http://localhost/health
```

## JWT Token Generation

You can generate JWT tokens using [jwt.io](https://jwt.io). Make sure to:

1. Use the same secret key as in your `jwt_secret.key` file
2. Select HS256 as the algorithm
3. Include any required claims in the payload

Example payload:

```json
{
  "sub": "1234567890",
  "name": "John Doe",
  "iat": 1516239022
}
```

## Error Responses

### Authentication Errors

- Missing token:

```json
{ "error": "Missing Authorization header" }
```

- Invalid token:

```json
{ "error": "Invalid token", "reason": "Verification failed" }
```

### Rate Limiting Errors

When rate limit is exceeded:

```json
{ "error": "Too many requests. Please try again later." }
```

## Headers

The gateway forwards the following headers to your services:

- `X-Real-IP`
- `X-Forwarded-For`
- `X-Forwarded-Proto`
- All headers starting with `X-JWT-` (containing JWT claims)

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
