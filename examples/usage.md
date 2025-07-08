# Примеры использования KV Storage API с Tarantool 3.4.0

## 🚀 Новые возможности Tarantool 3.4.0

### Производительность
- **До 30% быстрее** предыдущих версий
- **Улучшенная работа с JSON** - новые операторы и функции
- **Оптимизированные индексы** - поддержка функциональных индексов
- **Быстрая репликация** - более стабильная и эффективная

### Безопасность
- **Улучшенная аутентификация** - новые механизмы безопасности
- **Расширенное логирование** - детальная информация о событиях
- **Защита от атак** - встроенные механизмы защиты

### Мониторинг
- **Расширенная статистика** - больше метрик для мониторинга
- **Улучшенные логи** - структурированное логирование
- **Health checks** - встроенные проверки состояния

## Базовые операции

### 1. Создание записи

```bash
curl -X POST http://localhost:8080/api/v1/kv \
  -H "Content-Type: application/json" \
  -d '{
    "key": "user:123",
    "value": {
      "name": "John Doe",
      "email": "john@example.com",
      "age": 30,
      "active": true,
      "metadata": {
        "created_by": "admin",
        "tags": ["premium", "verified"]
      }
    }
  }'
```

**Ответ:**
```json
{
  "key": "user:123",
  "value": {
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30,
    "active": true,
    "metadata": {
      "created_by": "admin",
      "tags": ["premium", "verified"]
    }
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### 2. Получение записи

```bash
curl http://localhost:8080/api/v1/kv/user:123
```

**Ответ:**
```json
{
  "key": "user:123",
  "value": {
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30,
    "active": true,
    "metadata": {
      "created_by": "admin",
      "tags": ["premium", "verified"]
    }
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### 3. Обновление записи

```bash
curl -X PUT http://localhost:8080/api/v1/kv/user:123 \
  -H "Content-Type: application/json" \
  -d '{
    "value": {
      "name": "John Smith",
      "email": "john.smith@example.com",
      "age": 31,
      "active": true,
      "last_login": "2024-01-15T11:00:00Z",
      "metadata": {
        "created_by": "admin",
        "tags": ["premium", "verified", "active"],
        "updated_by": "system"
      }
    }
  }'
```

### 4. Удаление записей

#### Hard Delete (полное удаление)
```bash
curl -X DELETE http://localhost:8080/api/v1/kv/user:123
```

#### Soft Delete (мягкое удаление)
```bash
# Способ 1: Через специальный endpoint
curl -X DELETE http://localhost:8080/api/v1/kv/user:123/soft-delete

# Способ 2: Через основной DELETE с параметром
curl -X DELETE http://localhost:8080/api/v1/kv/user:123 \
  -H "Content-Type: application/json" \
  -d '{
    "soft_delete": true
  }'
```

**Ответ при soft delete:**
```json
{
  "key": "user:123",
  "value": {
    "name": "John Smith",
    "email": "john.smith@example.com",
    "age": 31,
    "active": true,
    "last_login": "2024-01-15T11:00:00Z",
    "metadata": {
      "created_by": "admin",
      "tags": ["premium", "verified", "active"],
      "updated_by": "system"
    }
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T11:00:00Z",
  "deleted_at": "2024-01-15T12:00:00Z",
  "is_deleted": true
}
```

### 5. Восстановление записи

```bash
curl -X POST http://localhost:8080/api/v1/kv/user:123/restore
```

**Ответ:**
```json
{
  "key": "user:123",
  "value": {
    "name": "John Smith",
    "email": "john.smith@example.com",
    "age": 31,
    "active": true,
    "last_login": "2024-01-15T11:00:00Z",
    "metadata": {
      "created_by": "admin",
      "tags": ["premium", "verified", "active"],
      "updated_by": "system"
    }
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T11:00:00Z"
}
```

### 6. Получение списка записей

#### Обычный список (без удаленных)
```bash
# Получить первые 10 записей
curl "http://localhost:8080/api/v1/kv?limit=10&offset=0"

# Получить следующие 10 записей
curl "http://localhost:8080/api/v1/kv?limit=10&offset=10"
```

#### Список включая удаленные
```bash
# Получить все записи включая soft-deleted
curl "http://localhost:8080/api/v1/kv/all?limit=10&offset=0"
```

**Ответ:**
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
      "updated_at": "2024-01-15T11:00:00Z",
      "deleted_at": "2024-01-15T12:00:00Z",
      "is_deleted": true
    }
  ],
  "total": 2,
  "limit": 10,
  "offset": 0
}
```

## Продвинутые примеры с Tarantool 3.4.0

### 1. Хранение конфигурации с версионированием

```bash
# Сохранение конфигурации приложения с версией
curl -X POST http://localhost:8080/api/v1/kv \
  -H "Content-Type: application/json" \
  -d '{
    "key": "config:app:v1.2.0",
    "value": {
      "version": "1.2.0",
      "database": {
        "host": "localhost",
        "port": 5432,
        "name": "myapp",
        "pool_size": 20,
        "timeout": "30s"
      },
      "redis": {
        "host": "localhost",
        "port": 6379,
        "db": 0,
        "password": null
      },
      "features": {
        "cache_enabled": true,
        "debug_mode": false,
        "rate_limiting": {
          "enabled": true,
          "requests_per_minute": 100
        }
      },
      "security": {
        "jwt_secret": "your-secret-key",
        "bcrypt_cost": 12,
        "session_timeout": "24h"
      }
    }
  }'
```

### 2. Хранение сессий пользователей с расширенными данными

```bash
# Создание сессии с детальной информацией
curl -X POST http://localhost:8080/api/v1/kv \
  -H "Content-Type: application/json" \
  -d '{
    "key": "session:abc123def456",
    "value": {
      "user_id": "user:123",
      "session_id": "abc123def456",
      "created_at": "2024-01-15T10:30:00Z",
      "expires_at": "2024-01-15T18:30:00Z",
      "last_activity": "2024-01-15T10:30:00Z",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
      "device_info": {
        "type": "desktop",
        "os": "Windows 10",
        "browser": "Chrome",
        "version": "120.0.0.0"
      },
      "permissions": ["read", "write", "admin"],
      "metadata": {
        "login_method": "password",
        "two_factor_enabled": true,
        "location": "Moscow, Russia"
      }
    }
  }'
```

### 3. Кэширование данных с TTL и метаданными

```bash
# Кэширование результата запроса с детальной информацией
curl -X POST http://localhost:8080/api/v1/kv \
  -H "Content-Type: application/json" \
  -d '{
    "key": "cache:users:list:2024-01-15",
    "value": {
      "data": [
        {
          "id": 1,
          "name": "John Doe",
          "email": "john@example.com",
          "status": "active",
          "last_login": "2024-01-15T09:00:00Z"
        },
        {
          "id": 2,
          "name": "Jane Smith",
          "email": "jane@example.com",
          "status": "active",
          "last_login": "2024-01-15T08:30:00Z"
        }
      ],
      "cached_at": "2024-01-15T10:30:00Z",
      "expires_at": "2024-01-15T11:30:00Z",
      "source": "database",
      "query_time": "15.2ms",
      "cache_hit_count": 0,
      "metadata": {
        "query": "SELECT * FROM users WHERE status = 'active'",
        "filters": {"status": "active"},
        "sort": {"last_login": "desc"},
        "limit": 100
      }
    }
  }'
```

### 4. Хранение метрик производительности

```bash
# Сохранение детальных метрик производительности
curl -X POST http://localhost:8080/api/v1/kv \
  -H "Content-Type: application/json" \
  -d '{
    "key": "metrics:api:requests:2024-01-15",
    "value": {
      "date": "2024-01-15",
      "total_requests": 1500,
      "successful_requests": 1450,
      "failed_requests": 50,
      "average_response_time": 120,
      "p95_response_time": 250,
      "p99_response_time": 500,
      "endpoints": {
        "GET /api/v1/kv": {
          "count": 800,
          "avg_time": 50,
          "errors": 5
        },
        "POST /api/v1/kv": {
          "count": 300,
          "avg_time": 150,
          "errors": 10
        },
        "PUT /api/v1/kv": {
          "count": 200,
          "avg_time": 120,
          "errors": 8
        },
        "DELETE /api/v1/kv": {
          "count": 200,
          "avg_time": 80,
          "errors": 12
        }
      },
      "error_codes": {
        "400": 15,
        "404": 20,
        "500": 15
      },
      "client_info": {
        "browsers": {
          "Chrome": 800,
          "Firefox": 400,
          "Safari": 200,
          "Edge": 100
        },
        "platforms": {
          "Windows": 900,
          "macOS": 300,
          "Linux": 200,
          "Mobile": 100
        }
      }
    }
  }'
```

### 5. Хранение событий аудита

```bash
# Сохранение событий аудита с детальной информацией
curl -X POST http://localhost:8080/api/v1/kv \
  -H "Content-Type: application/json" \
  -d '{
    "key": "audit:user:123:2024-01-15T10:30:00Z",
    "value": {
      "event_type": "user_login",
      "timestamp": "2024-01-15T10:30:00Z",
      "user_id": "user:123",
      "session_id": "abc123def456",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "location": {
        "country": "Russia",
        "city": "Moscow",
        "timezone": "Europe/Moscow"
      },
      "details": {
        "login_method": "password",
        "two_factor_used": true,
        "remember_me": false
      },
      "risk_score": 0.1,
      "metadata": {
        "request_id": "req-123456",
        "correlation_id": "corr-789012"
      }
    }
  }'
```

## Сценарии использования Soft Delete с Tarantool 3.4.0

### 1. Аудит и восстановление данных

```bash
# Создаем важный документ
curl -X POST http://localhost:8080/api/v1/kv \
  -H "Content-Type: application/json" \
  -d '{
    "key": "document:contract:123",
    "value": {
      "title": "Service Agreement",
      "content": "This is a legal document...",
      "version": "1.0",
      "author": "legal@company.com",
      "signatures": [
        {
          "name": "John Doe",
          "email": "john@company.com",
          "signed_at": "2024-01-15T10:00:00Z"
        }
      ],
      "metadata": {
        "document_type": "contract",
        "status": "active",
        "expires_at": "2025-01-15T00:00:00Z"
      }
    }
  }'

# "Удаляем" документ (soft delete)
curl -X DELETE http://localhost:8080/api/v1/kv/document:contract:123/soft-delete

# Позже восстанавливаем
curl -X POST http://localhost:8080/api/v1/kv/document:contract:123/restore
```

### 2. Временное отключение функций

```bash
# Отключаем функцию
curl -X DELETE http://localhost:8080/api/v1/kv/feature:beta-testing/soft-delete

# Включаем обратно
curl -X POST http://localhost:8080/api/v1/kv/feature:beta-testing/restore
```

### 3. Управление пользователями

```bash
# Деактивируем пользователя
curl -X DELETE http://localhost:8080/api/v1/kv/user:456/soft-delete

# Проверяем что пользователь не доступен
curl http://localhost:8080/api/v1/kv/user:456
# Ответ: {"error": "key not found"}

# Но можем восстановить
curl -X POST http://localhost:8080/api/v1/kv/user:456/restore
```

## Мониторинг и отладка с Tarantool 3.4.0

### 1. Health Check

```bash
curl http://localhost:8080/health
```

**Ответ:**
```json
{
  "status": "ok",
  "service": "kv-storage",
  "version": "1.0.0",
  "tarantool_version": "3.4.0"
}
```

### 2. Swagger документация

Откройте в браузере: `http://localhost:8080/swagger/index.html`

### 3. Логи приложения

Логи выводятся в консоль в структурированном формате:

```json
{
  "level": "info",
  "msg": "HTTP Request",
  "method": "POST",
  "path": "/api/v1/kv",
  "status": 201,
  "latency": "15.2ms",
  "client_ip": "192.168.1.100",
  "user_agent": "curl/7.68.0",
  "request_id": "req-123456"
}
```

### 4. Tarantool 3.4.0 мониторинг

```bash
# Подключение к консоли Tarantool
tarantoolctl connect admin:admin@localhost:3301

# Просмотр информации о системе
box.info()

# Просмотр статистики
box.stat()

# Просмотр пространств
box.space.kv:count()

# Просмотр индексов
box.space.kv.index

# Просмотр метрик производительности
box.stat.net()
box.stat.memtx()
```

## Использование с различными языками программирования

### Python

```python
import requests
import json
from datetime import datetime

class KVStorageClient:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url
        self.session = requests.Session()
    
    def create_kv(self, key, value):
        """Создание записи"""
        url = f"{self.base_url}/api/v1/kv"
        data = {"key": key, "value": value}
        response = self.session.post(url, json=data)
        response.raise_for_status()
        return response.json()
    
    def get_kv(self, key):
        """Получение записи"""
        url = f"{self.base_url}/api/v1/kv/{key}"
        response = self.session.get(url)
        response.raise_for_status()
        return response.json()
    
    def update_kv(self, key, value):
        """Обновление записи"""
        url = f"{self.base_url}/api/v1/kv/{key}"
        data = {"value": value}
        response = self.session.put(url, json=data)
        response.raise_for_status()
        return response.json()
    
    def soft_delete_kv(self, key):
        """Soft delete записи"""
        url = f"{self.base_url}/api/v1/kv/{key}/soft-delete"
        response = self.session.delete(url)
        response.raise_for_status()
        return response.json()
    
    def restore_kv(self, key):
        """Восстановление записи"""
        url = f"{self.base_url}/api/v1/kv/{key}/restore"
        response = self.session.post(url)
        response.raise_for_status()
        return response.json()
    
    def list_kv(self, limit=10, offset=0, include_deleted=False):
        """Получение списка записей"""
        endpoint = "/all" if include_deleted else ""
        url = f"{self.base_url}/api/v1/kv{endpoint}"
        params = {"limit": limit, "offset": offset}
        response = self.session.get(url, params=params)
        response.raise_for_status()
        return response.json()

# Пример использования
client = KVStorageClient()

# Создаем пользователя
user_data = {
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30,
    "metadata": {
        "created_by": "admin",
        "tags": ["premium"]
    }
}

try:
    # Создаем запись
    created = client.create_kv("user:123", user_data)
    print(f"Created: {created['key']}")
    
    # Получаем запись
    retrieved = client.get_kv("user:123")
    print(f"Retrieved: {retrieved['value']['name']}")
    
    # Soft delete
    deleted = client.soft_delete_kv("user:123")
    print(f"Soft deleted: {deleted['key']}")
    
    # Восстанавливаем
    restored = client.restore_kv("user:123")
    print(f"Restored: {restored['key']}")
    
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
```

### JavaScript/Node.js

```javascript
const axios = require('axios');

class KVStorageClient {
    constructor(baseURL = 'http://localhost:8080') {
        this.baseURL = baseURL;
        this.client = axios.create({
            baseURL,
            timeout: 10000,
            headers: {
                'Content-Type': 'application/json'
            }
        });
    }

    async createKV(key, value) {
        try {
            const response = await this.client.post('/api/v1/kv', { key, value });
            return response.data;
        } catch (error) {
            console.error('Error creating KV:', error.response?.data || error.message);
            throw error;
        }
    }

    async getKV(key) {
        try {
            const response = await this.client.get(`/api/v1/kv/${key}`);
            return response.data;
        } catch (error) {
            console.error('Error getting KV:', error.response?.data || error.message);
            throw error;
        }
    }

    async updateKV(key, value) {
        try {
            const response = await this.client.put(`/api/v1/kv/${key}`, { value });
            return response.data;
        } catch (error) {
            console.error('Error updating KV:', error.response?.data || error.message);
            throw error;
        }
    }

    async softDeleteKV(key) {
        try {
            const response = await this.client.delete(`/api/v1/kv/${key}/soft-delete`);
            return response.data;
        } catch (error) {
            console.error('Error soft deleting KV:', error.response?.data || error.message);
            throw error;
        }
    }

    async restoreKV(key) {
        try {
            const response = await this.client.post(`/api/v1/kv/${key}/restore`);
            return response.data;
        } catch (error) {
            console.error('Error restoring KV:', error.response?.data || error.message);
            throw error;
        }
    }

    async listKV(limit = 10, offset = 0, includeDeleted = false) {
        try {
            const endpoint = includeDeleted ? '/all' : '';
            const response = await this.client.get(`/api/v1/kv${endpoint}`, {
                params: { limit, offset }
            });
            return response.data;
        } catch (error) {
            console.error('Error listing KV:', error.response?.data || error.message);
            throw error;
        }
    }
}

// Пример использования
async function example() {
    const client = new KVStorageClient();

    const userData = {
        name: 'John Doe',
        email: 'john@example.com',
        age: 30,
        metadata: {
            created_by: 'admin',
            tags: ['premium']
        }
    };

    try {
        // Создаем запись
        const created = await client.createKV('user:123', userData);
        console.log('Created:', created.key);

        // Получаем запись
        const retrieved = await client.getKV('user:123');
        console.log('Retrieved:', retrieved.value.name);

        // Soft delete
        const deleted = await client.softDeleteKV('user:123');
        console.log('Soft deleted:', deleted.key);

        // Восстанавливаем
        const restored = await client.restoreKV('user:123');
        console.log('Restored:', restored.key);

    } catch (error) {
        console.error('Error:', error.message);
    }
}

example();
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type CreateKVRequest struct {
    Key   string                 `json:"key"`
    Value map[string]interface{} `json:"value"`
}

type UpdateKVRequest struct {
    Value map[string]interface{} `json:"value"`
}

type KV struct {
    Key       string                 `json:"key"`
    Value     map[string]interface{} `json:"value"`
    CreatedAt string                 `json:"created_at"`
    UpdatedAt string                 `json:"updated_at"`
    DeletedAt *string                `json:"deleted_at,omitempty"`
    IsDeleted bool                   `json:"is_deleted,omitempty"`
}

type ListKVResponse struct {
    Items  []*KV `json:"items"`
    Total  int   `json:"total"`
    Limit  int   `json:"limit"`
    Offset int   `json:"offset"`
}

type KVStorageClient struct {
    baseURL string
    client  *http.Client
}

func NewKVStorageClient(baseURL string) *KVStorageClient {
    return &KVStorageClient{
        baseURL: baseURL,
        client: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (c *KVStorageClient) CreateKV(key string, value map[string]interface{}) (*KV, error) {
    req := CreateKVRequest{
        Key:   key,
        Value: value,
    }
    
    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }
    
    resp, err := c.client.Post(c.baseURL+"/api/v1/kv", 
        "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var kv KV
    if err := json.NewDecoder(resp.Body).Decode(&kv); err != nil {
        return nil, err
    }
    
    return &kv, nil
}

func (c *KVStorageClient) GetKV(key string) (*KV, error) {
    resp, err := c.client.Get(c.baseURL + "/api/v1/kv/" + key)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var kv KV
    if err := json.NewDecoder(resp.Body).Decode(&kv); err != nil {
        return nil, err
    }
    
    return &kv, nil
}

func (c *KVStorageClient) UpdateKV(key string, value map[string]interface{}) (*KV, error) {
    req := UpdateKVRequest{Value: value}
    
    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }
    
    httpReq, err := http.NewRequest("PUT", 
        c.baseURL+"/api/v1/kv/"+key, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := c.client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var kv KV
    if err := json.NewDecoder(resp.Body).Decode(&kv); err != nil {
        return nil, err
    }
    
    return &kv, nil
}

func (c *KVStorageClient) SoftDeleteKV(key string) (*KV, error) {
    httpReq, err := http.NewRequest("DELETE", 
        c.baseURL+"/api/v1/kv/"+key+"/soft-delete", nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := c.client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var kv KV
    if err := json.NewDecoder(resp.Body).Decode(&kv); err != nil {
        return nil, err
    }
    
    return &kv, nil
}

func (c *KVStorageClient) RestoreKV(key string) (*KV, error) {
    resp, err := c.client.Post(c.baseURL+"/api/v1/kv/"+key+"/restore", 
        "application/json", nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var kv KV
    if err := json.NewDecoder(resp.Body).Decode(&kv); err != nil {
        return nil, err
    }
    
    return &kv, nil
}

func (c *KVStorageClient) ListKV(limit, offset int, includeDeleted bool) (*ListKVResponse, error) {
    endpoint := "/api/v1/kv"
    if includeDeleted {
        endpoint += "/all"
    }
    
    url := fmt.Sprintf("%s%s?limit=%d&offset=%d", c.baseURL, endpoint, limit, offset)
    resp, err := c.client.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var response ListKVResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, err
    }
    
    return &response, nil
}

func main() {
    client := NewKVStorageClient("http://localhost:8080")
    
    userData := map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
        "age":   30,
        "metadata": map[string]interface{}{
            "created_by": "admin",
            "tags":       []string{"premium"},
        },
    }
    
    // Создаем запись
    kv, err := client.CreateKV("user:123", userData)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Created: %s\n", kv.Key)
    
    // Получаем запись
    retrieved, err := client.GetKV("user:123")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Retrieved: %v\n", retrieved.Value["name"])
    
    // Soft delete
    deleted, err := client.SoftDeleteKV("user:123")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Soft deleted: %s\n", deleted.Key)
    
    // Восстанавливаем
    restored, err := client.RestoreKV("user:123")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Restored: %s\n", restored.Key)
}
```

## Производительность Tarantool 3.4.0

### Бенчмарки

```bash
# Тест производительности записи
ab -n 10000 -c 100 -p test_data.json -T application/json http://localhost:8080/api/v1/kv

# Тест производительности чтения
ab -n 10000 -c 100 http://localhost:8080/api/v1/kv/test-key

# Тест производительности обновления
ab -n 10000 -c 100 -u test_update.json -T application/json http://localhost:8080/api/v1/kv/test-key
```

### Ожидаемые результаты с Tarantool 3.4.0:

- **Запись**: ~50,000 ops/sec
- **Чтение**: ~100,000 ops/sec  
- **Обновление**: ~40,000 ops/sec
- **Latency**: < 1ms для большинства операций
- **Память**: Эффективное использование с автоматической очисткой

## Заключение

Tarantool 3.4.0 предоставляет значительные улучшения производительности и функциональности для KV Storage:

1. **Высокая производительность** - до 30% быстрее предыдущих версий
2. **Улучшенная работа с JSON** - новые операторы и функции
3. **Расширенная безопасность** - новые механизмы аутентификации
4. **Лучший мониторинг** - больше метрик и улучшенное логирование
5. **Автоматическая очистка** - старые soft-deleted записи удаляются автоматически
6. **Стабильная репликация** - более надежная синхронизация данных

Система готова для продакшн использования с поддержкой высоких нагрузок и требований к безопасности. 