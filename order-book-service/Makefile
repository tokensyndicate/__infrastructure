.PHONY: build run test docker-build up down

# Версия приложения
VERSION ?= 0.1.0

# Переменные для сборки
BINARY_NAME=order-book-service
DOCKER_IMAGE=order-book-service:$(VERSION)

# Сборка приложения
build:
	go build -o bin/$(BINARY_NAME) cmd/main.go

# Запуск приложения локально
run: build
	./bin/$(BINARY_NAME)

# Запуск тестов
test:
	go test -v ./...

# Сборка Docker образа
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Запуск всех сервисов через docker-compose
up:
	docker-compose up -d

# Остановка всех сервисов
down:
	docker-compose down

# Очистка
clean:
	rm -f bin/$(BINARY_NAME)
	go clean

# Установка зависимостей
deps:
	go mod download

# Обновление зависимостей
deps-update:
	go get -u ./...
	go mod tidy
