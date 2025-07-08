package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"kv-storage/internal/config"
	"kv-storage/internal/interfaces"

	"github.com/tarantool/go-tarantool/v2"
)

type ConnectionPool struct {
	connections chan *tarantool.Connection
	config      *config.Config
	logger      interfaces.Logger
	mu          sync.RWMutex
	closed      bool
}

func NewConnectionPool(cfg *config.Config, logger interfaces.Logger, poolSize int) (*ConnectionPool, error) {
	pool := &ConnectionPool{
		connections: make(chan *tarantool.Connection, poolSize),
		config:      cfg,
		logger:      logger,
	}

	for i := 0; i < poolSize; i++ {
		conn, err := pool.createConnection()
		if err != nil {
			pool.Close()
			return nil, fmt.Errorf("failed to create connection %d: %w", i, err)
		}
		pool.connections <- conn
	}

	logger.Info("Connection pool initialized", "size", poolSize)
	return pool, nil
}

func (p *ConnectionPool) createConnection() (*tarantool.Connection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dialer := tarantool.NetDialer{
		Address:  fmt.Sprintf("%s:%d", p.config.Tarantool.Host, p.config.Tarantool.Port),
		User:     p.config.Tarantool.Username,
		Password: p.config.Tarantool.Password,
	}

	opts := tarantool.Opts{
		Timeout:     p.config.Tarantool.Timeout,
		Concurrency: 32,
	}

	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tarantool: %w", err)
	}

	return conn, nil
}

func (p *ConnectionPool) Get() (*tarantool.Connection, error) {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return nil, fmt.Errorf("connection pool is closed")
	}
	p.mu.RUnlock()

	select {
	case conn := <-p.connections:
		if conn == nil {
			return nil, fmt.Errorf("connection pool is closed")
		}
		return conn, nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for connection")
	}
}

func (p *ConnectionPool) Put(conn *tarantool.Connection) {
	if conn == nil {
		return
	}

	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		conn.Close()
		return
	}
	p.mu.RUnlock()

	select {
	case p.connections <- conn:
	default:
		conn.Close()
	}
}

func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	close(p.connections)

	for conn := range p.connections {
		if conn != nil {
			conn.Close()
		}
	}

	p.logger.Info("Connection pool closed")
	return nil
}

func (p *ConnectionPool) Execute(fn func(*tarantool.Connection) error) error {
	conn, err := p.Get()
	if err != nil {
		return err
	}
	defer p.Put(conn)

	return fn(conn)
}
