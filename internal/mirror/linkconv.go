package mirror

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

// ConvertLinks walks all saved HTML files under rootDir and rewrites
// any links pointing to the original site into relative local paths.
// This makes the mirror browsable offline.
func ConvertLinks(rootURL *url.URL, rootDir string) error {
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process HTML files
		if info.IsDir() || !isHTMLFile(path) {
			return nil
		}

		if err := convertFile(path, rootURL, rootDir); err != nil {
			fmt.Printf("ConvertLinks: error processing %s: %v\n", path, err)
		}

		return nil
	})
}

// convertFile rewrites all links in a single HTML file.
func convertFile(filePath string, rootURL *url.URL, rootDir string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	converted, changed := rewriteLinks(content, filePath, rootURL, rootDir)
	if !changed {
		return nil
	}

	fmt.Printf("ConvertLinks: rewriting links in %s\n", filePath)
	return os.WriteFile(filePath, converted, 0o644)
}

// rewriteLinks parses the HTML, rewrites matching URLs to relative paths,
// and returns the updated HTML bytes plus whether anything changed.
func rewriteLinks(content []byte, filePath string, rootURL *url.URL, rootDir string) ([]byte, bool) {
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return content, false
	}

	changed := false

	// Walk every node in the HTML tree
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Find the URL attribute for this tag
			attrName, ok := linkTags[n.Data]
			if ok {
				for i, attr := range n.Attr {
					if attr.Key != attrName {
						continue
					}

					newVal := toRelativePath(attr.Val, filePath, rootURL, rootDir)
					if newVal != attr.Val {
						n.Attr[i].Val = newVal
						changed = true
					}
				}
			}
		}

		// Recurse into children
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(doc)

	if !changed {
		return content, false
	}

	// Render the modified tree back to bytes
	var buf bytes.Buffer
	html.Render(&buf, doc)
	return buf.Bytes(), true
}

// toRelativePath converts an absolute URL to a relative file path from the
// perspective of the current file on disk.
//
// Example:
//
//	currentFile: example.com/blog/post.html
//	targetURL:   https://example.com/images/logo.png
//	result:      ../images/logo.png
func toRelativePath(rawURL, currentFile string, rootURL *url.URL, rootDir string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Only rewrite URLs from the same domain
	if stripWWW(parsed.Host) != stripWWW(rootURL.Host) {
		return rawURL
	}

	// Convert the target URL to a local file path
	targetPath := urlToFilePath(rootURL, rawURL)
	targetPath = filepath.Join(rootDir, targetPath)

	// Compute relative path from the current file's directory to the target
	currentDir := filepath.Dir(currentFile)
	rel, err := filepath.Rel(currentDir, targetPath)
	if err != nil {
		return rawURL
	}

	// Use forward slashes for HTML compatibility on all platforms
	return filepath.ToSlash(rel)
}

// isHTMLFile returns true if the path has an .html or .htm extension.
func isHTMLFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".html" || ext == ".htm"
}