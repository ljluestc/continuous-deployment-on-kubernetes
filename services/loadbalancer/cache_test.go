package main

import (
	"testing"
	"time"
)

// TestHealthCache tests the health cache functionality
func TestHealthCache(t *testing.T) {
	cache := NewHealthCache(100*time.Millisecond, true)

	// Test cache miss
	_, found := cache.Get("http://backend1")
	if found {
		t.Error("Expected cache miss, got hit")
	}

	// Test cache set and hit
	cache.Set("http://backend1", true, 50*time.Millisecond)
	alive, found := cache.Get("http://backend1")
	if !found {
		t.Error("Expected cache hit, got miss")
	}
	if !alive {
		t.Error("Expected alive=true")
	}

	// Test TTL expiration
	time.Sleep(150 * time.Millisecond)
	_, found = cache.Get("http://backend1")
	if found {
		t.Error("Expected cache miss after TTL expiration")
	}

	// Test cache invalidation
	cache.Set("http://backend2", true, 10*time.Millisecond)
	cache.Invalidate("http://backend2")
	_, found = cache.Get("http://backend2")
	if found {
		t.Error("Expected cache miss after invalidation")
	}

	// Test metrics
	metrics := cache.GetMetrics()
	if metrics.HitCount != 1 {
		t.Errorf("Expected 1 hit, got %d", metrics.HitCount)
	}
	if metrics.MissCount != 3 {
		t.Errorf("Expected 3 misses, got %d", metrics.MissCount)
	}
}

// TestHealthCacheDisabled tests that cache doesn't function when disabled
func TestHealthCacheDisabled(t *testing.T) {
	cache := NewHealthCache(1*time.Second, false)

	cache.Set("http://backend1", true, 10*time.Millisecond)
	_, found := cache.Get("http://backend1")
	if found {
		t.Error("Expected cache miss when disabled")
	}
}

// TestHealthCacheConcurrent tests concurrent access to health cache
func TestHealthCacheConcurrent(t *testing.T) {
	cache := NewHealthCache(1*time.Second, true)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				url := "http://backend1"
				cache.Set(url, true, time.Millisecond)
				cache.Get(url)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	metrics := cache.GetMetrics()
	if metrics.Size <= 0 {
		t.Error("Expected non-zero cache size")
	}
}

// TestStatsCache tests the stats cache functionality
func TestStatsCache(t *testing.T) {
	cache := NewStatsCache(100*time.Millisecond, true)

	// Test cache miss
	_, found := cache.Get()
	if found {
		t.Error("Expected cache miss, got hit")
	}

	// Test cache set and hit
	stats := []map[string]interface{}{
		{"url": "http://backend1", "alive": true},
	}
	cache.Set(stats)
	cachedStats, found := cache.Get()
	if !found {
		t.Error("Expected cache hit, got miss")
	}
	if len(cachedStats) != 1 {
		t.Errorf("Expected 1 stat, got %d", len(cachedStats))
	}

	// Test TTL expiration
	time.Sleep(150 * time.Millisecond)
	_, found = cache.Get()
	if found {
		t.Error("Expected cache miss after TTL expiration")
	}

	// Test cache invalidation
	cache.Set(stats)
	cache.Invalidate()
	_, found = cache.Get()
	if found {
		t.Error("Expected cache miss after invalidation")
	}

	// Test metrics
	metrics := cache.GetMetrics()
	if metrics.HitCount != 1 {
		t.Errorf("Expected 1 hit, got %d", metrics.HitCount)
	}
}

// TestStatsCacheDisabled tests that cache doesn't function when disabled
func TestStatsCacheDisabled(t *testing.T) {
	cache := NewStatsCache(1*time.Second, false)

	stats := []map[string]interface{}{{"url": "http://backend1"}}
	cache.Set(stats)
	_, found := cache.Get()
	if found {
		t.Error("Expected cache miss when disabled")
	}
}

// TestRoutingCache tests the routing cache functionality
func TestRoutingCache(t *testing.T) {
	cache := NewRoutingCache(100*time.Millisecond, true)

	// Create test backends
	backend1 := &Backend{Alive: true}
	backend2 := &Backend{Alive: true}
	backends := []*Backend{backend1, backend2}

	// Test cache miss
	_, found := cache.Get()
	if found {
		t.Error("Expected cache miss, got hit")
	}

	// Test cache set and hit
	cache.Set(backends)
	cachedBackends, found := cache.Get()
	if !found {
		t.Error("Expected cache hit, got miss")
	}
	if len(cachedBackends) != 2 {
		t.Errorf("Expected 2 backends, got %d", len(cachedBackends))
	}

	// Test TTL expiration
	time.Sleep(150 * time.Millisecond)
	_, found = cache.Get()
	if found {
		t.Error("Expected cache miss after TTL expiration")
	}

	// Test cache invalidation
	cache.Set(backends)
	version1 := cache.GetVersion()
	cache.Invalidate()
	version2 := cache.GetVersion()
	if version2 <= version1 {
		t.Error("Expected version to increment after invalidation")
	}
	_, found = cache.Get()
	if found {
		t.Error("Expected cache miss after invalidation")
	}

	// Test metrics
	metrics := cache.GetMetrics()
	if metrics.HitCount != 1 {
		t.Errorf("Expected 1 hit, got %d", metrics.HitCount)
	}
}

// TestRoutingCacheEmptyBackends tests routing cache with empty backends
func TestRoutingCacheEmptyBackends(t *testing.T) {
	cache := NewRoutingCache(1*time.Second, true)

	// Set empty backends
	cache.Set([]*Backend{})

	// Should return cache miss for empty backends
	_, found := cache.Get()
	if found {
		t.Error("Expected cache miss for empty backends")
	}
}

// TestRoutingCacheCopyIsolation tests that cached backends are isolated
func TestRoutingCacheCopyIsolation(t *testing.T) {
	cache := NewRoutingCache(1*time.Second, true)

	backend := &Backend{Alive: true}
	backends := []*Backend{backend}

	cache.Set(backends)

	// Modify original slice
	backends[0] = nil

	// Cached version should be unaffected
	cachedBackends, found := cache.Get()
	if !found {
		t.Fatal("Expected cache hit")
	}
	if cachedBackends[0] == nil {
		t.Error("Cache should have isolated copy of backends")
	}
}

// TestCacheManager tests the cache manager
func TestCacheManager(t *testing.T) {
	config := CacheConfig{
		HealthCacheTTL:      1 * time.Second,
		HealthCacheEnabled:  true,
		StatsCacheTTL:       1 * time.Second,
		StatsCacheEnabled:   true,
		RoutingCacheTTL:     1 * time.Second,
		RoutingCacheEnabled: true,
	}

	manager := NewCacheManager(config)

	// Test health cache access
	manager.Health().Set("http://backend1", true, time.Millisecond)
	_, found := manager.Health().Get("http://backend1")
	if !found {
		t.Error("Expected health cache to work")
	}

	// Test stats cache access
	stats := []map[string]interface{}{{"test": "data"}}
	manager.Stats().Set(stats)
	_, found = manager.Stats().Get()
	if !found {
		t.Error("Expected stats cache to work")
	}

	// Test routing cache access
	backends := []*Backend{{Alive: true}}
	manager.Routing().Set(backends)
	_, found = manager.Routing().Get()
	if !found {
		t.Error("Expected routing cache to work")
	}

	// Test invalidate all
	manager.InvalidateAll()
	_, found = manager.Health().Get("http://backend1")
	if found {
		t.Error("Expected health cache to be cleared")
	}
	_, found = manager.Stats().Get()
	if found {
		t.Error("Expected stats cache to be cleared")
	}
	_, found = manager.Routing().Get()
	if found {
		t.Error("Expected routing cache to be cleared")
	}

	// Test metrics
	metrics := manager.GetAllMetrics()
	if len(metrics) != 3 {
		t.Errorf("Expected 3 cache metrics, got %d", len(metrics))
	}
	if _, ok := metrics["health"]; !ok {
		t.Error("Expected health cache metrics")
	}
	if _, ok := metrics["stats"]; !ok {
		t.Error("Expected stats cache metrics")
	}
	if _, ok := metrics["routing"]; !ok {
		t.Error("Expected routing cache metrics")
	}
}

// TestDefaultCacheConfig tests default configuration
func TestDefaultCacheConfig(t *testing.T) {
	config := DefaultCacheConfig()

	if config.HealthCacheTTL != 5*time.Second {
		t.Errorf("Expected health TTL 5s, got %v", config.HealthCacheTTL)
	}
	if config.StatsCacheTTL != 1*time.Second {
		t.Errorf("Expected stats TTL 1s, got %v", config.StatsCacheTTL)
	}
	if config.RoutingCacheTTL != 2*time.Second {
		t.Errorf("Expected routing TTL 2s, got %v", config.RoutingCacheTTL)
	}
	if !config.HealthCacheEnabled || !config.StatsCacheEnabled || !config.RoutingCacheEnabled {
		t.Error("Expected all caches to be enabled by default")
	}
}

// BenchmarkHealthCacheSet benchmarks cache set operations
func BenchmarkHealthCacheSet(b *testing.B) {
	cache := NewHealthCache(1*time.Second, true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("http://backend1", true, time.Millisecond)
	}
}

// BenchmarkHealthCacheGet benchmarks cache get operations
func BenchmarkHealthCacheGet(b *testing.B) {
	cache := NewHealthCache(1*time.Second, true)
	cache.Set("http://backend1", true, time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("http://backend1")
	}
}

// BenchmarkStatsCacheSet benchmarks stats cache set
func BenchmarkStatsCacheSet(b *testing.B) {
	cache := NewStatsCache(1*time.Second, true)
	stats := []map[string]interface{}{{"url": "http://backend1"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(stats)
	}
}

// BenchmarkStatsCacheGet benchmarks stats cache get
func BenchmarkStatsCacheGet(b *testing.B) {
	cache := NewStatsCache(1*time.Second, true)
	stats := []map[string]interface{}{{"url": "http://backend1"}}
	cache.Set(stats)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get()
	}
}

// BenchmarkRoutingCacheSet benchmarks routing cache set
func BenchmarkRoutingCacheSet(b *testing.B) {
	cache := NewRoutingCache(1*time.Second, true)
	backends := []*Backend{{Alive: true}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(backends)
	}
}

// BenchmarkRoutingCacheGet benchmarks routing cache get
func BenchmarkRoutingCacheGet(b *testing.B) {
	cache := NewRoutingCache(1*time.Second, true)
	backends := []*Backend{{Alive: true}}
	cache.Set(backends)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get()
	}
}
