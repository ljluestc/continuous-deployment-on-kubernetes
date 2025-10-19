package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// URLMapping represents a URL shortening entry
type URLMapping struct {
	ShortURL    string    `json:"short_url"`
	LongURL     string    `json:"long_url"`
	CreatedAt   time.Time `json:"created_at"`
	AccessCount int64     `json:"access_count"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
}

// TinyURLService handles URL shortening operations
type TinyURLService struct {
	mu       sync.RWMutex
	mappings map[string]*URLMapping
	reverse  map[string]string // longURL -> shortURL for deduplication
	baseURL  string
}

// NewTinyURLService creates a new TinyURL service
func NewTinyURLService(baseURL string) *TinyURLService {
	return &TinyURLService{
		mappings: make(map[string]*URLMapping),
		reverse:  make(map[string]string),
		baseURL:  baseURL,
	}
}

// GenerateShortURL generates a short URL from a long URL
func (s *TinyURLService) GenerateShortURL(longURL string) string {
	hash := md5.Sum([]byte(longURL + time.Now().String()))
	return hex.EncodeToString(hash[:])[:8]
}

// CreateShortURL creates a new short URL
func (s *TinyURLService) CreateShortURL(longURL string, customAlias string, ttl time.Duration) (*URLMapping, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if long URL already exists
	if shortURL, exists := s.reverse[longURL]; exists {
		return s.mappings[shortURL], nil
	}

	var shortURL string
	if customAlias != "" {
		// Check if custom alias is available
		if _, exists := s.mappings[customAlias]; exists {
			return nil, fmt.Errorf("custom alias already exists")
		}
		shortURL = customAlias
	} else {
		shortURL = s.GenerateShortURL(longURL)
		// Handle collision
		for {
			if _, exists := s.mappings[shortURL]; !exists {
				break
			}
			shortURL = s.GenerateShortURL(longURL + time.Now().String())
		}
	}

	mapping := &URLMapping{
		ShortURL:    shortURL,
		LongURL:     longURL,
		CreatedAt:   time.Now(),
		AccessCount: 0,
	}

	if ttl > 0 {
		mapping.ExpiresAt = time.Now().Add(ttl)
	}

	s.mappings[shortURL] = mapping
	s.reverse[longURL] = shortURL

	return mapping, nil
}

// GetLongURL retrieves the long URL for a short URL
func (s *TinyURLService) GetLongURL(shortURL string) (*URLMapping, error) {
	s.mu.RLock()
	mapping, exists := s.mappings[shortURL]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("short URL not found")
	}

	// Check expiration
	if !mapping.ExpiresAt.IsZero() && time.Now().After(mapping.ExpiresAt) {
		s.mu.Lock()
		delete(s.mappings, shortURL)
		delete(s.reverse, mapping.LongURL)
		s.mu.Unlock()
		return nil, fmt.Errorf("short URL expired")
	}

	// Increment access count
	s.mu.Lock()
	mapping.AccessCount++
	s.mu.Unlock()

	return mapping, nil
}

// DeleteShortURL deletes a short URL
func (s *TinyURLService) DeleteShortURL(shortURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mapping, exists := s.mappings[shortURL]
	if !exists {
		return fmt.Errorf("short URL not found")
	}

	delete(s.mappings, shortURL)
	delete(s.reverse, mapping.LongURL)

	return nil
}

// GetStats returns statistics for a short URL
func (s *TinyURLService) GetStats(shortURL string) (*URLMapping, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mapping, exists := s.mappings[shortURL]
	if !exists {
		return nil, fmt.Errorf("short URL not found")
	}

	return mapping, nil
}

// ListAllMappings returns all URL mappings
func (s *TinyURLService) ListAllMappings() []*URLMapping {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mappings := make([]*URLMapping, 0, len(s.mappings))
	for _, mapping := range s.mappings {
		mappings = append(mappings, mapping)
	}

	return mappings
}

// HTTP Handlers

var service *TinyURLService

func createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		LongURL      string `json:"long_url"`
		CustomAlias  string `json:"custom_alias,omitempty"`
		TTLSeconds   int    `json:"ttl_seconds,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.LongURL == "" {
		http.Error(w, "long_url is required", http.StatusBadRequest)
		return
	}

	var ttl time.Duration
	if req.TTLSeconds > 0 {
		ttl = time.Duration(req.TTLSeconds) * time.Second
	}

	mapping, err := service.CreateShortURL(req.LongURL, req.CustomAlias, ttl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapping)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Path[1:] // Remove leading slash

	mapping, err := service.GetLongURL(shortURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.Redirect(w, r, mapping.LongURL, http.StatusMovedPermanently)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Query().Get("short_url")
	if shortURL == "" {
		http.Error(w, "short_url parameter is required", http.StatusBadRequest)
		return
	}

	mapping, err := service.GetStats(shortURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapping)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shortURL := r.URL.Query().Get("short_url")
	if shortURL == "" {
		http.Error(w, "short_url parameter is required", http.StatusBadRequest)
		return
	}

	if err := service.DeleteShortURL(shortURL); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	mappings := service.ListAllMappings()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mappings)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func main() {
	service = NewTinyURLService("http://localhost:8080")

	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/stats", statsHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", redirectHandler)

	port := ":8080"
	log.Printf("TinyURL service starting on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

