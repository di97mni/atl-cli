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
