package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// DNSRecord represents a DNS record
type DNSRecord struct {
	Domain    string    `json:"domain"`
	IPAddress string    `json:"ip_address"`
	Type      string    `json:"type"` // A, AAAA, CNAME, MX, etc.
	TTL       int       `json:"ttl"`
	CreatedAt time.Time `json:"created_at"`
}

// DNSService manages DNS records
type DNSService struct {
	mu      sync.RWMutex
	records map[string]*DNSRecord // domain -> record
	cache   map[string]*cacheEntry
}

type cacheEntry struct {
	record    *DNSRecord
	expiresAt time.Time
}

// NewDNSService creates a new DNS service
func NewDNSService() *DNSService {
	return &DNSService{
		records: make(map[string]*DNSRecord),
		cache:   make(map[string]*cacheEntry),
	}
}

// AddRecord adds a DNS record
func (s *DNSService) AddRecord(domain, ipAddress, recordType string, ttl int) (*DNSRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := &DNSRecord{
		Domain:    domain,
		IPAddress: ipAddress,
		Type:      recordType,
		TTL:       ttl,
		CreatedAt: time.Now(),
	}

	s.records[domain] = record
	s.cache[domain] = &cacheEntry{
		record:    record,
		expiresAt: time.Now().Add(time.Duration(ttl) * time.Second),
	}

	return record, nil
}

// Resolve resolves a domain to an IP address
func (s *DNSService) Resolve(domain string) (*DNSRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check cache first
	if entry, exists := s.cache[domain]; exists {
		if time.Now().Before(entry.expiresAt) {
			return entry.record, nil
		}
		// Cache expired
		delete(s.cache, domain)
	}

	// Check records
	if record, exists := s.records[domain]; exists {
		// Update cache
		s.cache[domain] = &cacheEntry{
			record:    record,
			expiresAt: time.Now().Add(time.Duration(record.TTL) * time.Second),
		}
		return record, nil
	}

	return nil, nil
}

// DeleteRecord deletes a DNS record
func (s *DNSService) DeleteRecord(domain string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.records, domain)
	delete(s.cache, domain)
	return nil
}

// ListRecords lists all DNS records
func (s *DNSService) ListRecords() []*DNSRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := make([]*DNSRecord, 0, len(s.records))
	for _, record := range s.records {
		records = append(records, record)
	}
	return records
}

var service *DNSService

func addRecordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Domain     string `json:"domain"`
		IPAddress  string `json:"ip_address"`
		Type       string `json:"type"`
		TTL        int    `json:"ttl"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	record, err := service.AddRecord(req.Domain, req.IPAddress, req.Type, req.TTL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

func resolveHandler(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "domain parameter is required", http.StatusBadRequest)
		return
	}

	record, err := service.Resolve(domain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if record == nil {
		http.Error(w, "domain not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

func deleteRecordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "domain parameter is required", http.StatusBadRequest)
		return
	}

	if err := service.DeleteRecord(domain); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func listRecordsHandler(w http.ResponseWriter, r *http.Request) {
	records := service.ListRecords()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func main() {
	service = NewDNSService()

	http.HandleFunc("/add", addRecordHandler)
	http.HandleFunc("/resolve", resolveHandler)
	http.HandleFunc("/delete", deleteRecordHandler)
	http.HandleFunc("/list", listRecordsHandler)
	http.HandleFunc("/health", healthHandler)

	port := ":8085"
	log.Printf("DNS service starting on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

