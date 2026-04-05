package pool

import (
	"errors"
	"sync"
	"time"

	"github.com/unidb/unidb-go/internal"
)

var ErrPoolExhausted = errors.New("connection pool exhausted")

type ConnectionPool struct {
	mu          sync.Mutex
	connections chan internal.Connection
	maxSize     int
	active      int
	createConn  func() (internal.Connection, error)
}

func NewConnectionPool(maxSize int, create func() (internal.Connection, error)) *ConnectionPool {
	return &ConnectionPool{
		connections: make(chan internal.Connection, maxSize),
		maxSize:     maxSize,
		createConn:  create,
	}
}

func (p *ConnectionPool) Acquire() (internal.Connection, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case conn := <-p.connections:
		return conn, nil
	default:
		if p.active < p.maxSize {
			conn, err := p.createConn()
			if err != nil {
				return nil, err
			}
			p.active++
			return conn, nil
		}
		return nil, ErrPoolExhausted
	}
}

func (p *ConnectionPool) Release(conn internal.Connection) {
	p.connections <- conn
}

// Adaptive tuning placeholder
func (p *ConnectionPool) AutoTune() {
	for {
		time.Sleep(5 * time.Second)
		// gather metrics and adjust p.maxSize
	}
}

type PoolManager struct {
	pools map[string]*ConnectionPool
	mu    sync.RWMutex
}

func NewPoolManager() *PoolManager {
	return &PoolManager{
		pools: make(map[string]*ConnectionPool),
	}
}

func (m *PoolManager) GetPool(dbName string) *ConnectionPool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pools[dbName]
}

func (m *PoolManager) AddPool(dbName string, pool *ConnectionPool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pools[dbName] = pool
}
