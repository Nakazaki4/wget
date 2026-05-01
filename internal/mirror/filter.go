package mirror

import (
	"net/url"
	"path/filepath"
	"strings"
)

// Filter decides whether a URL should be crawled/downloaded.
// It checks three things in order:
//  1. Is it on the same domain as the root?
//  2. Is its file extension in the reject list?
//  3. Is its path under an excluded directory?
type Filter struct {
	rootHost string   // e.g. "example.com"
	reject   []string // file extensions to skip, e.g. ["jpg", "png"]
	exclude  []string // path prefixes to skip, e.g. ["/admin", "/private"]
}

// NewFilter creates a Filter from the root URL and raw flag strings.
// reject and exclude are comma-separated strings from --reject and --exclude flags.
func NewFilter(rootURL *url.URL, reject, exclude string) *Filter {
	return &Filter{
		rootHost: rootURL.Host,
		reject:   splitAndClean(reject),
		exclude:  splitAndClean(exclude),
	}
}

// Allow returns true if the URL should be downloaded, false if it should be skipped.
func (f *Filter) Allow(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	if !f.sameDomain(parsed) {
		return false
	}

	if f.isRejected(parsed) {
		return false
	}

	if f.isExcluded(parsed) {
		return false
	}

	return true
}

// sameDomain checks the URL host matches the root host.
// It strips www. from both sides so that example.com and www.example.com are treated the same.
func (f *Filter) sameDomain(u *url.URL) bool {
	return stripWWW(u.Host) == stripWWW(f.rootHost)
}

// isRejected checks whether the URL's file extension is in the reject list.
// e.g. --reject jpg,png will skip any URL ending in .jpg or .png
func (f *Filter) isRejected(u *url.URL) bool {
	if len(f.reject) == 0 {
		return false
	}

	ext := filepath.Ext(u.Path) // e.g. ".jpg"
	if ext == "" {
		return false
	}

	ext = strings.TrimPrefix(ext, ".") // strip the dot -> "jpg"
	ext = strings.ToLower(ext)

	for _, r := range f.reject {
		if ext == r {
			return true
		}
	}
	return false
}

// isExcluded checks whether the URL's path starts with any excluded prefix.
// e.g. --exclude /admin will skip /admin/login, /admin/dashboard, etc.
func (f *Filter) isExcluded(u *url.URL) bool {
	if len(f.exclude) == 0 {
		return false
	}

	for _, ex := range f.exclude {
		// Ensure the prefix ends with / so /adminpanel doesn't match /admin
		prefix := ex
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}

		path := u.Path
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}

		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// splitAndClean splits a comma-separated string and lowercases/trims each part.
// Returns nil if the input is empty.
func splitAndClean(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// stripWWW removes a leading "www." from a hostname.
func stripWWW(host string) string {
	return strings.TrimPrefix(host, "www.")
}