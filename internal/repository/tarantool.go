package repository

import (
	"errors"
	"fmt"
	"github.com/tarantool/go-tarantool/v2"
	"go.uber.org/zap"
	"kv-storage/internal/config"
	"kv-storage/internal/domain"
	"kv-storage/internal/interfaces"
	"time"
)

type TarantoolRepository struct {
	pool   *ConnectionPool
	logger interfaces.Logger
	config *config.Config
}

func NewTarantoolRepository(cfg *config.Config, logger interfaces.Logger) (interfaces.KVRepository, error) {
	pool, err := NewConnectionPool(cfg, logger, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &TarantoolRepository{
		pool:   pool,
		logger: logger,
		config: cfg,
	}, nil
}

func (r *TarantoolRepository) Create(kv *domain.KV) error {
	now := time.Now().Unix()

	return r.pool.Execute(func(conn *tarantool.Connection) error {
		_, err := conn.Do(
			tarantool.NewInsertRequest("kv").Tuple([]interface{}{
				kv.Key,
				kv.Value,
				uint32(now),
				uint32(now),
				uint32(0),
				false,
			}),
		).Get()

		if err != nil {
			if terr, ok := err.(tarantool.Error); ok && terr.Code == 3 {
				r.logger.Warn("Key already exists", "key", kv.Key)
				return domain.ErrKeyAlreadyExists
			}
			r.logger.Error("Failed to create KV record", "key", kv.Key, "error", err)
			return domain.ErrDatabaseError
		}

		kv.CreatedAt = time.Unix(now, 0)
		kv.UpdatedAt = time.Unix(now, 0)
		kv.DeletedAt = nil
		kv.IsDeleted = false

		r.logger.Info("KV record created", "key", kv.Key)
		return nil
	})
}

func (r *TarantoolRepository) Get(key string) (*domain.KV, error) {
	var result []interface{}

	err := r.pool.Execute(func(conn *tarantool.Connection) error {
		resp, err := conn.Do(
			tarantool.NewSelectRequest("kv").Limit(1).Iterator(tarantool.IterEq).Key([]interface{}{key}),
		).Get()
		if err != nil {
			return fmt.Errorf("select failed: %w", err)
		}
		result = resp
		return nil
	})

	if err != nil {
		r.logger.Error("Failed to get KV record", "key", key, "error", err)
		return nil, domain.ErrDatabaseError
	}

	if len(result) == 0 {
		return nil, domain.ErrKeyNotFound
	}

	record, ok := result[0].([]interface{})
	if !ok {
		return nil, domain.ErrDatabaseError
	}
	kv := r.parseRecord(record)

	if kv.IsDeleted {
		return nil, domain.ErrKeyNotFound
	}

	return kv, nil
}

func (r *TarantoolRepository) Update(kv *domain.KV) error {
	now := time.Now().Unix()

	return r.pool.Execute(func(conn *tarantool.Connection) error {
		_, err := conn.Do(
			tarantool.NewUpdateRequest("kv").
				Key([]interface{}{kv.Key}).
				Operations(
					tarantool.NewOperations().
						Assign(1, kv.Value).
						Assign(3, uint32(now)),
				),
		).Get()

		if err != nil {
			r.logger.Error("Failed to update KV record", "key", kv.Key, "error", err)
			return domain.ErrDatabaseError
		}

		kv.UpdatedAt = time.Unix(now, 0)
		r.logger.Info("KV record updated", "key", kv.Key)
		return nil
	})
}

func (r *TarantoolRepository) Delete(key string) (*domain.KV, error) {
	var result []interface{}

	err := r.pool.Execute(func(conn *tarantool.Connection) error {
		resp, err := conn.Do(
			tarantool.NewSelectRequest("kv").Limit(1).Iterator(tarantool.IterEq).Key([]interface{}{key}),
		).Get()
		if err != nil {
			return fmt.Errorf("select before delete failed: %w", err)
		}
		if len(resp) == 0 {
			return domain.ErrKeyNotFound
		}
		record, ok := resp[0].([]interface{})
		if !ok {
			return domain.ErrDatabaseError
		}
		resp, err = conn.Do(
			tarantool.NewReplaceRequest("kv").Tuple([]interface{}{
				key,
				record[1],
				record[2],
				record[3],
				uint32(time.Now().Unix()),
				true,
			}),
		).Get()
		if err != nil {
			return fmt.Errorf("replace failed: %w", err)
		}
		if len(resp) == 0 {
			return nil
		}
		result = resp
		return nil
	})

	if err != nil {
		r.logger.Error("Failed to delete KV record", "key", key, "error", err)
		return nil, domain.ErrDatabaseError
	}

	if len(result) == 0 {
		return nil, domain.ErrKeyNotFound
	}

	record, ok := result[0].([]interface{})
	if !ok {
		return nil, domain.ErrDatabaseError
	}
	kv := r.parseRecord(record)

	r.logger.Info("KV record deleted", "key", key)
	return kv, nil
}

func (r *TarantoolRepository) SoftDelete(key string) error {
	now := time.Now().Unix()

	ops := tarantool.NewOperations().
		Assign(3, uint32(now)).
		Assign(4, uint32(now)).
		Assign(5, true)
	return r.pool.Execute(func(conn *tarantool.Connection) error {
		_, err := conn.Do(
			tarantool.NewUpdateRequest("kv").
				Key([]interface{}{key}).
				Operations(ops),
		).Get()
		if err != nil {
			return fmt.Errorf("soft delete failed: %w", err)
		}
		return nil
	})
}

func (r *TarantoolRepository) Restore(key string) (*domain.KV, error) {
	now := time.Now().Unix()
	var kv *domain.KV

	err := r.pool.Execute(func(conn *tarantool.Connection) error {
		resp, err := conn.Do(
			tarantool.NewSelectRequest("kv").Limit(1).Iterator(tarantool.IterEq).Key([]interface{}{key}),
		).Get()
		if err != nil {
			return fmt.Errorf("select failed: %w", err)
		}
		if len(resp) == 0 {
			return domain.ErrKeyNotFound
		}

		record, ok := resp[0].([]interface{})
		if !ok {
			return domain.ErrDatabaseError
		}
		if len(record) > 5 {
			if isDeleted, ok := record[5].(bool); ok && !isDeleted {
				return domain.ErrNotDeleted
			}
		}

		ops := tarantool.NewOperations().
			Assign(3, uint32(now)).
			Assign(4, uint32(0)).
			Assign(5, false)
		resp, err = conn.Do(
			tarantool.NewUpdateRequest("kv").
				Key([]interface{}{key}).
				Operations(ops),
		).Get()
		if err != nil {
			return fmt.Errorf("update failed: %w", err)
		}
		if len(resp) == 0 {
			return fmt.Errorf("no data returned after update")
		}
		updatedRecord, ok := resp[0].([]interface{})
		if !ok {
			return fmt.Errorf("invalid record format")
		}
		kv = r.parseRecord(updatedRecord)
		return nil
	})

	switch {
	case errors.Is(err, domain.ErrKeyNotFound):
		r.logger.Debug("Key not found for restoration", "key", key)
		return nil, domain.ErrKeyNotFound
	case errors.Is(err, domain.ErrNotDeleted):
		r.logger.Debug("Record was not deleted, cannot restore", "key", key)
		return nil, domain.ErrNotDeleted
	case err != nil:
		r.logger.Error("Failed to restore KV record",
			"key", key,
			"error", err,
			zap.Error(err),
		)
		return nil, domain.ErrDatabaseError
	default:
		r.logger.Info("KV record restored successfully",
			"key", key,
			"restored_at", now,
		)
		return kv, nil
	}
}

func (r *TarantoolRepository) List(limit, offset int) ([]*domain.KV, int, error) {
	var result []interface{}
	var countResult []interface{}

	err := r.pool.Execute(func(conn *tarantool.Connection) error {
		if err := conn.Do(
			tarantool.NewSelectRequest("kv").
				Index("deleted").
				Offset(uint32(offset)).
				Limit(uint32(limit)).
				Iterator(tarantool.IterEq).
				Key([]interface{}{false}),
		).GetTyped(&result); err != nil {
			return err
		}
		return conn.Do(
			tarantool.NewSelectRequest("kv").
				Index("deleted").
				Offset(0).
				Limit(1).
				Iterator(tarantool.IterEq).
				Key([]interface{}{false}),
		).GetTyped(&countResult)
	})

	if err != nil {
		r.logger.Error("Failed to list KV records", "error", err)
		return nil, 0, domain.ErrDatabaseError
	}

	total := len(result)
	items := make([]*domain.KV, 0, len(result))

	for _, record := range result {
		recordData := record.([]interface{})
		kv := r.parseRecord(recordData)
		items = append(items, kv)
	}

	return items, total, nil
}

func (r *TarantoolRepository) ListIncludingDeleted(limit, offset int) ([]*domain.KV, int, error) {
	var result []interface{}
	var countResult []interface{}

	err := r.pool.Execute(func(conn *tarantool.Connection) error {
		if err := conn.Do(
			tarantool.NewSelectRequest("kv").
				Index("primary").
				Offset(uint32(offset)).
				Limit(uint32(limit)).
				Iterator(tarantool.IterAll).
				Key([]interface{}{}),
		).GetTyped(&result); err != nil {
			return err
		}
		return conn.Do(
			tarantool.NewSelectRequest("kv").
				Index("primary").
				Offset(0).
				Limit(1).
				Iterator(tarantool.IterAll).
				Key([]interface{}{}),
		).GetTyped(&countResult)
	})

	if err != nil {
		r.logger.Error("Failed to list KV records including deleted", "error", err)
		return nil, 0, domain.ErrDatabaseError
	}

	total := len(countResult)
	items := make([]*domain.KV, 0, len(result))

	for _, record := range result {
		recordData := record.([]interface{})
		kv := r.parseRecord(recordData)
		items = append(items, kv)
	}

	return items, total, nil
}

func (r *TarantoolRepository) parseRecord(record []interface{}) *domain.KV {
	kv := &domain.KV{
		Key:   record[0].(string),
		Value: record[1].(string),
	}

	switch v := record[2].(type) {
	case int64:
		kv.CreatedAt = time.Unix(v, 0)
	case uint64:
		kv.CreatedAt = time.Unix(int64(v), 0)
	case uint32:
		kv.CreatedAt = time.Unix(int64(v), 0)
	}

	switch v := record[3].(type) {
	case int64:
		kv.UpdatedAt = time.Unix(v, 0)
	case uint64:
		kv.UpdatedAt = time.Unix(int64(v), 0)
	case uint32:
		kv.UpdatedAt = time.Unix(int64(v), 0)
	}

	var deletedAt *time.Time
	switch v := record[4].(type) {
	case int64:
		if v != 0 {
			t := time.Unix(v, 0)
			deletedAt = &t
		}
	case uint64:
		if v != 0 {
			t := time.Unix(int64(v), 0)
			deletedAt = &t
		}
	case uint32:
		if v != 0 {
			t := time.Unix(int64(v), 0)
			deletedAt = &t
		}
	}
	kv.DeletedAt = deletedAt

	if len(record) > 5 && record[5] != nil {
		kv.IsDeleted = record[5].(bool)
	}

	return kv
}

func (r *TarantoolRepository) Close() error {
	r.logger.Info("Closing tarantool connection pool")
	return r.pool.Close()
}
