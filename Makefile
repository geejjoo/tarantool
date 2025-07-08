.PHONY: help build run test test-coverage clean docker-build docker-run docker-stop docker-logs swagger install-tarantool

# Переменные
BINARY_NAME=kv-storage
BUILD_DIR=bin
DOCKER_IMAGE=kv-storage
DOCKER_TAG=latest

help: ## Показать справку
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install-tarantool: ## Установить Tarantool 3.4.0
	@echo "Установка Tarantool 3.4.0..."
	@if [ "$(OS)" = "Windows_NT" ]; then \
		echo "Для Windows используйте WSL2 или скачайте с https://tarantool.io/download/"; \
	else \
		if command -v apt-get >/dev/null 2>&1; then \
			curl -L https://tarantool.io/installer.sh | bash; \
			sudo apt-get install tarantool; \
		elif command -v brew >/dev/null 2>&1; then \
			brew install tarantool; \
		else \
			echo "Неизвестная система. Установите Tarantool вручную."; \
		fi; \
	fi

install-deps: ## Установить зависимости
	@echo "Установка зависимостей..."
	go mod download
	go install github.com/swaggo/swag/cmd/swag@latest

build: ## Собрать приложение
	@echo "Сборка приложения..."
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/main.go

run: ## Запустить приложение в режиме разработки
	@echo "Запуск приложения..."
	go run ./cmd/main.go

test: ## Запустить unit тесты
	@echo "Запуск unit тестов..."
	go test -v ./...

test-coverage: ## Запустить тесты с покрытием
	@echo "Запуск тестов с покрытием..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Отчет о покрытии сохранен в coverage.html"

test-integration: ## Запустить integration тесты
	@echo "Запуск integration тестов..."
	go test -v ./tests/integration/

clean: ## Очистить сборки
	@echo "Очистка..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

swagger: ## Генерировать Swagger документацию
	@echo "Генерация Swagger документации..."
	swag init -g cmd/main.go -o docs

docker-build: ## Собрать Docker образ
	@echo "Сборка Docker образа..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: ## Запустить через Docker Compose
	@echo "Запуск через Docker Compose..."
	docker-compose up -d

docker-stop: ## Остановить Docker Compose
	@echo "Остановка Docker Compose..."
	docker-compose down

docker-logs: ## Показать логи Docker
	@echo "Логи Docker Compose..."
	docker-compose logs -f

docker-clean: ## Очистить Docker ресурсы
	@echo "Очистка Docker ресурсов..."
	docker-compose down -v
	docker system prune -f

tarantool-console: ## Подключиться к консоли Tarantool
	@echo "Подключение к консоли Tarantool..."
	tarantoolctl connect admin:admin@localhost:3301

tarantool-status: ## Проверить статус Tarantool
	@echo "Статус Tarantool..."
	tarantoolctl status

tarantool-logs: ## Показать логи Tarantool
	@echo "Логи Tarantool..."
	tarantoolctl logrotate
	tail -f /var/log/tarantool/tarantool.log

dev-setup: install-deps swagger ## Настройка окружения разработки
	@echo "Окружение разработки настроено!"

prod-setup: install-tarantool install-deps build ## Настройка продакшн окружения
	@echo "Продакшн окружение настроено!"

lint: ## Проверить код линтером
	@echo "Проверка кода..."
	golangci-lint run

fmt: ## Форматировать код
	@echo "Форматирование кода..."
	go fmt ./...

vet: ## Проверить код go vet
	@echo "Проверка go vet..."
	go vet ./...

security: ## Проверить безопасность зависимостей
	@echo "Проверка безопасности..."
	go list -json -deps ./... | nancy sleuth

benchmark: ## Запустить бенчмарки
	@echo "Запуск бенчмарков..."
	go test -bench=. -benchmem ./...

# Команды для разработки
dev: install-deps swagger run ## Полная настройка и запуск для разработки

# Команды для продакшна
prod: install-tarantool install-deps build ## Полная настройка для продакшна

# Команды для Docker
docker: docker-build docker-run ## Сборка и запуск Docker

# Команды для тестирования
test-all: test test-integration test-coverage ## Все тесты

# Команды для Tarantool
tarantool: tarantool-status ## Информация о Tarantool 