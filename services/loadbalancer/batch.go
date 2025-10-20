package main

import (
	"context"
	"net/url"
	"sync"
	"time"
)

// BatchRequest represents a batched request
type BatchRequest struct {
	key        string
	responseCh chan interface{}
	errorCh    chan error
	waiters    []*waiter
}

// waiter represents a goroutine waiting for a result
type waiter struct {
	responseCh chan interface{}
	errorCh    chan error
}

// Batcher handles batching of requests to reduce overhead
type Batcher struct {
	mu            sync.RWMutex
	pending       map[string]*BatchRequest
	batchSize     int
	batchTimeout  time.Duration
	processFn     func(keys []string) (map[string]interface{}, error)
	flushInterval time.Duration

	// Metrics
	batchCount     int64
	requestCount   int64
	coalescedCount int64
}

// BatcherConfig holds batcher configuration
type BatcherConfig struct {
	BatchSize     int           // Max requests per batch (default: 10)
	BatchTimeout  time.Duration // Max wait time (default: 100ms)
	FlushInterval time.Duration // Periodic flush (default: 50ms)
}

// NewBatcher creates a new request batcher
func NewBatcher(config BatcherConfig, processFn func([]string) (map[string]interface{}, error)) *Batcher {
	if config.BatchSize == 0 {
		config.BatchSize = 10
	}
	if config.BatchTimeout == 0 {
		config.BatchTimeout = 100 * time.Millisecond
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 50 * time.Millisecond
	}

	b := &Batcher{
		pending:       make(map[string]*BatchRequest),
		batchSize:     config.BatchSize,
		batchTimeout:  config.BatchTimeout,
		flushInterval: config.FlushInterval,
		processFn:     processFn,
	}

	// Start background flush goroutine
	go b.flushLoop()

	return b
}

// Submit submits a request for batching
func (b *Batcher) Submit(ctx context.Context, key string) (interface{}, error) {
	// Create waiter channels for this specific goroutine
	w := &waiter{
		responseCh: make(chan interface{}, 1),
		errorCh:    make(chan error, 1),
	}

	b.mu.Lock()

	// Check if request already exists (coalesce)
	if req, exists := b.pending[key]; exists {
		// Add this waiter to the existing request
		req.waiters = append(req.waiters, w)
		b.mu.Unlock()
		b.coalescedCount++
	} else {
		// Create new request
		req := &BatchRequest{
			key:        key,
			responseCh: make(chan interface{}, 1),
			errorCh:    make(chan error, 1),
			waiters:    []*waiter{w},
		}
		b.pending[key] = req
		b.requestCount++

		// Check if batch is full
		shouldFlush := len(b.pending) >= b.batchSize
		b.mu.Unlock()

		if shouldFlush {
			go b.flush()
		} else {
			// Start timeout timer
			time.AfterFunc(b.batchTimeout, func() {
				b.flush()
			})
		}
	}

	// Wait for result on this waiter's channels
	select {
	case result := <-w.responseCh:
		return result, nil
	case err := <-w.errorCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// flushLoop periodically flushes pending requests
func (b *Batcher) flushLoop() {
	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	for range ticker.C {
		b.flush()
	}
}

// flush processes all pending requests
func (b *Batcher) flush() {
	b.mu.Lock()
	if len(b.pending) == 0 {
		b.mu.Unlock()
		return
	}

	// Extract pending requests
	pending := b.pending
	b.pending = make(map[string]*BatchRequest)
	b.batchCount++
	b.mu.Unlock()

	// Collect keys
	keys := make([]string, 0, len(pending))
	for key := range pending {
		keys = append(keys, key)
	}

	// Process batch
	results, err := b.processFn(keys)

	// Send results to all waiting goroutines
	for key, req := range pending {
		for _, w := range req.waiters {
			if err != nil {
				w.errorCh <- err
			} else if result, ok := results[key]; ok {
				w.responseCh <- result
			} else {
				// No result found for this key - send nil error
				w.errorCh <- nil
			}
		}
	}
}

// GetMetrics returns batcher metrics
func (b *Batcher) GetMetrics() BatcherMetrics {
	b.mu.RLock()
	pendingCount := len(b.pending)
	b.mu.RUnlock()

	metrics := BatcherMetrics{
		BatchCount:     b.batchCount,
		RequestCount:   b.requestCount,
		CoalescedCount: b.coalescedCount,
		PendingCount:   int64(pendingCount),
	}

	if metrics.BatchCount > 0 {
		metrics.AvgBatchSize = float64(metrics.RequestCount) / float64(metrics.BatchCount)
	}

	return metrics
}

// BatcherMetrics holds batcher metrics
type BatcherMetrics struct {
	BatchCount     int64
	RequestCount   int64
	CoalescedCount int64
	PendingCount   int64
	AvgBatchSize   float64
}

// HealthCheckBatcher batches health check operations
type HealthCheckBatcher struct {
	batcher *Batcher
	pool    *ConnectionPool
}

// NewHealthCheckBatcher creates a health check batcher
func NewHealthCheckBatcher(config BatcherConfig, pool *ConnectionPool) *HealthCheckBatcher {
	hcb := &HealthCheckBatcher{
		pool: pool,
	}

	processFn := func(urls []string) (map[string]interface{}, error) {
		results := make(map[string]interface{})
		var wg sync.WaitGroup

		// Process health checks concurrently
		for _, urlStr := range urls {
			wg.Add(1)
			go func(u string) {
				defer wg.Done()

				// Parse URL
				parsedURL, err := parseURL(u)
				if err != nil {
					results[u] = false
					return
				}

				// Perform health check
				alive := isBackendAliveWithPool(parsedURL, pool, nil)
				results[u] = alive
			}(urlStr)
		}

		wg.Wait()
		return results, nil
	}

	hcb.batcher = NewBatcher(config, processFn)
	return hcb
}

// Check performs a batched health check
func (hcb *HealthCheckBatcher) Check(ctx context.Context, url string) (bool, error) {
	result, err := hcb.batcher.Submit(ctx, url)
	if err != nil {
		return false, err
	}

	alive, ok := result.(bool)
	if !ok {
		return false, nil
	}

	return alive, nil
}

// GetMetrics returns batcher metrics
func (hcb *HealthCheckBatcher) GetMetrics() BatcherMetrics {
	return hcb.batcher.GetMetrics()
}

// Helper function to parse URL from string
func parseURL(urlStr string) (*url.URL, error) {
	return url.Parse(urlStr)
}

// StatsBatcher batches statistics computation
type StatsBatcher struct {
	batcher *Batcher
	lb      *LoadBalancer
}

// NewStatsBatcher creates a stats batcher
func NewStatsBatcher(config BatcherConfig, lb *LoadBalancer) *StatsBatcher {
	sb := &StatsBatcher{
		lb: lb,
	}

	processFn := func(keys []string) (map[string]interface{}, error) {
		// Compute stats once for all requests
		stats := lb.GetStats()

		// Return same stats for all keys
		results := make(map[string]interface{})
		for _, key := range keys {
			results[key] = stats
		}

		return results, nil
	}

	sb.batcher = NewBatcher(config, processFn)
	return sb
}

// Get retrieves batched stats
func (sb *StatsBatcher) Get(ctx context.Context) ([]map[string]interface{}, error) {
	result, err := sb.batcher.Submit(ctx, "stats")
	if err != nil {
		return nil, err
	}

	stats, ok := result.([]map[string]interface{})
	if !ok {
		return nil, nil
	}

	return stats, nil
}

// GetMetrics returns batcher metrics
func (sb *StatsBatcher) GetMetrics() BatcherMetrics {
	return sb.batcher.GetMetrics()
}
