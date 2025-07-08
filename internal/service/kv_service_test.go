package service

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"kv-storage/internal/domain"
)

// MockRepository мок репозитория для тестирования
type MockRepository struct {
	store map[string]*domain.KV
	mu    sync.Mutex
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		store: make(map[string]*domain.KV),
		mu:    sync.Mutex{},
	}
}

func (m *MockRepository) Create(kv *domain.KV) error {
	if _, exists := m.store[kv.Key]; exists {
		return domain.ErrKeyExists
	}
	m.store[kv.Key] = kv
	return nil
}

func (m *MockRepository) Get(key string) (*domain.KV, error) {
	if kv, exists := m.store[key]; exists {
		return kv, nil
	}
	return nil, domain.ErrKeyNotFound
}

func (m *MockRepository) Update(kv *domain.KV) error {
	if _, exists := m.store[kv.Key]; !exists {
		return domain.ErrKeyNotFound
	}
	m.store[kv.Key] = kv
	return nil
}

func (m *MockRepository) Delete(key string) (*domain.KV, error) {
	if kv, exists := m.store[key]; exists {
		delete(m.store, key)
		return kv, nil
	}
	return nil, domain.ErrKeyNotFound
}

func (m *MockRepository) List(limit, offset int) ([]*domain.KV, int, error) {
	items := make([]*domain.KV, 0)
	count := 0

	for _, kv := range m.store {
		if count >= offset && len(items) < limit {
			items = append(items, kv)
		}
		count++
	}

	return items, len(m.store), nil
}

func (m *MockRepository) SoftDelete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	kv, exists := m.store[key]
	if !exists {
		return domain.ErrKeyNotFound
	}

	now := time.Now()
	kv.UpdatedAt = now
	kv.DeletedAt = &now
	kv.IsDeleted = true

	return nil
}

func (m *MockRepository) Restore(key string) (*domain.KV, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	kv, exists := m.store[key]
	if !exists {
		return nil, domain.ErrKeyNotFound
	}

	if !kv.IsDeleted {
		return nil, domain.ErrNotDeleted
	}

	now := time.Now()
	kv.UpdatedAt = now
	kv.DeletedAt = nil
	kv.IsDeleted = false

	return kv, nil
}

func (m *MockRepository) ListIncludingDeleted(limit, offset int) ([]*domain.KV, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	items := make([]*domain.KV, 0)
	count := 0

	for _, kv := range m.store {
		if count >= offset && len(items) < limit {
			items = append(items, kv)
		}
		count++
	}

	return items, len(m.store), nil
}

func (m *MockRepository) Close() error {
	return nil
}

// MockLogger мок логгера для тестирования
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (m *MockLogger) Info(msg string, keysAndValues ...interface{})  {}
func (m *MockLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {}
func (m *MockLogger) Fatal(msg string, keysAndValues ...interface{}) {}
func (m *MockLogger) Sync() error                                    { return nil }

func TestKVService_Create(t *testing.T) {
	repo := NewMockRepository()
	logger := &MockLogger{}
	service := NewKVService(repo, logger)

	tests := []struct {
		name    string
		req     *domain.CreateKVRequest
		wantErr error
	}{
		{
			name: "valid request",
			req: &domain.CreateKVRequest{
				Key:   "test-key",
				Value: map[string]interface{}{"test": "value"},
			},
			wantErr: nil,
		},
		{
			name: "empty key",
			req: &domain.CreateKVRequest{
				Key:   "",
				Value: map[string]interface{}{"test": "value"},
			},
			wantErr: domain.ErrInvalidKey,
		},
		{
			name: "empty value",
			req: &domain.CreateKVRequest{
				Key:   "test-key",
				Value: map[string]interface{}{},
			},
			wantErr: domain.ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Create(tt.req)
			if err != tt.wantErr {
				t.Errorf("KVService.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestKVService_Get(t *testing.T) {
	repo := NewMockRepository()
	logger := &MockLogger{}
	service := NewKVService(repo, logger)

	// Создаем тестовую запись
	testKV := &domain.KV{
		Key:   "test-key",
		Value: map[string]interface{}{"test": "value"},
	}
	err := repo.Create(testKV)
	if err != nil {
		fmt.Printf("create err: %v", err)
	}

	tests := []struct {
		name    string
		key     string
		wantErr error
	}{
		{
			name:    "existing key",
			key:     "test-key",
			wantErr: nil,
		},
		{
			name:    "non-existing key",
			key:     "non-existing",
			wantErr: domain.ErrKeyNotFound,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: domain.ErrInvalidKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Get(tt.key)
			if err != tt.wantErr {
				t.Errorf("KVService.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestKVService_Update(t *testing.T) {
	repo := NewMockRepository()
	logger := &MockLogger{}
	service := NewKVService(repo, logger)

	// Создаем тестовую запись
	testKV := &domain.KV{
		Key:   "test-key",
		Value: map[string]interface{}{"name": "old"},
	}
	repo.Create(testKV)

	tests := []struct {
		name    string
		key     string
		req     *domain.UpdateKVRequest
		wantErr error
	}{
		{
			name: "valid update",
			key:  "test-key",
			req: &domain.UpdateKVRequest{
				Value: map[string]interface{}{"name": "new"},
			},
			wantErr: nil,
		},
		{
			name: "non-existing key",
			key:  "non-existing",
			req: &domain.UpdateKVRequest{
				Value: map[string]interface{}{"name": "new"},
			},
			wantErr: domain.ErrKeyNotFound,
		},
		{
			name: "empty value",
			key:  "test-key",
			req: &domain.UpdateKVRequest{
				Value: map[string]interface{}{},
			},
			wantErr: domain.ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Update(tt.key, tt.req)
			if err != tt.wantErr {
				t.Errorf("KVService.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestKVService_Delete(t *testing.T) {
	repo := NewMockRepository()
	logger := &MockLogger{}
	service := NewKVService(repo, logger)

	// Создаем тестовую запись
	testKV := &domain.KV{
		Key:   "test-key",
		Value: map[string]interface{}{"test": "value"},
	}
	repo.Create(testKV)

	tests := []struct {
		name    string
		key     string
		wantErr error
	}{
		{
			name:    "existing key",
			key:     "test-key",
			wantErr: nil,
		},
		{
			name:    "non-existing key",
			key:     "non-existing",
			wantErr: domain.ErrKeyNotFound,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: domain.ErrInvalidKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Delete(tt.key)
			if err != tt.wantErr {
				t.Errorf("KVService.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
