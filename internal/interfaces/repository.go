package interfaces

import "kv-storage/internal/domain"

type KVRepository interface {
	Create(kv *domain.KV) error
	Get(key string) (*domain.KV, error)
	Update(kv *domain.KV) error
	Delete(key string) (*domain.KV, error)
	SoftDelete(key string) error
	Restore(key string) (*domain.KV, error)
	List(limit, offset int) ([]*domain.KV, int, error)
	ListIncludingDeleted(limit, offset int) ([]*domain.KV, int, error)
	Close() error
}
