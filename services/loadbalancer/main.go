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
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.backends) == 0 {
		return nil
	}

	// Start from next index
	next := s.NextIndex()
	l := len(s.backends) + next

	for i := next; i < l; i++ {
		idx := i % len(s.backends)
		if s.backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.backends[idx]
		}
	}
	return nil
}

// HealthCheck pings the backends and updates the status
func (s *ServerPool) HealthCheck() {
	s.mu.RLock()
	backends := make([]*Backend, len(s.backends))
	copy(backends, s.backends)
	s.mu.RUnlock()

	for _, b := range backends {
		alive := isBackendAlive(b.URL)
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
	timeout := 2 * time.Second
	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(u.String() + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// LoadBalancer represents the load balancer
type LoadBalancer struct {
	serverPool *ServerPool
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		serverPool: &ServerPool{
			backends: []*Backend{},
		},
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
	return nil
}

// ServeHTTP handles incoming requests
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	peer := lb.serverPool.GetNextPeer()
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
			lb.serverPool.HealthCheck()
		}
	}()
}

// GetStats returns statistics about the backends
func (lb *LoadBalancer) GetStats() []map[string]interface{} {
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

func main() {
	lb = NewLoadBalancer()

	// Start health check every 10 seconds
	lb.StartHealthCheck(10 * time.Second)

	http.HandleFunc("/add-backend", addBackendHandler)
	http.HandleFunc("/stats", statsHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", lb.ServeHTTP)

	port := ":8082"
	log.Printf("Load balancer starting on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

