package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// Page represents a crawled web page
type Page struct {
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Links       []string  `json:"links"`
	CrawledAt   time.Time `json:"crawled_at"`
	StatusCode  int       `json:"status_code"`
	ContentHash string    `json:"content_hash"`
}

// CrawlJob represents a crawl job
type CrawlJob struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Depth     int       `json:"depth"`
	Status    string    `json:"status"` // pending, running, completed, failed
	CreatedAt time.Time `json:"created_at"`
	Pages     int       `json:"pages"`
}

// WebCrawlerService manages web crawling
type WebCrawlerService struct {
	mu       sync.RWMutex
	pages    map[string]*Page      // URL -> Page
	jobs     map[string]*CrawlJob
	visited  map[string]bool
	jobIndex int64
}

// NewWebCrawlerService creates a new web crawler service
func NewWebCrawlerService() *WebCrawlerService {
	return &WebCrawlerService{
		pages:   make(map[string]*Page),
		jobs:    make(map[string]*CrawlJob),
		visited: make(map[string]bool),
	}
}

// CreateCrawlJob creates a new crawl job
func (s *WebCrawlerService) CreateCrawlJob(url string, depth int) (*CrawlJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobIndex++
	jobID := generateJobID(s.jobIndex)

	job := &CrawlJob{
		ID:        jobID,
		URL:       url,
		Depth:     depth,
		Status:    "pending",
		CreatedAt: time.Now(),
		Pages:     0,
	}

	s.jobs[jobID] = job

	// Start crawling in background
	go s.crawl(job)

	return job, nil
}

// crawl performs the actual crawling
func (s *WebCrawlerService) crawl(job *CrawlJob) {
	s.mu.Lock()
	job.Status = "running"
	s.mu.Unlock()

	// Simulate crawling
	urls := []string{job.URL}
	for i := 0; i < job.Depth && len(urls) > 0; i++ {
		currentURL := urls[0]
		urls = urls[1:]

		if s.isVisited(currentURL) {
			continue
		}

		page := s.crawlPage(currentURL)
		if page != nil {
			s.storePage(page)
			urls = append(urls, page.Links...)
			
			s.mu.Lock()
			job.Pages++
			s.mu.Unlock()
		}

		s.markVisited(currentURL)
	}

	s.mu.Lock()
	job.Status = "completed"
	s.mu.Unlock()
}

// crawlPage crawls a single page (simulated)
func (s *WebCrawlerService) crawlPage(url string) *Page {
	// Simulate HTTP request
	page := &Page{
		URL:        url,
		Title:      "Page Title for " + url,
		Content:    "Content for " + url,
		Links:      []string{url + "/link1", url + "/link2"},
		CrawledAt:  time.Now(),
		StatusCode: 200,
	}

	// Generate content hash
	hash := md5.Sum([]byte(page.Content))
	page.ContentHash = hex.EncodeToString(hash[:])

	return page
}

// storePage stores a crawled page
func (s *WebCrawlerService) storePage(page *Page) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pages[page.URL] = page
}

// isVisited checks if a URL has been visited
func (s *WebCrawlerService) isVisited(url string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.visited[url]
}

// markVisited marks a URL as visited
func (s *WebCrawlerService) markVisited(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.visited[url] = true
}

// GetJob retrieves a crawl job
func (s *WebCrawlerService) GetJob(jobID string) (*CrawlJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return nil, nil
	}

	return job, nil
}

// GetPage retrieves a crawled page
func (s *WebCrawlerService) GetPage(url string) (*Page, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	page, exists := s.pages[url]
	if !exists {
		return nil, nil
	}

	return page, nil
}

// ListPages lists all crawled pages
func (s *WebCrawlerService) ListPages() []*Page {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pages := make([]*Page, 0, len(s.pages))
	for _, page := range s.pages {
		pages = append(pages, page)
	}

	return pages
}

func generateJobID(index int64) string {
	return "job_" + string(rune(index+'0'))
}

var service *WebCrawlerService

func createJobHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL   string `json:"url"`
		Depth int    `json:"depth"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	job, err := service.CreateCrawlJob(req.URL, req.Depth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func getJobHandler(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("job_id")
	if jobID == "" {
		http.Error(w, "job_id parameter is required", http.StatusBadRequest)
		return
	}

	job, err := service.GetJob(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if job == nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func getPageHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "url parameter is required", http.StatusBadRequest)
		return
	}

	page, err := service.GetPage(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if page == nil {
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(page)
}

func listPagesHandler(w http.ResponseWriter, r *http.Request) {
	pages := service.ListPages()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pages)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func main() {
	service = NewWebCrawlerService()

	http.HandleFunc("/crawl", createJobHandler)
	http.HandleFunc("/job", getJobHandler)
	http.HandleFunc("/page", getPageHandler)
	http.HandleFunc("/pages", listPagesHandler)
	http.HandleFunc("/health", healthHandler)

	port := ":8086"
	log.Printf("Web crawler service starting on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

