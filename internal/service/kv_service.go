package service

import (
	"kv-storage/internal/domain"
	"kv-storage/internal/interfaces"
)

type KVService struct {
	repo   interfaces.KVRepository
	logger interfaces.Logger
}

func NewKVService(repo interfaces.KVRepository, logger interfaces.Logger) *KVService {
	return &KVService{
		repo:   repo,
		logger: logger,
	}
}

func (s *KVService) Create(req *domain.CreateKVRequest) (*domain.KV, error) {
	if req.Key == "" {
		return nil, domain.ErrInvalidKey
	}

	kv := &domain.KV{
		Key:   req.Key,
		Value: req.Value,
	}

	if err := s.repo.Create(kv); err != nil {
		return nil, err
	}

	return kv, nil
}

func (s *KVService) Get(key string) (*domain.KV, error) {
	if key == "" {
		return nil, domain.ErrInvalidKey
	}

	return s.repo.Get(key)
}

func (s *KVService) Update(key string, req *domain.UpdateKVRequest) (*domain.KV, error) {
	if key == "" {
		return nil, domain.ErrInvalidKey
	}

	kv := &domain.KV{
		Key:   key,
		Value: req.Value,
	}

	if err := s.repo.Update(kv); err != nil {
		return nil, err
	}

	return kv, nil
}

func (s *KVService) Delete(key string) (*domain.KV, error) {
	if key == "" {
		return nil, domain.ErrInvalidKey
	}

	return s.repo.Delete(key)
}

func (s *KVService) SoftDelete(key string) (*domain.KV, error) {
	if key == "" {
		return nil, domain.ErrInvalidKey
	}

	err := s.repo.SoftDelete(key)
	if err != nil {
		return nil, err
	}

	return s.repo.Get(key)
}

func (s *KVService) Restore(key string) (*domain.KV, error) {
	if key == "" {
		return nil, domain.ErrInvalidKey
	}

	_, err := s.repo.Restore(key)
	if err != nil {
		return nil, err
	}

	return s.repo.Get(key)
}

func (s *KVService) List(limit, offset int) (*domain.ListKVResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	items, total, err := s.repo.List(limit, offset)
	if err != nil {
		return nil, err
	}

	return &domain.ListKVResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *KVService) ListIncludingDeleted(limit, offset int) (*domain.ListKVResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	items, total, err := s.repo.ListIncludingDeleted(limit, offset)
	if err != nil {
		return nil, err
	}

	return &domain.ListKVResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}
