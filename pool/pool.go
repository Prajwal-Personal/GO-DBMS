package pool

import (
	"errors"
	"sync"
	"time"

	"github.com/unidb/unidb-go/circuit"
	"github.com/unidb/unidb-go/internal"
)

var ErrPoolExhausted = errors.New("connection pool exhausted")

type ConnectionPool struct {
	mu          sync.Mutex
	connections chan internal.Connection
	maxSize     int
	active      int
	exhausted   int
	createConn  func() (internal.Connection, error)
}

func NewConnectionPool(maxSize int, create func() (internal.Connection, error)) *ConnectionPool {
	p := &ConnectionPool{
		connections: make(chan internal.Connection, maxSize),
		maxSize:     maxSize,
		createConn:  create,
	}
	go p.AutoTune()
	return p
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
		p.exhausted++
		return nil, ErrPoolExhausted
	}
}

func (p *ConnectionPool) Release(conn internal.Connection) {
	p.connections <- conn
}

// AutoTune continuously monitors and resizes the connection pool based on workload
func (p *ConnectionPool) AutoTune() {
	for {
		time.Sleep(5 * time.Second)
		p.mu.Lock()

		exhaustedRate := p.exhausted
		p.exhausted = 0 // reset window
		idleCount := len(p.connections)

		// Scale Up if we hit connection pool limits recently
		if exhaustedRate > 5 {
			p.maxSize += 5
			newChan := make(chan internal.Connection, p.maxSize)
			for i := 0; i < idleCount; i++ {
				newChan <- <-p.connections
			}
			p.connections = newChan
			
		// Scale Down if we have too many idle connections sitting around
		} else if idleCount > 10 && p.maxSize > 10 {
			p.maxSize -= 2
			newChan := make(chan internal.Connection, p.maxSize)

			transferCount := idleCount
			if transferCount > p.maxSize {
				transferCount = p.maxSize
			}

			// Drop excess idle connections beyond new capacity
			for i := 0; i < transferCount; i++ {
				newChan <- <-p.connections
			}
			p.connections = newChan
		}

		p.mu.Unlock()
	}
}

type PoolManager struct {
	pools    map[string]*ConnectionPool
	breakers map[string]*circuit.CircuitBreaker
	mu       sync.RWMutex
}

func NewPoolManager() *PoolManager {
	return &PoolManager{
		pools:    make(map[string]*ConnectionPool),
		breakers: make(map[string]*circuit.CircuitBreaker),
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
	// Allocate a circuit breaker per connection pool: 5 failures threshold, 30s open timeout
	m.breakers[dbName] = circuit.NewCircuitBreaker(5, 30*time.Second)
}

// ExecuteWithBreaker executes a database operation guarded by the circuit breaker for that DB.
func (m *PoolManager) ExecuteWithBreaker(dbName string, operation func() error) error {
	m.mu.RLock()
	cb, ok := m.breakers[dbName]
	m.mu.RUnlock()

	if !ok {
		// no breaker defined, fallback to unprotected operation
		return operation()
	}

	return cb.Execute(operation)
}
