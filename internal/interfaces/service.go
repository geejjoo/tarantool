package interfaces

import "kv-storage/internal/domain"

type KVService interface {
	Create(req *domain.CreateKVRequest) (*domain.KV, error)

	Get(key string) (*domain.KV, error)

	Update(key string, req *domain.UpdateKVRequest) (*domain.KV, error)

	Delete(key string) (*domain.KV, error)

	Restore(key string) (*domain.KV, error)

	List(limit, offset int) (*domain.ListKVResponse, error)

	ListIncludingDeleted(limit, offset int) (*domain.ListKVResponse, error)
}
