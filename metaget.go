// Package metadata provides functionality to extract meta tags from web pages.
// It is designed to be used in web crawlers, scrapers, and other tools that need to extract metadata from HTML pages.
package gohtmlmetadata

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

// MetaTag represents a single HTML meta tag with name and content attributes.
type MetaTag struct {
	Name    string
	Content string
}

// Extractor handles the retrieval and parsing of meta tags from web pages.
type Extractor struct {
	client *http.Client
}

// NewExtractor creates a new Extractor instance with a configurable transport.
// If transport is nil, http.DefaultTransport will be used.
func NewExtractor(transport http.RoundTripper) *Extractor {
	client := &http.Client{
		Transport: transport,
	}
	if transport == nil {
		client.Transport = http.DefaultTransport
	}
	return &Extractor{client: client}
}

// Extract fetches the page at the given URL and extracts all meta tags.
func (e *Extractor) Extract(url string) ([]MetaTag, error) {
	// Validate URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("invalid URL scheme: %s", url)
	}

	// Fetch the page
	resp, err := e.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the HTML
	return e.extractMetaTags(resp.Body)
}

// extractMetaTags parses HTML content and extracts meta tags.
func (e *Extractor) extractMetaTags(r io.Reader) ([]MetaTag, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var tags []MetaTag
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var name, content string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "name", "property":
					name = attr.Val
				case "content":
					content = attr.Val
				}
			}
			if name != "" && content != "" {
				tags = append(tags, MetaTag{
					Name:    name,
					Content: content,
				})
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return tags, nil
}
