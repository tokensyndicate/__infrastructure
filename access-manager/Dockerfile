FROM golang:1.22-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o access-manager ./cmd/server

FROM alpine:3.18

WORKDIR /app

COPY --from=0 /app/access-manager .

RUN mkdir -p /app/config

EXPOSE 8080

CMD ["./access-manager", "-config", "/app/config/config.yaml"]
