# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Установка зависимостей для сборки
RUN apk add --no-cache git make

# Копирование файлов проекта
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o order-book-service cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Копирование бинарного файла из build stage
COPY --from=builder /app/order-book-service .
COPY --from=builder /app/config /app/config

# Создание пользователя без прав root
RUN adduser -D appuser
USER appuser

CMD ["./order-book-service"]
