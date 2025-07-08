# KV Storage - Modern Key-Value Storage with HTTP API

Современное key-value хранилище с HTTP API, построенное на Tarantool 3.4.0 с использованием Go и clean architecture.

## Особенности

- **Clean Architecture**
- **HTTP API**
- **Connection Pooling**
- **Rate Limiting**
- **Graceful Shutdown** 
- **Soft Delete**
- **Tarantool 3.4.0**
- **Logging** 

## Требования

- Go 1.21+
- Tarantool 3.4.0+
- Docker & Docker Compose (рекомендуется)

## Быстрый запуск через Docker Compose

Самый простой способ запустить проект:

```bash
# 1. Клонировать репозиторий
git clone <repository-url>
cd kv-storage

# 2. Запустить через Docker Compose
docker-compose up -d

# 3. Проверить статус
docker-compose ps

# 4. Посмотреть логи
docker-compose logs -f
```

### Доступные сервисы:
- **KV Storage API**: http://localhost:8080
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Tarantool**: localhost:3301

### Полезные команды Docker Compose:
```bash
# Остановить сервисы
docker-compose down

# Перезапустить
docker-compose restart

# Остановить и удалить volumes
docker-compose down -v

# Собрать заново
docker-compose up --build -d
```

## Управление проектом через Makefile

Проект включает Makefile для упрощения разработки и развертывания:

### Основные команды:
```bash
# Показать все доступные команды
make help

# Установка зависимостей и генерация Swagger
make install-deps

# Сборка приложения
make build

# Запуск в режиме разработки
make run

# Запуск тестов
make test

# Генерация Swagger документации
make swagger
```

## API Endpoints

### Swagger UI
- **URL**: http://localhost:8080/swagger/index.html

### Основные endpoints:

#### Создание записи
```bash
POST /api/v1/kv
Content-Type: application/json

{
  "key": "user:123",
  "value": "{\"name\":\"John Doe\",\"email\":\"john@example.com\"}"
}
```

#### Получение записи
```bash
GET /api/v1/kv/user:123
```

#### Обновление записи
```bash
PUT /api/v1/kv/user:123
Content-Type: application/json

{
  "value": "{\"name\":\"John Smith\",\"email\":\"john.smith@example.com\"}"
}
```

#### Удаление записей

##### Hard Delete (полное удаление)
```bash
DELETE /api/v1/kv/user:123
```

##### Soft Delete (мягкое удаление)
```bash
DELETE /api/v1/kv/user:123
Content-Type: application/json

{
  "soft_delete": true
}
```

#### Восстановление записи
```bash
POST /api/v1/kv/user:123/restore
```

#### Список записей
```bash
# Обычный список (без удаленных)
GET /api/v1/kv?limit=10&offset=0

# Список включая удаленные
GET /api/v1/kv/all?limit=10&offset=0
```

#### Health Check
```bash
GET /health
```

Система поддерживает два типа удаления:

### Hard Delete
- Полное удаление записи из хранилища
- Данные невозможно восстановить
- Подходит для кэша и временных данных

### Soft Delete
- Запись помечается как удаленная
- Данные остаются в хранилище
- Возможность восстановления
- Подходит для важных данных, аудита, пользователей

## Структура проекта

```
kv-storage/
├── cmd/
│   └── main.go                 # Точка входа
├── config/
│   └── config.yaml             # Конфигурация
├── docs/                       # Swagger документация
├── internal/
│   ├── app/
│   │   ├── bootstrap.go        # Инициализация приложения
│   │   └── logger.go           # Логгер
│   ├── config/
│   │   └── config.go           # Конфигурация
│   ├── domain/
│   │   ├── errors.go           # Ошибки домена
│   │   └── models.go           # Модели данных
│   ├── interfaces/
│   │   ├── logger.go           # Интерфейс логгера
│   │   ├── repository.go       # Интерфейс репозитория
│   │   ├── router.go           # Интерфейс роутера
│   │   └── service.go          # Интерфейс сервиса
│   ├── repository/
│   │   ├── pool.go             # Connection pooling
│   │   └── tarantool.go        # Tarantool репозиторий
│   ├── service/
│   │   ├── kv_service.go       # Бизнес-логика
│   │   └── kv_service_test.go  # Тесты сервиса
│   └── transport/
│       └── http/
│           ├── handler.go      # HTTP обработчики
│           ├── router.go       # HTTP роутер
│           └── middleware/
│               ├── logger.go   # Логирование
│               └── rate_limiter.go # Rate limiting
├── Dockerfile
├── docker-compose.yaml
├── go.mod
├── go.sum
├── init.lua                    # Tarantool инициализация
├── Makefile                    # Команды для управления проектом
└── README.md
```

## ⚙️ Конфигурация

### config/config.yaml
```yaml
app:
  name: "kv-storage"
  environment: "development"

http_server:
  port: "8080"
  read_timeout: 30s
  write_timeout: 30s

tarantool:
  host: "localhost"
  port: 3301
  username: "admin"
  password: "admin"
  timeout: 5s
```

#### Конфигурация в init.lua:
- **memtx_memory**: 1GB для хранения данных в памяти
- **checkpoint_interval**: 1 час для создания снапшотов
- **Автоматическая очистка**: старые soft-deleted записи удаляются через 30 дней

## Мониторинг

### Логи
Приложение использует структурированное логирование с Zap:
- **Development**: Цветной вывод в консоль
- **Production**: JSON формат

### Метрики
- HTTP запросы/ответы
- Latency
- Rate limiting статистика
- Connection pool статистика

## Сценарии использования

### Разработка
```bash
# 1. Клонировать проект
git clone <repository-url>
cd kv-storage

# 2. Настроить окружение
make dev-setup

# 3. Запустить
make dev
```

### Продакшн
```bash
# 1. Клонировать проект
git clone <repository-url>
cd kv-storage

# 2. Запустить через Docker
docker-compose up -d

# 3. Проверить статус
docker-compose ps
```

### Тестирование
```bash
# Запустить все тесты
make test-all

# Запустить только unit тесты
make test

# Запустить тесты с покрытием
make test-coverage
```
