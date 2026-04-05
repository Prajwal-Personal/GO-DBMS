package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/unidb/unidb-go/federation"
)

type CacheEntry struct {
	Result    []federation.Row
	ExpiresAt time.Time
}

type Engine struct {
	cache map[string]CacheEntry
	mu    sync.RWMutex
}

func NewEngine() *Engine {
	return &Engine{
		cache: make(map[string]CacheEntry),
	}
}

func (e *Engine) hash(query string, args []any) string {
	str := query + fmt.Sprint(args)
	hash := sha256.Sum256([]byte(str))
	return hex.EncodeToString(hash[:])
}

func (e *Engine) Get(query string, args []any) ([]federation.Row, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	key := e.hash(query, args)
	val, ok := e.cache[key]
	if ok && time.Now().Before(val.ExpiresAt) {
		return val.Result, true
	}
	return nil, false
}

func (e *Engine) Set(query string, args []any, result []federation.Row, ttl time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := e.hash(query, args)
	e.cache[key] = CacheEntry{
		Result:    result,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func (e *Engine) Invalidate() {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	// simple implementation: flush everything on modification
	e.cache = make(map[string]CacheEntry)
}
