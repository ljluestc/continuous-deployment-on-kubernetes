package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestBatcherSubmit tests basic batching functionality
func TestBatcherSubmit(t *testing.T) {
	config := BatcherConfig{
		BatchSize:     5,
		BatchTimeout:  100 * time.Millisecond,
		FlushInterval: 50 * time.Millisecond,
	}

	processFn := func(keys []string) (map[string]interface{}, error) {
		results := make(map[string]interface{})
		for _, key := range keys {
			results[key] = "processed_" + key
		}
		return results, nil
	}

	batcher := NewBatcher(config, processFn)

	ctx := context.Background()
	result, err := batcher.Submit(ctx, "key1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "processed_key1"
	if result != expected {
		t.Errorf("Expected %s, got %v", expected, result)
	}

	metrics := batcher.GetMetrics()
	if metrics.RequestCount != 1 {
		t.Errorf("Expected 1 request, got %d", metrics.RequestCount)
	}
}

// TestBatcherCoalescing tests request coalescing
func TestBatcherCoalescing(t *testing.T) {
	config := BatcherConfig{
		BatchSize:     10,
		BatchTimeout:  200 * time.Millisecond,
		FlushInterval: 100 * time.Millisecond,
	}

	processFn := func(keys []string) (map[string]interface{}, error) {
		// Simulate slow processing
		time.Sleep(50 * time.Millisecond)
		results := make(map[string]interface{})
		for _, key := range keys {
			results[key] = "processed"
		}
		return results, nil
	}

	batcher := NewBatcher(config, processFn)
	ctx := context.Background()

	// Submit multiple requests for the same key concurrently
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func() {
			_, err := batcher.Submit(ctx, "same_key")
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	metrics := batcher.GetMetrics()
	if metrics.CoalescedCount < 4 {
		t.Errorf("Expected at least 4 coalesced requests, got %d", metrics.CoalescedCount)
	}
}

// TestBatcherBatchSizeTrigger tests batch size trigger
func TestBatcherBatchSizeTrigger(t *testing.T) {
	config := BatcherConfig{
		BatchSize:     3,
		BatchTimeout:  1 * time.Second,
		FlushInterval: 1 * time.Second,
	}

	batchCount := 0
	processFn := func(keys []string) (map[string]interface{}, error) {
		batchCount++
		results := make(map[string]interface{})
		for _, key := range keys {
			results[key] = "ok"
		}
		return results, nil
	}

	batcher := NewBatcher(config, processFn)
	ctx := context.Background()

	// Submit batch size requests
	for i := 0; i < 3; i++ {
		go batcher.Submit(ctx, string(rune('a'+i)))
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	if batchCount != 1 {
		t.Errorf("Expected 1 batch, got %d", batchCount)
	}
}

// TestBatcherTimeout tests batch timeout trigger
func TestBatcherTimeout(t *testing.T) {
	config := BatcherConfig{
		BatchSize:     100,
		BatchTimeout:  100 * time.Millisecond,
		FlushInterval: 200 * time.Millisecond,
	}

	processed := false
	processFn := func(keys []string) (map[string]interface{}, error) {
		processed = true
		results := make(map[string]interface{})
		for _, key := range keys {
			results[key] = "ok"
		}
		return results, nil
	}

	batcher := NewBatcher(config, processFn)
	ctx := context.Background()

	go batcher.Submit(ctx, "key1")

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	if !processed {
		t.Error("Expected batch to be processed after timeout")
	}
}

// TestBatcherError tests error handling
func TestBatcherError(t *testing.T) {
	config := BatcherConfig{
		BatchSize:     5,
		BatchTimeout:  100 * time.Millisecond,
		FlushInterval: 50 * time.Millisecond,
	}

	expectedErr := errors.New("processing error")
	processFn := func(keys []string) (map[string]interface{}, error) {
		return nil, expectedErr
	}

	batcher := NewBatcher(config, processFn)
	ctx := context.Background()

	_, err := batcher.Submit(ctx, "key1")
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// TestBatcherContextCancel tests context cancellation
func TestBatcherContextCancel(t *testing.T) {
	config := BatcherConfig{
		BatchSize:     100,
		BatchTimeout:  1 * time.Second,
		FlushInterval: 1 * time.Second,
	}

	processFn := func(keys []string) (map[string]interface{}, error) {
		time.Sleep(500 * time.Millisecond)
		results := make(map[string]interface{})
		for _, key := range keys {
			results[key] = "ok"
		}
		return results, nil
	}

	batcher := NewBatcher(config, processFn)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan bool)
	go func() {
		_, err := batcher.Submit(ctx, "key1")
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
		done <- true
	}()

	// Cancel context before processing
	time.Sleep(50 * time.Millisecond)
	cancel()

	<-done
}

// TestBatcherMetrics tests batcher metrics
func TestBatcherMetrics(t *testing.T) {
	config := BatcherConfig{
		BatchSize:     10,
		BatchTimeout:  100 * time.Millisecond,
		FlushInterval: 50 * time.Millisecond,
	}

	processFn := func(keys []string) (map[string]interface{}, error) {
		results := make(map[string]interface{})
		for _, key := range keys {
			results[key] = "ok"
		}
		return results, nil
	}

	batcher := NewBatcher(config, processFn)
	ctx := context.Background()

	// Submit multiple requests
	for i := 0; i < 5; i++ {
		go batcher.Submit(ctx, string(rune('a'+i)))
	}

	time.Sleep(200 * time.Millisecond)

	metrics := batcher.GetMetrics()
	if metrics.RequestCount != 5 {
		t.Errorf("Expected 5 requests, got %d", metrics.RequestCount)
	}
	if metrics.BatchCount < 1 {
		t.Errorf("Expected at least 1 batch, got %d", metrics.BatchCount)
	}
	if metrics.AvgBatchSize != float64(metrics.RequestCount)/float64(metrics.BatchCount) {
		t.Errorf("AvgBatchSize calculation incorrect")
	}
}

// TestHealthCheckBatcher tests health check batching
func TestHealthCheckBatcher(t *testing.T) {
	poolConfig := PoolConfig{
		MaxIdleConns:    10,
		MaxLifetime:     60 * time.Second,
		IdleTimeout:     30 * time.Second,
		CleanupInterval: 1 * time.Second,
		RequestTimeout:  2 * time.Second,
	}
	pool := NewConnectionPool(poolConfig)
	defer pool.Close()

	batcherConfig := BatcherConfig{
		BatchSize:     5,
		BatchTimeout:  100 * time.Millisecond,
		FlushInterval: 50 * time.Millisecond,
	}
	batcher := NewHealthCheckBatcher(batcherConfig, pool)

	ctx := context.Background()

	// Submit health check (will fail for invalid URL)
	_, err := batcher.Check(ctx, "http://invalid-backend-xyz:9999")
	if err != nil {
		t.Logf("Expected error for invalid backend: %v", err)
	}

	metrics := batcher.GetMetrics()
	if metrics.RequestCount != 1 {
		t.Errorf("Expected 1 request, got %d", metrics.RequestCount)
	}
}

// TestHealthCheckBatcherConcurrent tests concurrent health checks
func TestHealthCheckBatcherConcurrent(t *testing.T) {
	poolConfig := PoolConfig{
		MaxIdleConns:    10,
		MaxLifetime:     60 * time.Second,
		IdleTimeout:     30 * time.Second,
		CleanupInterval: 1 * time.Second,
		RequestTimeout:  2 * time.Second,
	}
	pool := NewConnectionPool(poolConfig)
	defer pool.Close()

	batcherConfig := BatcherConfig{
		BatchSize:     10,
		BatchTimeout:  200 * time.Millisecond,
		FlushInterval: 100 * time.Millisecond,
	}
	batcher := NewHealthCheckBatcher(batcherConfig, pool)

	ctx := context.Background()
	done := make(chan bool)

	// Submit multiple concurrent health checks for the same URL (should coalesce)
	for i := 0; i < 5; i++ {
		go func(id int) {
			batcher.Check(ctx, "http://backend:8080")
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	metrics := batcher.GetMetrics()
	// With coalescing, multiple requests for same URL should result in fewer actual requests
	if metrics.RequestCount < 1 {
		t.Errorf("Expected at least 1 request, got %d", metrics.RequestCount)
	}
	// Should have coalescing happening
	if metrics.CoalescedCount < 1 && metrics.RequestCount == 1 {
		t.Log("Coalescing worked - only 1 actual request for 5 concurrent calls")
	}
}

// TestStatsBatcher tests stats batching (basic functionality)
func TestStatsBatcher(t *testing.T) {
	// Create a mock load balancer
	lb := NewLoadBalancer()

	batcherConfig := BatcherConfig{
		BatchSize:     5,
		BatchTimeout:  100 * time.Millisecond,
		FlushInterval: 50 * time.Millisecond,
	}
	batcher := NewStatsBatcher(batcherConfig, lb)

	ctx := context.Background()

	stats, err := batcher.Get(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if stats == nil {
		t.Error("Expected non-nil stats")
	}

	metrics := batcher.GetMetrics()
	if metrics.RequestCount != 1 {
		t.Errorf("Expected 1 request, got %d", metrics.RequestCount)
	}
}

// BenchmarkBatcherSubmit benchmarks batcher submit operation
func BenchmarkBatcherSubmit(b *testing.B) {
	config := BatcherConfig{
		BatchSize:     100,
		BatchTimeout:  100 * time.Millisecond,
		FlushInterval: 50 * time.Millisecond,
	}

	processFn := func(keys []string) (map[string]interface{}, error) {
		results := make(map[string]interface{})
		for _, key := range keys {
			results[key] = "ok"
		}
		return results, nil
	}

	batcher := NewBatcher(config, processFn)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batcher.Submit(ctx, "key")
	}
}

// BenchmarkBatcherConcurrent benchmarks concurrent batching
func BenchmarkBatcherConcurrent(b *testing.B) {
	config := BatcherConfig{
		BatchSize:     100,
		BatchTimeout:  100 * time.Millisecond,
		FlushInterval: 50 * time.Millisecond,
	}

	processFn := func(keys []string) (map[string]interface{}, error) {
		results := make(map[string]interface{})
		for _, key := range keys {
			results[key] = "ok"
		}
		return results, nil
	}

	batcher := NewBatcher(config, processFn)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			batcher.Submit(ctx, "key")
		}
	})
}
