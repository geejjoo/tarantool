# Примеры использования KV Storage API

## Быстрый старт

### 1. Запуск приложения

```bash
# Через Docker Compose (рекомендуется)
docker-compose up -d

# Или локально (требует Go и Tarantool)
go run cmd/main.go
```

### 2. Проверка работоспособности

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:
```json
{
  "status": "ok",
  "service": "kv-storage"
}
```

## Примеры API запросов

### Создание записи

```bash
curl -X POST http://localhost:8080/api/v1/kv \
  -H "Content-Type: application/json" \
  -d '{
    "key": "user:123",
    "value": {
      "name": "John Doe",
      "email": "john@example.com",
      "age": 30,
      "active": true
    }
  }'
```

Ожидаемый ответ (201 Created):
```json
{
  "key": "user:123",
  "value": {
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30,
    "active": true
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Получение записи

```bash
curl http://localhost:8080/api/v1/kv/user:123
```

Ожидаемый ответ (200 OK):
```json
{
  "key": "user:123",
  "value": {
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30,
    "active": true
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Обновление записи

```bash
curl -X PUT http://localhost:8080/api/v1/kv/user:123 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Smith",
    "email": "john.smith@example.com",
    "age": 31,
    "active": true,
    "last_login": "2024-01-15T10:35:00Z"
  }'
```

Ожидаемый ответ (200 OK):
```json
{
  "key": "user:123",
  "value": {
    "name": "John Smith",
    "email": "john.smith@example.com",
    "age": 31,
    "active": true,
    "last_login": "2024-01-15T10:35:00Z"
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:35:00Z"
}
```

### Удаление записи

```bash
curl -X DELETE http://localhost:8080/api/v1/kv/user:123
```

Ожидаемый ответ (200 OK):
```json
{
  "key": "user:123",
  "value": {
    "name": "John Smith",
    "email": "john.smith@example.com",
    "age": 31,
    "active": true,
    "last_login": "2024-01-15T10:35:00Z"
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:35:00Z"
}
```

### Получение списка записей

```bash
# Получить первые 10 записей
curl "http://localhost:8080/api/v1/kv?limit=10&offset=0"

# Получить следующие 10 записей
curl "http://localhost:8080/api/v1/kv?limit=10&offset=10"
```

Ожидаемый ответ (200 OK):
```json
{
  "items": [
    {
      "key": "user:123",
      "value": {
        "name": "John Doe",
        "email": "john@example.com"
      },
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    },
    {
      "key": "user:456",
      "value": {
        "name": "Jane Smith",
        "email": "jane@example.com"
      },
      "created_at": "2024-01-15T11:00:00Z",
      "updated_at": "2024-01-15T11:00:00Z"
    }
  ],
  "total": 2,
  "limit": 10,
  "offset": 0
}
```

## Примеры ошибок

### Ключ не найден (404)

```bash
curl http://localhost:8080/api/v1/kv/non-existing-key
```

Ожидаемый ответ:
```json
{
  "error": "Key not found"
}
```

### Ключ уже существует (409)

```bash
# Создаем запись
curl -X POST http://localhost:8080/api/v1/kv \
  -H "Content-Type: application/json" \
  -d '{"key": "test", "value": {"data": "first"}}'

# Пытаемся создать запись с тем же ключом
curl -X POST http://localhost:8080/api/v1/kv \
  -H "Content-Type: application/json" \
  -d '{"key": "test", "value": {"data": "second"}}'
```

Ожидаемый ответ:
```json
{
  "error": "Key already exists"
}
```

### Неверный JSON (400)

```bash
curl -X POST http://localhost:8080/api/v1/kv \
  -H "Content-Type: application/json" \
  -d '{"invalid": json}'
```

Ожидаемый ответ:
```json
{
  "error": "Invalid JSON format"
}
```

## Примеры использования в коде

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type CreateKVRequest struct {
    Key   string                 `json:"key"`
    Value string                 `json:"value"`
}

type KVResponse struct {
    Key       string                 `json:"key"`
    Value     string                 `json:"value"`
    CreatedAt string                 `json:"created_at"`
    UpdatedAt string                 `json:"updated_at"`
}

func main() {
    // Создание записи
    createReq := CreateKVRequest{
        Key: "user:123",
        Value: map[string]interface{}{
            "name":  "John Doe",
            "email": "john@example.com",
        },
    }

    jsonData, _ := json.Marshal(createReq)
    resp, _ := http.Post("http://localhost:8080/api/v1/kv", 
                        "application/json", bytes.NewBuffer(jsonData))
    defer resp.Body.Close()

    // Получение записи
    getResp, _ := http.Get("http://localhost:8080/api/v1/kv/user:123")
    defer getResp.Body.Close()

    var kv KVResponse
    json.NewDecoder(getResp.Body).Decode(&kv)
    fmt.Printf("Retrieved: %+v\n", kv)
}
```

## Swagger документация

После запуска приложения документация доступна по адресу:
http://localhost:8080/swagger/index.html

Здесь можно интерактивно тестировать все API эндпоинты. 