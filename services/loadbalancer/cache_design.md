# Load Balancer Caching Layer Design

## Overview
This document outlines the design for a multi-level caching system to optimize the load balancer's performance by reducing redundant health checks, expensive stat computations, and improving routing efficiency.

## Current Performance Bottlenecks

### 1. Health Check Overhead
- **Issue**: Health checks run every 10s for all backends (`isBackendAlive()` at line 109)
- **Cost**: Each check creates new HTTP client, performs network request with 2s timeout
- **Impact**: N backends × health check every 10s = high overhead for large pools

### 2. Statistics Computation
- **Issue**: `GetStats()` at line 179 computes stats on-demand with atomic loads
- **Cost**: 4 atomic operations per backend per stats request
- **Impact**: High contention on stats endpoints

### 3. Backend Selection
- **Issue**: `GetNextPeer()` at line 58 traverses backends linearly on every request
- **Cost**: O(n) worst case when backends are down
- **Impact**: Routing latency increases with backend count

## Proposed Caching Architecture

### Layer 1: Health Status Cache

```go
type HealthCache struct {
    mu          sync.RWMutex
    entries     map[string]*HealthCacheEntry  // URL -> entry
    ttl         time.Duration                  // Cache TTL (default: 5s)
    maxAge      time.Duration                  // Max age before forced refresh
}

type HealthCacheEntry struct {
    alive       bool
    lastCheck   time.Time
    checkCount  int64
    errorCount  int64
    avgLatency  time.Duration
}
```

**Features:**
- Cache health status for configurable TTL (default: 5 seconds)
- Reduce redundant health checks
- Store health check metrics (latency, error rate)
- Automatic cache invalidation on TTL expiry
- Thread-safe with read-write locks

**Benefits:**
- Reduces health check frequency by up to 50%
- Provides health trend data for smarter routing
- Enables fast-fail for consistently unhealthy backends

### Layer 2: Statistics Cache

```go
type StatsCache struct {
    mu          sync.RWMutex
    snapshot    []map[string]interface{}  // Cached stats
    lastUpdate  time.Time
    ttl         time.Duration              // Cache TTL (default: 1s)
    dirty       bool                       // Invalidation flag
}
```

**Features:**
- Cache computed statistics for 1-2 seconds
- Lazy invalidation on backend changes
- Aggregated metrics (total requests, error rate, avg latency)
- Snapshot consistency for reporting

**Benefits:**
- Reduces atomic operations by 95%+ for high-traffic stats endpoints
- Provides consistent stats snapshots
- Enables dashboard-friendly aggregations

### Layer 3: Routing Cache (Active Backend List)

```go
type RoutingCache struct {
    mu             sync.RWMutex
    activeBackends []*Backend           // Cached alive backends
    lastUpdate     time.Time
    ttl            time.Duration        // Cache TTL (default: 2s)
    version        uint64               // Cache version for invalidation
}
```

**Features:**
- Pre-filtered list of alive backends
- Round-robin index maintained for active list only
- Automatic rebuild on backend changes or health status changes
- Version-based cache invalidation

**Benefits:**
- O(1) backend selection instead of O(n)
- Eliminates per-request IsAlive() checks
- Reduces lock contention in hot path

### Layer 4: Connection Pool Cache

```go
type ConnectionPool struct {
    mu          sync.RWMutex
    connections map[string]*PooledConnection  // URL -> connection
    maxIdle     int                            // Max idle connections per backend
    maxLifetime time.Duration                  // Max connection lifetime
}

type PooledConnection struct {
    client      *http.Client
    lastUsed    time.Time
    useCount    int64
    created     time.Time
}
```

**Features:**
- Reusable HTTP clients for health checks
- Connection lifecycle management
- Idle connection timeout
- Per-backend connection limits

**Benefits:**
- Eliminates HTTP client creation overhead (line 111-113)
- Reduces TCP handshake overhead
- Improves health check latency

## Cache Invalidation Strategy

### 1. Time-Based Invalidation (TTL)
- Health cache: 5 seconds
- Stats cache: 1-2 seconds
- Routing cache: 2 seconds
- Connection pool: 60 seconds idle timeout

### 2. Event-Based Invalidation
- Backend added → invalidate routing cache
- Backend removed → invalidate all caches
- Health status changed → invalidate routing cache
- Manual flush via admin endpoint

### 3. Version-Based Invalidation
- Each cache maintains version number
- Backend pool changes increment global version
- Caches check version before returning cached data

## Implementation Plan

### Phase 1: Health Status Cache
**Files to modify:**
- `services/loadbalancer/cache.go` (new file)
- `services/loadbalancer/main.go` (integrate cache)

**Key changes:**
- Add `HealthCache` struct and methods
- Modify `isBackendAlive()` to check cache first
- Update `HealthCheck()` to populate cache
- Add cache metrics

### Phase 2: Statistics Cache
**Files to modify:**
- `services/loadbalancer/cache.go` (extend)
- `services/loadbalancer/main.go` (update GetStats)

**Key changes:**
- Add `StatsCache` struct
- Modify `GetStats()` to use cache
- Add cache invalidation on backend changes
- Add aggregated metrics

### Phase 3: Routing Cache
**Files to modify:**
- `services/loadbalancer/cache.go` (extend)
- `services/loadbalancer/main.go` (update GetNextPeer)

**Key changes:**
- Add `RoutingCache` struct
- Pre-filter active backends
- Optimize round-robin selection
- Add version-based invalidation

### Phase 4: Connection Pool
**Files to modify:**
- `services/loadbalancer/pool.go` (new file)
- `services/loadbalancer/main.go` (integrate pool)

**Key changes:**
- Add `ConnectionPool` struct
- Reuse HTTP clients for health checks
- Add connection lifecycle management
- Add pool metrics

## Cache Configuration

```go
type CacheConfig struct {
    // Health cache settings
    HealthCacheTTL      time.Duration  // Default: 5s
    HealthCacheEnabled  bool           // Default: true

    // Stats cache settings
    StatsCacheTTL       time.Duration  // Default: 1s
    StatsCacheEnabled   bool           // Default: true

    // Routing cache settings
    RoutingCacheTTL     time.Duration  // Default: 2s
    RoutingCacheEnabled bool           // Default: true

    // Connection pool settings
    PoolMaxIdleConns    int            // Default: 10
    PoolMaxLifetime     time.Duration  // Default: 60s
    PoolEnabled         bool           // Default: true
}
```

## Performance Targets

### Current Baseline (from performance tests)
- Request routing: ~50ms average
- Health check: 2s timeout per backend
- Stats computation: 4 atomic ops × N backends

### Target Improvements
- **Request routing**: <10ms (80% reduction)
- **Health check overhead**: 50% reduction via caching
- **Stats endpoint**: <5ms (90% reduction for cached responses)
- **Cache hit rate**: >90% for all caches

## Monitoring and Observability

### Cache Metrics to Track
```go
type CacheMetrics struct {
    HitCount      int64  // Cache hits
    MissCount     int64  // Cache misses
    EvictionCount int64  // Cache evictions
    Size          int64  // Current cache size
    HitRate       float64 // Hit rate percentage
}
```

### Prometheus Metrics
- `cache_hit_total{cache="health|stats|routing"}`
- `cache_miss_total{cache="health|stats|routing"}`
- `cache_size{cache="health|stats|routing"}`
- `cache_eviction_total{cache="health|stats|routing"}`
- `cache_hit_rate{cache="health|stats|routing"}`

### Logging
- Cache hit/miss at DEBUG level
- Cache eviction at INFO level
- Cache configuration at startup (INFO)
- Cache invalidation events at DEBUG level

## Thread Safety Guarantees

### Lock Hierarchy (to prevent deadlocks)
1. ServerPool.mu (outermost)
2. Backend.mu
3. Cache.mu (innermost)

### Atomic Operations
- Cache hit/miss counters (lock-free)
- Cache version numbers (lock-free)
- Backend success/fail counts (existing, maintained)

### Race Condition Prevention
- Use `sync.RWMutex` for read-heavy operations
- Atomic operations for counters
- Copy-on-read for cached slices
- Version checks before cache use

## Backward Compatibility

### Configuration
- All caching disabled by default (opt-in via config/flags)
- Existing behavior preserved when caching disabled
- Graceful degradation if cache initialization fails

### API
- No changes to external HTTP endpoints
- Internal API maintains existing signatures
- Cache methods are additive, not replacing

## Testing Strategy

### Unit Tests
- Cache hit/miss behavior
- TTL expiration
- Version-based invalidation
- Concurrent access safety
- Memory leak prevention

### Integration Tests
- End-to-end routing with cache
- Health check with cache
- Stats endpoint with cache
- Cache warming on startup

### Performance Tests
- Routing latency (cached vs uncached)
- Stats endpoint throughput
- Health check overhead reduction
- Cache memory usage
- Concurrent request handling

### Benchmark Tests
```go
BenchmarkGetNextPeer_NoCache
BenchmarkGetNextPeer_WithCache
BenchmarkGetStats_NoCache
BenchmarkGetStats_WithCache
BenchmarkHealthCheck_NoCache
BenchmarkHealthCheck_WithCache
```

## Security Considerations

### Cache Poisoning Prevention
- Validate health check responses before caching
- Limit cache size to prevent memory exhaustion
- Sanitize cache keys (URLs)

### Timing Attacks
- Avoid leaking cache state via timing
- Consistent response times for hit/miss (if required)

### Resource Limits
- Max cache entries per type (default: 1000)
- Max memory usage per cache (default: 10MB)
- Automatic eviction on limit

## Migration Path

### Phase 1: Development (Week 1)
- Implement cache structs and basic operations
- Unit tests for cache operations
- Integration with existing code (feature-flagged)

### Phase 2: Testing (Week 2)
- Performance benchmarking
- Load testing with cache enabled
- Memory profiling
- Fix issues

### Phase 3: Deployment (Week 3)
- Deploy with caching disabled
- Enable caching for 10% traffic
- Monitor metrics
- Gradual rollout to 100%

### Phase 4: Optimization (Week 4)
- Tune TTL values based on metrics
- Optimize cache hit rates
- Add advanced features (LRU, adaptive TTL)

## Future Enhancements

### Adaptive TTL
- Adjust TTL based on backend stability
- Shorter TTL for flapping backends
- Longer TTL for stable backends

### LRU Eviction
- Replace time-based eviction with LRU
- Better memory utilization
- Configurable max cache size

### Predictive Health Checks
- Learn backend health patterns
- Skip health checks for highly stable backends
- Increase checks for unstable backends

### Distributed Caching
- Share cache across multiple load balancer instances
- Redis/Memcached integration
- Cache coherency protocol

## Conclusion

This caching layer design provides:
- **50-80% performance improvement** for routing and stats
- **50% reduction** in health check overhead
- **90%+ cache hit rates** under normal operation
- **Backward compatible** with existing deployment
- **Observable** via comprehensive metrics
- **Thread-safe** with proven concurrency patterns

The phased implementation approach ensures safe rollout with continuous validation at each step.
