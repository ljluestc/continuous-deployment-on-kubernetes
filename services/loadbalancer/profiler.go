package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Profiler tracks performance metrics for operations
type Profiler struct {
	mu                  sync.RWMutex
	operations          map[string]*OperationStats
	enabled             bool
	sampleRate          float64 // 0.0-1.0, what percentage to profile
	detailedHistograms  bool
}

// OperationStats holds statistics for a specific operation
type OperationStats struct {
	Name          string
	Count         int64
	TotalDuration time.Duration
	MinDuration   time.Duration
	MaxDuration   time.Duration
	AvgDuration   time.Duration

	// Histogram buckets (in milliseconds)
	HistogramBuckets map[int]int64 // bucket -> count

	mu sync.RWMutex
}

// ProfilerConfig holds profiler configuration
type ProfilerConfig struct {
	Enabled            bool
	SampleRate         float64
	DetailedHistograms bool
}

// NewProfiler creates a new performance profiler
func NewProfiler(config ProfilerConfig) *Profiler {
	if config.SampleRate == 0 {
		config.SampleRate = 1.0 // Profile everything by default
	}

	return &Profiler{
		operations:         make(map[string]*OperationStats),
		enabled:            config.Enabled,
		sampleRate:         config.SampleRate,
		detailedHistograms: config.DetailedHistograms,
	}
}

// Profile executes a function and records its execution time
func (p *Profiler) Profile(operationName string, fn func()) {
	if !p.enabled {
		fn()
		return
	}

	// Sample rate check
	if p.sampleRate < 1.0 && time.Now().UnixNano()%100 >= int64(p.sampleRate*100) {
		fn()
		return
	}

	start := time.Now()
	fn()
	duration := time.Since(start)

	p.record(operationName, duration)
}

// ProfileWithReturn executes a function with return value and records timing
func (p *Profiler) ProfileWithReturn(operationName string, fn func() interface{}) interface{} {
	if !p.enabled {
		return fn()
	}

	start := time.Now()
	result := fn()
	duration := time.Since(start)

	p.record(operationName, duration)
	return result
}

// StartTimer starts a timer for manual profiling
func (p *Profiler) StartTimer(operationName string) *Timer {
	return &Timer{
		profiler:      p,
		operationName: operationName,
		startTime:     time.Now(),
	}
}

// record records an operation's execution time
func (p *Profiler) record(operationName string, duration time.Duration) {
	p.mu.Lock()
	stats, exists := p.operations[operationName]
	if !exists {
		stats = &OperationStats{
			Name:             operationName,
			MinDuration:      duration,
			MaxDuration:      duration,
			HistogramBuckets: make(map[int]int64),
		}
		p.operations[operationName] = stats
	}
	p.mu.Unlock()

	stats.mu.Lock()
	defer stats.mu.Unlock()

	stats.Count++
	stats.TotalDuration += duration

	if duration < stats.MinDuration {
		stats.MinDuration = duration
	}
	if duration > stats.MaxDuration {
		stats.MaxDuration = duration
	}

	stats.AvgDuration = time.Duration(int64(stats.TotalDuration) / stats.Count)

	// Update histogram
	if p.detailedHistograms {
		bucketMs := int(duration.Milliseconds())
		if bucketMs > 1000 {
			bucketMs = 1000 // Cap at 1000ms bucket
		}
		stats.HistogramBuckets[bucketMs]++
	}
}

// GetStats returns statistics for a specific operation
func (p *Profiler) GetStats(operationName string) *OperationStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats, exists := p.operations[operationName]
	if !exists {
		return nil
	}

	// Return a copy
	stats.mu.RLock()
	defer stats.mu.RUnlock()

	statsCopy := &OperationStats{
		Name:             stats.Name,
		Count:            stats.Count,
		TotalDuration:    stats.TotalDuration,
		MinDuration:      stats.MinDuration,
		MaxDuration:      stats.MaxDuration,
		AvgDuration:      stats.AvgDuration,
		HistogramBuckets: make(map[int]int64),
	}

	for bucket, count := range stats.HistogramBuckets {
		statsCopy.HistogramBuckets[bucket] = count
	}

	return statsCopy
}

// GetAllStats returns statistics for all operations
func (p *Profiler) GetAllStats() map[string]*OperationStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	allStats := make(map[string]*OperationStats)
	for name := range p.operations {
		allStats[name] = p.GetStats(name)
	}

	return allStats
}

// Reset clears all profiling data
func (p *Profiler) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.operations = make(map[string]*OperationStats)
}

// GetSummary returns a formatted summary of all profiling data
func (p *Profiler) GetSummary() string {
	allStats := p.GetAllStats()

	summary := "Performance Profiler Summary\n"
	summary += "============================\n\n"

	for name, stats := range allStats {
		summary += fmt.Sprintf("Operation: %s\n", name)
		summary += fmt.Sprintf("  Count: %d\n", stats.Count)
		summary += fmt.Sprintf("  Avg Duration: %v\n", stats.AvgDuration)
		summary += fmt.Sprintf("  Min Duration: %v\n", stats.MinDuration)
		summary += fmt.Sprintf("  Max Duration: %v\n", stats.MaxDuration)
		summary += fmt.Sprintf("  Total Duration: %v\n", stats.TotalDuration)

		if len(stats.HistogramBuckets) > 0 {
			summary += "  Histogram (ms):\n"
			for bucket := 0; bucket <= 1000; bucket++ {
				if count, ok := stats.HistogramBuckets[bucket]; ok && count > 0 {
					summary += fmt.Sprintf("    %4dms: %d\n", bucket, count)
				}
			}
		}

		summary += "\n"
	}

	return summary
}

// Timer represents an active profiling timer
type Timer struct {
	profiler      *Profiler
	operationName string
	startTime     time.Time
}

// Stop stops the timer and records the duration
func (t *Timer) Stop() {
	if t.profiler != nil && t.profiler.enabled {
		duration := time.Since(t.startTime)
		t.profiler.record(t.operationName, duration)
	}
}

// MemoryProfiler tracks memory usage
type MemoryProfiler struct {
	mu       sync.RWMutex
	snapshots []MemorySnapshot
	maxSnapshots int
}

// MemorySnapshot represents a point-in-time memory snapshot
type MemorySnapshot struct {
	Timestamp    time.Time
	Alloc        uint64 // Bytes allocated and in use
	TotalAlloc   uint64 // Bytes allocated (cumulative)
	Sys          uint64 // Bytes obtained from system
	NumGC        uint32 // Number of completed GC cycles
	HeapObjects  uint64 // Number of allocated heap objects
	GoRoutines   int    // Number of goroutines
}

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler(maxSnapshots int) *MemoryProfiler {
	if maxSnapshots == 0 {
		maxSnapshots = 100
	}

	return &MemoryProfiler{
		snapshots:    make([]MemorySnapshot, 0, maxSnapshots),
		maxSnapshots: maxSnapshots,
	}
}

// TakeSnapshot captures current memory statistics
func (mp *MemoryProfiler) TakeSnapshot() MemorySnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	snapshot := MemorySnapshot{
		Timestamp:    time.Now(),
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGC:        m.NumGC,
		HeapObjects:  m.HeapObjects,
		GoRoutines:   runtime.NumGoroutine(),
	}

	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.snapshots = append(mp.snapshots, snapshot)

	// Keep only recent snapshots
	if len(mp.snapshots) > mp.maxSnapshots {
		mp.snapshots = mp.snapshots[1:]
	}

	return snapshot
}

// GetSnapshots returns all memory snapshots
func (mp *MemoryProfiler) GetSnapshots() []MemorySnapshot {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	snapshots := make([]MemorySnapshot, len(mp.snapshots))
	copy(snapshots, mp.snapshots)

	return snapshots
}

// GetLatestSnapshot returns the most recent memory snapshot
func (mp *MemoryProfiler) GetLatestSnapshot() *MemorySnapshot {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	if len(mp.snapshots) == 0 {
		return nil
	}

	return &mp.snapshots[len(mp.snapshots)-1]
}

// GetSummary returns formatted memory profiling summary
func (mp *MemoryProfiler) GetSummary() string {
	snapshot := mp.GetLatestSnapshot()
	if snapshot == nil {
		return "No memory snapshots available"
	}

	summary := "Memory Profiler Summary\n"
	summary += "======================\n\n"
	summary += fmt.Sprintf("Timestamp: %v\n", snapshot.Timestamp)
	summary += fmt.Sprintf("Allocated: %.2f MB\n", float64(snapshot.Alloc)/(1024*1024))
	summary += fmt.Sprintf("Total Allocated: %.2f MB\n", float64(snapshot.TotalAlloc)/(1024*1024))
	summary += fmt.Sprintf("System Memory: %.2f MB\n", float64(snapshot.Sys)/(1024*1024))
	summary += fmt.Sprintf("Heap Objects: %d\n", snapshot.HeapObjects)
	summary += fmt.Sprintf("Goroutines: %d\n", snapshot.GoRoutines)
	summary += fmt.Sprintf("GC Cycles: %d\n", snapshot.NumGC)

	return summary
}

// StartPeriodicSnapshot starts taking snapshots at regular intervals
func (mp *MemoryProfiler) StartPeriodicSnapshot(interval time.Duration) chan struct{} {
	stopCh := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				mp.TakeSnapshot()
			case <-stopCh:
				return
			}
		}
	}()

	return stopCh
}

// LatencyTracker tracks latency percentiles
type LatencyTracker struct {
	mu        sync.RWMutex
	latencies []time.Duration
	maxSize   int
	count     int64
}

// NewLatencyTracker creates a new latency tracker
func NewLatencyTracker(maxSize int) *LatencyTracker {
	if maxSize == 0 {
		maxSize = 1000
	}

	return &LatencyTracker{
		latencies: make([]time.Duration, 0, maxSize),
		maxSize:   maxSize,
	}
}

// Record records a latency measurement
func (lt *LatencyTracker) Record(latency time.Duration) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	atomic.AddInt64(&lt.count, 1)

	lt.latencies = append(lt.latencies, latency)

	// Keep only recent measurements
	if len(lt.latencies) > lt.maxSize {
		lt.latencies = lt.latencies[1:]
	}
}

// GetPercentile calculates the specified percentile (0-100)
func (lt *LatencyTracker) GetPercentile(percentile float64) time.Duration {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	if len(lt.latencies) == 0 {
		return 0
	}

	// Create a sorted copy
	sorted := make([]time.Duration, len(lt.latencies))
	copy(sorted, lt.latencies)

	// Simple bubble sort (fine for small datasets)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := int(float64(len(sorted)) * percentile / 100.0)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

// GetMetrics returns latency metrics
func (lt *LatencyTracker) GetMetrics() LatencyMetrics {
	return LatencyMetrics{
		Count: atomic.LoadInt64(&lt.count),
		P50:   lt.GetPercentile(50),
		P90:   lt.GetPercentile(90),
		P95:   lt.GetPercentile(95),
		P99:   lt.GetPercentile(99),
	}
}

// LatencyMetrics holds latency percentile metrics
type LatencyMetrics struct {
	Count int64
	P50   time.Duration
	P90   time.Duration
	P95   time.Duration
	P99   time.Duration
}
