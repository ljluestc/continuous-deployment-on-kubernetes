package main

import (
	"sync"
	"sync/atomic"
	"time"
)

// CacheConfig holds configuration for all cache layers
type CacheConfig struct {
	// Health cache settings
	HealthCacheTTL     time.Duration
	HealthCacheEnabled bool

	// Stats cache settings
	StatsCacheTTL     time.Duration
	StatsCacheEnabled bool

	// Routing cache settings
	RoutingCacheTTL     time.Duration
	RoutingCacheEnabled bool
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		HealthCacheTTL:      5 * time.Second,
		HealthCacheEnabled:  true,
		StatsCacheTTL:       1 * time.Second,
		StatsCacheEnabled:   true,
		RoutingCacheTTL:     2 * time.Second,
		RoutingCacheEnabled: true,
	}
}

// HealthCacheEntry stores cached health check results
type HealthCacheEntry struct {
	alive      bool
	lastCheck  time.Time
	checkCount int64
	errorCount int64
	avgLatency time.Duration
}

// HealthCache caches backend health status
type HealthCache struct {
	mu      sync.RWMutex
	entries map[string]*HealthCacheEntry // URL -> entry
	ttl     time.Duration
	enabled bool

	// Metrics
	hitCount  int64
	missCount int64
}

// NewHealthCache creates a new health cache
func NewHealthCache(ttl time.Duration, enabled bool) *HealthCache {
	return &HealthCache{
		entries: make(map[string]*HealthCacheEntry),
		ttl:     ttl,
		enabled: enabled,
	}
}

// Get retrieves cached health status
func (hc *HealthCache) Get(url string) (bool, bool) {
	if !hc.enabled {
		return false, false
	}

	hc.mu.RLock()
	defer hc.mu.RUnlock()

	entry, exists := hc.entries[url]
	if !exists {
		atomic.AddInt64(&hc.missCount, 1)
		return false, false
	}

	// Check if entry is expired
	if time.Since(entry.lastCheck) > hc.ttl {
		atomic.AddInt64(&hc.missCount, 1)
		return false, false
	}

	atomic.AddInt64(&hc.hitCount, 1)
	return entry.alive, true
}

// Set stores health status in cache
func (hc *HealthCache) Set(url string, alive bool, latency time.Duration) {
	if !hc.enabled {
		return
	}

	hc.mu.Lock()
	defer hc.mu.Unlock()

	entry, exists := hc.entries[url]
	if !exists {
		entry = &HealthCacheEntry{
			checkCount: 0,
			errorCount: 0,
		}
		hc.entries[url] = entry
	}

	entry.alive = alive
	entry.lastCheck = time.Now()
	entry.checkCount++
	if !alive {
		entry.errorCount++
	}

	// Update average latency (simple moving average)
	if entry.avgLatency == 0 {
		entry.avgLatency = latency
	} else {
		entry.avgLatency = (entry.avgLatency + latency) / 2
	}
}

// Invalidate removes an entry from cache
func (hc *HealthCache) Invalidate(url string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	delete(hc.entries, url)
}

// Clear removes all entries
func (hc *HealthCache) Clear() {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.entries = make(map[string]*HealthCacheEntry)
}

// GetMetrics returns cache metrics
func (hc *HealthCache) GetMetrics() CacheMetrics {
	hc.mu.RLock()
	size := len(hc.entries)
	hc.mu.RUnlock()

	hits := atomic.LoadInt64(&hc.hitCount)
	misses := atomic.LoadInt64(&hc.missCount)
	total := hits + misses

	var hitRate float64
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	return CacheMetrics{
		HitCount:  hits,
		MissCount: misses,
		Size:      int64(size),
		HitRate:   hitRate,
	}
}

// StatsCache caches computed statistics
type StatsCache struct {
	mu         sync.RWMutex
	snapshot   []map[string]interface{}
	lastUpdate time.Time
	ttl        time.Duration
	enabled    bool
	dirty      bool

	// Metrics
	hitCount  int64
	missCount int64
}

// NewStatsCache creates a new stats cache
func NewStatsCache(ttl time.Duration, enabled bool) *StatsCache {
	return &StatsCache{
		ttl:     ttl,
		enabled: enabled,
		dirty:   true,
	}
}

// Get retrieves cached stats
func (sc *StatsCache) Get() ([]map[string]interface{}, bool) {
	if !sc.enabled {
		return nil, false
	}

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	// Check if cache is dirty or expired
	if sc.dirty || time.Since(sc.lastUpdate) > sc.ttl {
		atomic.AddInt64(&sc.missCount, 1)
		return nil, false
	}

	atomic.AddInt64(&sc.hitCount, 1)
	return sc.snapshot, true
}

// Set stores stats in cache
func (sc *StatsCache) Set(stats []map[string]interface{}) {
	if !sc.enabled {
		return
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.snapshot = stats
	sc.lastUpdate = time.Now()
	sc.dirty = false
}

// Invalidate marks cache as dirty
func (sc *StatsCache) Invalidate() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.dirty = true
}

// GetMetrics returns cache metrics
func (sc *StatsCache) GetMetrics() CacheMetrics {
	sc.mu.RLock()
	size := len(sc.snapshot)
	sc.mu.RUnlock()

	hits := atomic.LoadInt64(&sc.hitCount)
	misses := atomic.LoadInt64(&sc.missCount)
	total := hits + misses

	var hitRate float64
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	return CacheMetrics{
		HitCount:  hits,
		MissCount: misses,
		Size:      int64(size),
		HitRate:   hitRate,
	}
}

// RoutingCache caches list of active backends for fast routing
type RoutingCache struct {
	mu             sync.RWMutex
	activeBackends []*Backend
	lastUpdate     time.Time
	ttl            time.Duration
	enabled        bool
	version        uint64

	// Metrics
	hitCount  int64
	missCount int64
}

// NewRoutingCache creates a new routing cache
func NewRoutingCache(ttl time.Duration, enabled bool) *RoutingCache {
	return &RoutingCache{
		ttl:     ttl,
		enabled: enabled,
	}
}

// Get retrieves cached active backends
func (rc *RoutingCache) Get() ([]*Backend, bool) {
	if !rc.enabled {
		return nil, false
	}

	rc.mu.RLock()
	defer rc.mu.RUnlock()

	// Check if cache is expired
	if time.Since(rc.lastUpdate) > rc.ttl {
		atomic.AddInt64(&rc.missCount, 1)
		return nil, false
	}

	if len(rc.activeBackends) == 0 {
		atomic.AddInt64(&rc.missCount, 1)
		return nil, false
	}

	atomic.AddInt64(&rc.hitCount, 1)
	return rc.activeBackends, true
}

// Set stores active backends in cache
func (rc *RoutingCache) Set(backends []*Backend) {
	if !rc.enabled {
		return
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Create a copy to avoid external modifications
	rc.activeBackends = make([]*Backend, len(backends))
	copy(rc.activeBackends, backends)
	rc.lastUpdate = time.Now()
	atomic.AddUint64(&rc.version, 1)
}

// Invalidate clears the routing cache
func (rc *RoutingCache) Invalidate() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.activeBackends = nil
	atomic.AddUint64(&rc.version, 1)
}

// GetVersion returns current cache version
func (rc *RoutingCache) GetVersion() uint64 {
	return atomic.LoadUint64(&rc.version)
}

// GetMetrics returns cache metrics
func (rc *RoutingCache) GetMetrics() CacheMetrics {
	rc.mu.RLock()
	size := len(rc.activeBackends)
	rc.mu.RUnlock()

	hits := atomic.LoadInt64(&rc.hitCount)
	misses := atomic.LoadInt64(&rc.missCount)
	total := hits + misses

	var hitRate float64
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	return CacheMetrics{
		HitCount:  hits,
		MissCount: misses,
		Size:      int64(size),
		HitRate:   hitRate,
	}
}

// CacheMetrics holds metrics for any cache
type CacheMetrics struct {
	HitCount      int64
	MissCount     int64
	Size          int64
	HitRate       float64
	EvictionCount int64
}

// CacheManager manages all caches
type CacheManager struct {
	healthCache  *HealthCache
	statsCache   *StatsCache
	routingCache *RoutingCache
	config       CacheConfig
}

// NewCacheManager creates a new cache manager
func NewCacheManager(config CacheConfig) *CacheManager {
	return &CacheManager{
		healthCache:  NewHealthCache(config.HealthCacheTTL, config.HealthCacheEnabled),
		statsCache:   NewStatsCache(config.StatsCacheTTL, config.StatsCacheEnabled),
		routingCache: NewRoutingCache(config.RoutingCacheTTL, config.RoutingCacheEnabled),
		config:       config,
	}
}

// Health returns the health cache
func (cm *CacheManager) Health() *HealthCache {
	return cm.healthCache
}

// Stats returns the stats cache
func (cm *CacheManager) Stats() *StatsCache {
	return cm.statsCache
}

// Routing returns the routing cache
func (cm *CacheManager) Routing() *RoutingCache {
	return cm.routingCache
}

// InvalidateAll invalidates all caches
func (cm *CacheManager) InvalidateAll() {
	cm.healthCache.Clear()
	cm.statsCache.Invalidate()
	cm.routingCache.Invalidate()
}

// GetAllMetrics returns metrics for all caches
func (cm *CacheManager) GetAllMetrics() map[string]CacheMetrics {
	return map[string]CacheMetrics{
		"health":  cm.healthCache.GetMetrics(),
		"stats":   cm.statsCache.GetMetrics(),
		"routing": cm.routingCache.GetMetrics(),
	}
}
