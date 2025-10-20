package main

import (
	"net/url"
	"testing"
	"time"
)

// TestConnectionPoolGet tests getting connections from the pool
func TestConnectionPoolGet(t *testing.T) {
	config := PoolConfig{
		MaxIdleConns:    10,
		MaxLifetime:     60 * time.Second,
		IdleTimeout:     30 * time.Second,
		CleanupInterval: 100 * time.Millisecond,
		RequestTimeout:  2 * time.Second,
	}
	pool := NewConnectionPool(config)
	defer pool.Close()

	u, _ := url.Parse("http://backend1:8080")

	// First get should be a cache miss
	client1 := pool.Get(u, 2*time.Second)
	if client1 == nil {
		t.Fatal("Expected non-nil client")
	}

	metrics := pool.GetMetrics()
	if metrics.MissCount != 1 {
		t.Errorf("Expected 1 miss, got %d", metrics.MissCount)
	}
	if metrics.CreateCount != 1 {
		t.Errorf("Expected 1 create, got %d", metrics.CreateCount)
	}

	// Second get should be a cache hit
	client2 := pool.Get(u, 2*time.Second)
	if client2 == nil {
		t.Fatal("Expected non-nil client")
	}

	metrics = pool.GetMetrics()
	if metrics.HitCount != 1 {
		t.Errorf("Expected 1 hit, got %d", metrics.HitCount)
	}
}

// TestConnectionPoolMultipleURLs tests pool with multiple URLs
func TestConnectionPoolMultipleURLs(t *testing.T) {
	config := PoolConfig{
		MaxIdleConns:    10,
		MaxLifetime:     60 * time.Second,
		IdleTimeout:     30 * time.Second,
		CleanupInterval: 100 * time.Millisecond,
		RequestTimeout:  2 * time.Second,
	}
	pool := NewConnectionPool(config)
	defer pool.Close()

	urls := []string{
		"http://backend1:8080",
		"http://backend2:8080",
		"http://backend3:8080",
	}

	// Get clients for each URL
	for _, urlStr := range urls {
		u, _ := url.Parse(urlStr)
		client := pool.Get(u, 2*time.Second)
		if client == nil {
			t.Fatalf("Expected non-nil client for %s", urlStr)
		}
	}

	metrics := pool.GetMetrics()
	if metrics.Size != 3 {
		t.Errorf("Expected pool size 3, got %d", metrics.Size)
	}
	if metrics.CreateCount != 3 {
		t.Errorf("Expected 3 creates, got %d", metrics.CreateCount)
	}

	// Get same URLs again - should be cache hits
	for _, urlStr := range urls {
		u, _ := url.Parse(urlStr)
		pool.Get(u, 2*time.Second)
	}

	metrics = pool.GetMetrics()
	if metrics.HitCount != 3 {
		t.Errorf("Expected 3 hits, got %d", metrics.HitCount)
	}
}

// TestConnectionPoolExpiration tests connection expiration
func TestConnectionPoolExpiration(t *testing.T) {
	config := PoolConfig{
		MaxIdleConns:    10,
		MaxLifetime:     100 * time.Millisecond,
		IdleTimeout:     50 * time.Millisecond,
		CleanupInterval: 30 * time.Millisecond,
		RequestTimeout:  2 * time.Second,
	}
	pool := NewConnectionPool(config)
	defer pool.Close()

	u, _ := url.Parse("http://backend1:8080")

	// Get initial connection
	pool.Get(u, 2*time.Second)

	metrics := pool.GetMetrics()
	if metrics.Size != 1 {
		t.Errorf("Expected pool size 1, got %d", metrics.Size)
	}

	// Wait for expiration and cleanup
	time.Sleep(200 * time.Millisecond)

	metrics = pool.GetMetrics()
	if metrics.Size != 0 {
		t.Errorf("Expected pool size 0 after expiration, got %d", metrics.Size)
	}
	if metrics.EvictionCount != 1 {
		t.Errorf("Expected 1 eviction, got %d", metrics.EvictionCount)
	}
}

// TestConnectionPoolIdleTimeout tests idle timeout
func TestConnectionPoolIdleTimeout(t *testing.T) {
	config := PoolConfig{
		MaxIdleConns:    10,
		MaxLifetime:     10 * time.Second,
		IdleTimeout:     100 * time.Millisecond,
		CleanupInterval: 50 * time.Millisecond,
		RequestTimeout:  2 * time.Second,
	}
	pool := NewConnectionPool(config)
	defer pool.Close()

	u, _ := url.Parse("http://backend1:8080")

	// Get and use connection
	pool.Get(u, 2*time.Second)

	// Wait for idle timeout
	time.Sleep(200 * time.Millisecond)

	metrics := pool.GetMetrics()
	if metrics.EvictionCount != 1 {
		t.Errorf("Expected 1 eviction, got %d", metrics.EvictionCount)
	}
}

// TestConnectionPoolConcurrent tests concurrent access to pool
func TestConnectionPoolConcurrent(t *testing.T) {
	config := PoolConfig{
		MaxIdleConns:    10,
		MaxLifetime:     60 * time.Second,
		IdleTimeout:     30 * time.Second,
		CleanupInterval: 100 * time.Millisecond,
		RequestTimeout:  2 * time.Second,
	}
	pool := NewConnectionPool(config)
	defer pool.Close()

	done := make(chan bool)
	numGoroutines := 10
	requestsPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			u, _ := url.Parse("http://backend1:8080")
			for j := 0; j < requestsPerGoroutine; j++ {
				client := pool.Get(u, 2*time.Second)
				if client == nil {
					t.Error("Expected non-nil client")
				}
			}
			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	metrics := pool.GetMetrics()
	totalRequests := int64(numGoroutines * requestsPerGoroutine)
	if metrics.HitCount+metrics.MissCount != totalRequests {
		t.Errorf("Expected %d total requests, got %d", totalRequests, metrics.HitCount+metrics.MissCount)
	}

	// Should have high hit rate
	hitRate := metrics.HitRate
	if hitRate < 90.0 {
		t.Errorf("Expected hit rate >90%%, got %.2f%%", hitRate)
	}
}

// TestConnectionPoolClose tests pool cleanup on close
func TestConnectionPoolClose(t *testing.T) {
	config := PoolConfig{
		MaxIdleConns:    10,
		MaxLifetime:     60 * time.Second,
		IdleTimeout:     30 * time.Second,
		CleanupInterval: 100 * time.Millisecond,
		RequestTimeout:  2 * time.Second,
	}
	pool := NewConnectionPool(config)

	u, _ := url.Parse("http://backend1:8080")
	pool.Get(u, 2*time.Second)

	metrics := pool.GetMetrics()
	if metrics.Size != 1 {
		t.Errorf("Expected pool size 1, got %d", metrics.Size)
	}

	pool.Close()

	metrics = pool.GetMetrics()
	if metrics.Size != 0 {
		t.Errorf("Expected pool size 0 after close, got %d", metrics.Size)
	}
}

// TestConnectionPoolReset tests metrics reset
func TestConnectionPoolReset(t *testing.T) {
	config := PoolConfig{
		MaxIdleConns:    10,
		MaxLifetime:     60 * time.Second,
		IdleTimeout:     30 * time.Second,
		CleanupInterval: 100 * time.Millisecond,
		RequestTimeout:  2 * time.Second,
	}
	pool := NewConnectionPool(config)
	defer pool.Close()

	u, _ := url.Parse("http://backend1:8080")
	pool.Get(u, 2*time.Second)
	pool.Get(u, 2*time.Second)

	metrics := pool.GetMetrics()
	if metrics.HitCount == 0 && metrics.MissCount == 0 {
		t.Error("Expected non-zero metrics before reset")
	}

	pool.Reset()

	metrics = pool.GetMetrics()
	if metrics.HitCount != 0 || metrics.MissCount != 0 {
		t.Error("Expected zero metrics after reset")
	}
}

// TestConnectionPoolDefaultConfig tests default configuration
func TestConnectionPoolDefaultConfig(t *testing.T) {
	config := PoolConfig{} // Empty config - should use defaults
	pool := NewConnectionPool(config)
	defer pool.Close()

	if pool.maxIdle != 10 {
		t.Errorf("Expected default maxIdle 10, got %d", pool.maxIdle)
	}
	if pool.maxLifetime != 60*time.Second {
		t.Errorf("Expected default maxLifetime 60s, got %v", pool.maxLifetime)
	}
	if pool.idleTimeout != 30*time.Second {
		t.Errorf("Expected default idleTimeout 30s, got %v", pool.idleTimeout)
	}
}

// TestConnectionPoolMetricsHitRate tests hit rate calculation
func TestConnectionPoolMetricsHitRate(t *testing.T) {
	config := PoolConfig{
		MaxIdleConns:    10,
		MaxLifetime:     60 * time.Second,
		IdleTimeout:     30 * time.Second,
		CleanupInterval: 100 * time.Millisecond,
		RequestTimeout:  2 * time.Second,
	}
	pool := NewConnectionPool(config)
	defer pool.Close()

	u, _ := url.Parse("http://backend1:8080")

	// 1 miss
	pool.Get(u, 2*time.Second)

	// 4 hits
	for i := 0; i < 4; i++ {
		pool.Get(u, 2*time.Second)
	}

	metrics := pool.GetMetrics()
	expectedHitRate := 80.0 // 4 hits / 5 total = 80%

	if metrics.HitRate < expectedHitRate-0.01 || metrics.HitRate > expectedHitRate+0.01 {
		t.Errorf("Expected hit rate %.2f%%, got %.2f%%", expectedHitRate, metrics.HitRate)
	}
}

// BenchmarkConnectionPoolGet benchmarks pool get operation
func BenchmarkConnectionPoolGet(b *testing.B) {
	config := PoolConfig{
		MaxIdleConns:    10,
		MaxLifetime:     60 * time.Second,
		IdleTimeout:     30 * time.Second,
		CleanupInterval: 1 * time.Second,
		RequestTimeout:  2 * time.Second,
	}
	pool := NewConnectionPool(config)
	defer pool.Close()

	u, _ := url.Parse("http://backend1:8080")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Get(u, 2*time.Second)
	}
}

// BenchmarkConnectionPoolConcurrent benchmarks concurrent pool access
func BenchmarkConnectionPoolConcurrent(b *testing.B) {
	config := PoolConfig{
		MaxIdleConns:    10,
		MaxLifetime:     60 * time.Second,
		IdleTimeout:     30 * time.Second,
		CleanupInterval: 1 * time.Second,
		RequestTimeout:  2 * time.Second,
	}
	pool := NewConnectionPool(config)
	defer pool.Close()

	u, _ := url.Parse("http://backend1:8080")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pool.Get(u, 2*time.Second)
		}
	})
}
