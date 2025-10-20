package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// Backend represents a backend server
type Backend struct {
	URL          *url.URL
	Alive        bool
	mu           sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
	FailCount    int64
	SuccessCount int64
}

// SetAlive sets the alive status of the backend
func (b *Backend) SetAlive(alive bool) {
	b.mu.Lock()
	b.Alive = alive
	b.mu.Unlock()
}

// IsAlive returns the alive status of the backend
func (b *Backend) IsAlive() bool {
	b.mu.RLock()
	alive := b.Alive
	b.mu.RUnlock()
	return alive
}

// ServerPool holds information about reachable backends
type ServerPool struct {
	backends []*Backend
	current  uint64
	mu       sync.RWMutex
}

// AddBackend adds a backend to the server pool
func (s *ServerPool) AddBackend(backend *Backend) {
	s.mu.Lock()
	s.backends = append(s.backends, backend)
	s.mu.Unlock()
}

// NextIndex atomically increases the counter and returns an index
func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

// GetNextPeer returns the next active peer using round-robin
func (s *ServerPool) GetNextPeer() *Backend {
	return s.GetNextPeerWithCache(nil)
}

// GetNextPeerWithCache returns next active peer using routing cache
func (s *ServerPool) GetNextPeerWithCache(routingCache *RoutingCache) *Backend {
	// Try cache first
	if routingCache != nil {
		if cached, found := routingCache.Get(); found && len(cached) > 0 {
			// Use cached active backends for faster selection
			next := int(atomic.AddUint64(&s.current, 1) % uint64(len(cached)))
			return cached[next]
		}
	}

	// Fallback to full scan
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.backends) == 0 {
		return nil
	}

	// Collect active backends
	var activeBackends []*Backend
	for _, b := range s.backends {
		if b.IsAlive() {
			activeBackends = append(activeBackends, b)
		}
	}

	if len(activeBackends) == 0 {
		return nil
	}

	// Update cache
	if routingCache != nil {
		routingCache.Set(activeBackends)
	}

	// Select from active backends
	next := int(atomic.AddUint64(&s.current, 1) % uint64(len(activeBackends)))
	return activeBackends[next]
}

// HealthCheck pings the backends and updates the status
func (s *ServerPool) HealthCheck() {
	s.HealthCheckWithCache(nil, nil)
}

// HealthCheckWithCache pings backends using connection pool and cache
func (s *ServerPool) HealthCheckWithCache(pool *ConnectionPool, healthCache *HealthCache) {
	s.mu.RLock()
	backends := make([]*Backend, len(s.backends))
	copy(backends, s.backends)
	s.mu.RUnlock()

	for _, b := range backends {
		alive := isBackendAliveWithPool(b.URL, pool, healthCache)
		b.SetAlive(alive)
		if alive {
			log.Printf("Backend %s is alive", b.URL)
		} else {
			log.Printf("Backend %s is down", b.URL)
		}
	}
}

// GetBackends returns all backends
func (s *ServerPool) GetBackends() []*Backend {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.backends
}

// isBackendAlive checks if a backend is alive
func isBackendAlive(u *url.URL) bool {
	return isBackendAliveWithPool(u, nil, nil)
}

// isBackendAliveWithPool checks if a backend is alive using connection pool and cache
func isBackendAliveWithPool(u *url.URL, pool *ConnectionPool, healthCache *HealthCache) bool {
	urlStr := u.String()

	// Check cache first
	if healthCache != nil {
		if alive, found := healthCache.Get(urlStr); found {
			return alive
		}
	}

	// Perform health check
	start := time.Now()
	timeout := 2 * time.Second

	var client *http.Client
	if pool != nil {
		client = pool.Get(u, timeout)
	} else {
		client = &http.Client{Timeout: timeout}
	}

	resp, err := client.Get(urlStr + "/health")
	latency := time.Since(start)

	alive := err == nil && resp != nil && resp.StatusCode == http.StatusOK
	if resp != nil {
		resp.Body.Close()
	}

	// Store in cache
	if healthCache != nil {
		healthCache.Set(urlStr, alive, latency)
	}

	return alive
}

// LoadBalancer represents the load balancer
type LoadBalancer struct {
	serverPool     *ServerPool
	cacheManager   *CacheManager
	connectionPool *ConnectionPool
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer() *LoadBalancer {
	cacheConfig := DefaultCacheConfig()
	poolConfig := PoolConfig{
		MaxIdleConns:    10,
		MaxLifetime:     60 * time.Second,
		IdleTimeout:     30 * time.Second,
		CleanupInterval: 10 * time.Second,
		RequestTimeout:  2 * time.Second,
	}

	return &LoadBalancer{
		serverPool: &ServerPool{
			backends: []*Backend{},
		},
		cacheManager:   NewCacheManager(cacheConfig),
		connectionPool: NewConnectionPool(poolConfig),
	}
}

// AddBackend adds a backend to the load balancer
func (lb *LoadBalancer) AddBackend(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(u)
	backend := &Backend{
		URL:          u,
		Alive:        true,
		ReverseProxy: proxy,
	}

	lb.serverPool.AddBackend(backend)

	// Invalidate caches when backend is added
	lb.cacheManager.Routing().Invalidate()
	lb.cacheManager.Stats().Invalidate()

	return nil
}

// ServeHTTP handles incoming requests
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	peer := lb.serverPool.GetNextPeerWithCache(lb.cacheManager.Routing())
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		atomic.AddInt64(&peer.SuccessCount, 1)
		return
	}

	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

// StartHealthCheck starts the health check routine
func (lb *LoadBalancer) StartHealthCheck(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			lb.serverPool.HealthCheckWithCache(lb.connectionPool, lb.cacheManager.Health())
			// Invalidate routing cache after health check
			lb.cacheManager.Routing().Invalidate()
		}
	}()
}

// GetStats returns statistics about the backends
func (lb *LoadBalancer) GetStats() []map[string]interface{} {
	// Try cache first
	if cached, found := lb.cacheManager.Stats().Get(); found {
		return cached
	}

	// Compute stats
	backends := lb.serverPool.GetBackends()
	stats := make([]map[string]interface{}, len(backends))

	for i, b := range backends {
		stats[i] = map[string]interface{}{
			"url":           b.URL.String(),
			"alive":         b.IsAlive(),
			"success_count": atomic.LoadInt64(&b.SuccessCount),
			"fail_count":    atomic.LoadInt64(&b.FailCount),
		}
	}

	// Cache the result
	lb.cacheManager.Stats().Set(stats)

	return stats
}

var lb *LoadBalancer

func addBackendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := lb.AddBackend(req.URL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	stats := lb.GetStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func cacheMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	metrics := map[string]interface{}{
		"cache_metrics": lb.cacheManager.GetAllMetrics(),
		"pool_metrics":  lb.connectionPool.GetMetrics(),
	}

	json.NewEncoder(w).Encode(metrics)
}

func main() {
	lb = NewLoadBalancer()

	// Start health check every 10 seconds
	lb.StartHealthCheck(10 * time.Second)

	http.HandleFunc("/add-backend", addBackendHandler)
	http.HandleFunc("/stats", statsHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/cache-metrics", cacheMetricsHandler)
	http.HandleFunc("/", lb.ServeHTTP)

	port := ":8082"
	log.Printf("Load balancer starting on %s", port)
	log.Printf("Caching enabled - Health: %v, Stats: %v, Routing: %v",
		lb.cacheManager.config.HealthCacheEnabled,
		lb.cacheManager.config.StatsCacheEnabled,
		lb.cacheManager.config.RoutingCacheEnabled)
	log.Fatal(http.ListenAndServe(port, nil))
}

