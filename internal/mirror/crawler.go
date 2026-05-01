package mirror

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const numWorkers = 8

type Crawler struct {
	rootURL  *url.URL
	filter   *Filter
	reject   string
	exclude  string
	visited  map[string]bool
	mu       sync.Mutex
	wg       sync.WaitGroup
	queue    chan string
}

// NewCrawler creates a Crawler from the root URL and config flags.
func NewCrawler(rootURL, reject, exclude string) (*Crawler, error) {
	parsed, err := url.Parse(rootURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	return &Crawler{
		rootURL: parsed,
		filter:  NewFilter(parsed, reject, exclude),
		visited: make(map[string]bool),
		queue:   make(chan string, 512),
	}, nil
}

// Start begins the crawl from the root URL and blocks until complete.
func (c *Crawler) Start() {
	fmt.Printf("Mirror: starting crawl of %s\n", c.rootURL.String())

	for i := range numWorkers {
		go c.worker(i)
	}

	c.enqueue(c.rootURL.String())

	// Wait for all work to finish
	c.wg.Wait()
	close(c.queue)

	fmt.Printf("Mirror: finished crawling %s\n", c.rootURL.String())
}

// worker pulls URLs from the queue and processes them.
func (c *Crawler) worker(id int) {
	for rawURL := range c.queue {
		c.process(rawURL)
		c.wg.Done()
	}
}

// enqueue adds a URL to the queue if it hasn't been visited yet.
func (c *Crawler) enqueue(rawURL string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.visited[rawURL] {
		return
	}
	c.visited[rawURL] = true

	c.wg.Add(1)
	c.queue <- rawURL
}

// process downloads a single URL, saves it to disk, then enqueues any links found.
func (c *Crawler) process(rawURL string) {
	fmt.Printf("Mirror: fetching %s\n", rawURL)

	resp, err := http.Get(rawURL)
	if err != nil {
		fmt.Printf("Mirror: error fetching %s: %v\n", rawURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Mirror: skipping %s (status %d)\n", rawURL, resp.StatusCode)
		return
	}

	// Read the body once — we may need it twice (save + parse links)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Mirror: error reading body of %s: %v\n", rawURL, err)
		return
	}

	// Save to disk
	savePath := urlToFilePath(c.rootURL, rawURL)
	if err := saveFile(savePath, body); err != nil {
		fmt.Printf("Mirror: error saving %s: %v\n", savePath, err)
		return
	}

	// Only parse links from HTML pages
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return
	}

	// Parse links and enqueue allowed ones
	pageURL, _ := url.Parse(rawURL)
	links := ParseLinks(bytes.NewReader(body), pageURL)

	for _, link := range links {
		if c.filter.Allow(link) {
			c.enqueue(link)
		}
	}
}

// urlToFilePath converts a URL to a local file path mirroring the site structure.
//
//	https://example.com/blog/post.html  →  example.com/blog/post.html
//	https://example.com/               →  example.com/index.html
//	https://example.com/about          →  example.com/about/index.html
func urlToFilePath(rootURL *url.URL, rawURL string) string {
	parsed, _ := url.Parse(rawURL)

	// Base directory is the hostname
	base := parsed.Host

	urlPath := parsed.Path

	// Root or directory-like path → save as index.html
	if urlPath == "" || urlPath == "/" {
		return filepath.Join(base, "index.html")
	}

	// If the path has no extension, treat it as a directory and add index.html
	if filepath.Ext(urlPath) == "" {
		urlPath = strings.TrimSuffix(urlPath, "/")
		urlPath = urlPath + "/index.html"
	}

	// filepath.Join cleans up any double slashes etc.
	return filepath.Join(base, filepath.FromSlash(urlPath))
}

// saveFile writes content to a path, creating any missing directories.
func saveFile(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	fmt.Printf("Mirror: saved %s\n", path)
	return nil
}