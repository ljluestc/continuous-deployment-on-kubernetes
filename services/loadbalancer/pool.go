package main

import (
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// PooledConnection represents a pooled HTTP connection
type PooledConnection struct {
	client   *http.Client
	lastUsed time.Time
	useCount int64
	created  time.Time
	mu       sync.RWMutex
}

// ConnectionPool manages HTTP client connections for health checks and backend communication
type ConnectionPool struct {
	mu          sync.RWMutex
	connections map[string]*PooledConnection // URL -> connection
	maxIdle     int                           // Max idle connections per backend
	maxLifetime time.Duration                 // Max connection lifetime
	idleTimeout time.Duration                 // Idle timeout before cleanup

	// Metrics
	hitCount      int64
	missCount     int64
	evictionCount int64
	createCount   int64
}

// PoolConfig holds connection pool configuration
type PoolConfig struct {
	MaxIdleConns    int           // Default: 10 per backend
	MaxLifetime     time.Duration // Default: 60s
	IdleTimeout     time.Duration // Default: 30s
	CleanupInterval time.Duration // Default: 10s
	RequestTimeout  time.Duration // Default: 2s for health checks
}

// NewConnectionPool creates a new connection pool with the given configuration
func NewConnectionPool(config PoolConfig) *ConnectionPool {
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 10
	}
	if config.MaxLifetime == 0 {
		config.MaxLifetime = 60 * time.Second
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = 30 * time.Second
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 10 * time.Second
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 2 * time.Second
	}

	pool := &ConnectionPool{
		connections: make(map[string]*PooledConnection),
		maxIdle:     config.MaxIdleConns,
		maxLifetime: config.MaxLifetime,
		idleTimeout: config.IdleTimeout,
	}

	// Start cleanup goroutine
	go pool.cleanupLoop(config.CleanupInterval)

	return pool
}

// Get retrieves or creates a pooled connection for the given URL
func (p *ConnectionPool) Get(u *url.URL, timeout time.Duration) *http.Client {
	key := u.String()

	// Try to get existing connection (fast path with read lock)
	p.mu.RLock()
	conn, exists := p.connections[key]
	p.mu.RUnlock()

	if exists && !p.isExpired(conn) {
		conn.mu.Lock()
		conn.lastUsed = time.Now()
		atomic.AddInt64(&conn.useCount, 1)
		client := conn.client
		conn.mu.Unlock()

		atomic.AddInt64(&p.hitCount, 1)
		return client
	}

	// Create new connection (slow path with write lock)
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	conn, exists = p.connections[key]
	if exists && !p.isExpired(conn) {
		conn.mu.Lock()
		conn.lastUsed = time.Now()
		atomic.AddInt64(&conn.useCount, 1)
		client := conn.client
		conn.mu.Unlock()

		atomic.AddInt64(&p.hitCount, 1)
		return client
	}

	// Create new pooled connection
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        p.maxIdle,
			MaxIdleConnsPerHost: p.maxIdle,
			IdleConnTimeout:     p.idleTimeout,
			DisableKeepAlives:   false,
		},
	}

	conn = &PooledConnection{
		client:   client,
		lastUsed: time.Now(),
		created:  time.Now(),
		useCount: 1,
	}

	p.connections[key] = conn
	atomic.AddInt64(&p.createCount, 1)
	atomic.AddInt64(&p.missCount, 1)

	return client
}

// isExpired checks if a connection has expired based on lifetime or idle time
func (p *ConnectionPool) isExpired(conn *PooledConnection) bool {
	conn.mu.RLock()
	defer conn.mu.RUnlock()

	now := time.Now()

	// Check max lifetime
	if now.Sub(conn.created) > p.maxLifetime {
		return true
	}

	// Check idle timeout
	if now.Sub(conn.lastUsed) > p.idleTimeout {
		return true
	}

	return false
}

// cleanupLoop periodically removes expired connections
func (p *ConnectionPool) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		p.cleanup()
	}
}

// cleanup removes expired connections from the pool
func (p *ConnectionPool) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for key, conn := range p.connections {
		conn.mu.RLock()
		expired := now.Sub(conn.created) > p.maxLifetime ||
			now.Sub(conn.lastUsed) > p.idleTimeout
		conn.mu.RUnlock()

		if expired {
			delete(p.connections, key)
			atomic.AddInt64(&p.evictionCount, 1)
		}
	}
}

// Close closes all connections in the pool
func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Clear all connections
	p.connections = make(map[string]*PooledConnection)
}

// GetMetrics returns connection pool metrics
func (p *ConnectionPool) GetMetrics() PoolMetrics {
	p.mu.RLock()
	size := len(p.connections)
	p.mu.RUnlock()

	hits := atomic.LoadInt64(&p.hitCount)
	misses := atomic.LoadInt64(&p.missCount)
	total := hits + misses

	var hitRate float64
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	return PoolMetrics{
		Size:          size,
		HitCount:      hits,
		MissCount:     misses,
		HitRate:       hitRate,
		EvictionCount: atomic.LoadInt64(&p.evictionCount),
		CreateCount:   atomic.LoadInt64(&p.createCount),
	}
}

// PoolMetrics holds connection pool metrics
type PoolMetrics struct {
	Size          int
	HitCount      int64
	MissCount     int64
	HitRate       float64
	EvictionCount int64
	CreateCount   int64
}

// Reset resets the pool metrics
func (p *ConnectionPool) Reset() {
	atomic.StoreInt64(&p.hitCount, 0)
	atomic.StoreInt64(&p.missCount, 0)
	atomic.StoreInt64(&p.evictionCount, 0)
	atomic.StoreInt64(&p.createCount, 0)
}
