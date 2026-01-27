// Package confluence provides a client for the Confluence REST API.
package confluence

import (
	"encoding/json"
	"fmt"
)

// Page represents a Confluence page returned by atl-cli.
type Page struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	SpaceKey string `json:"spaceKey"` // Actually spaceId from API
	Version  int    `json:"version"`
	Updated  string `json:"updated"`
	Body     string `json:"body"`
}

// apiPageResponse represents the Confluence API v2 response structure.
type apiPageResponse struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	SpaceID string `json:"spaceId"`
	Version struct {
		Number    int    `json:"number"`
		CreatedAt string `json:"createdAt"`
	} `json:"version"`
	Body struct {
		Storage *struct {
			Value string `json:"value"`
		} `json:"storage"`
	} `json:"body"`
}

// ParseAPIResponse parses a Confluence API response into a Page.
func ParseAPIResponse(data []byte) (*Page, error) {
	var resp apiPageResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse Confluence response: %w", err)
	}

	page := &Page{
		ID:       resp.ID,
		Title:    resp.Title,
		SpaceKey: resp.SpaceID, // Note: v2 API returns spaceId, not space key
		Version:  resp.Version.Number,
		Updated:  resp.Version.CreatedAt,
	}

	// Handle body content
	if resp.Body.Storage != nil {
		page.Body = resp.Body.Storage.Value
	} else {
		page.Body = ""
	}

	return page, nil
}

// Write writes the page as JSON to the given writer.
func (p *Page) Write(w interface{ Write([]byte) (int, error) }) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(p)
}

// WriteMarkdown writes the page as JSON with the body converted to Markdown.
func (p *Page) WriteMarkdown(w interface{ Write([]byte) (int, error) }) error {
	markdown, err := ToMarkdown(p.Body)
	if err != nil {
		return fmt.Errorf("failed to convert body to markdown: %w", err)
	}

	// Create a copy with converted body
	converted := &Page{
		ID:       p.ID,
		Title:    p.Title,
		SpaceKey: p.SpaceKey,
		Version:  p.Version,
		Updated:  p.Updated,
		Body:     markdown,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(converted)
}

// WriteBodyOnly writes just the Markdown body without JSON wrapper.
func (p *Page) WriteBodyOnly(w interface{ Write([]byte) (int, error) }) error {
	markdown, err := ToMarkdown(p.Body)
	if err != nil {
		return fmt.Errorf("failed to convert body to markdown: %w", err)
	}

	_, err = w.Write([]byte(markdown))
	if err != nil {
		return err
	}
	// Add a trailing newline for cleaner output
	_, err = w.Write([]byte("\n"))
	return err
}
