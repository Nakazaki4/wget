package mirror

import (
	"io"
	"net/url"

	"golang.org/x/net/html"
)

var linkTags = map[string]string{
	"a":      "href",
	"img":    "src",
	"link":   "href",
	"script": "src",
}

func ParseLinks(body io.Reader, baseURL *url.URL) []string {
	var links []string
	seen := make(map[string]bool)

	tokenizer := html.NewTokenizer(body)

	for {
		tt := tokenizer.Next()

		switch tt {
		case html.ErrorToken:
			return links

		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()

			attr, ok := linkTags[token.Data]
			if !ok {
				continue
			}

			rawURL := extractAttr(token, attr)
			if rawURL == "" {
				continue
			}

			absURL := resolveURL(baseURL, rawURL)
			if absURL == "" {
				continue
			}

			if !seen[absURL] {
				seen[absURL] = true
				links = append(links, absURL)
			}
		}
	}
}

func extractAttr(token html.Token, attrName string) string {
	for _, attr := range token.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}
	return ""
}

func resolveURL(base *url.URL, raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	if parsed.Scheme != "" && parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}

	resolved := base.ResolveReference(parsed)

	resolved.Fragment = ""

	return resolved.String()
}
