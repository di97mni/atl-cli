package confluence

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPage_JSON(t *testing.T) {
	page := &Page{
		ID:       "123456789",
		Title:    "Test Page",
		SpaceKey: "65011",
		Version:  5,
		Updated:  "2026-01-20T15:45:00.000Z",
		Body:     "<p>Page content</p>",
	}

	data, err := json.Marshal(page)
	if err != nil {
		t.Fatalf("failed to marshal page: %v", err)
	}

	var parsed Page
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal page: %v", err)
	}

	if parsed.ID != page.ID {
		t.Errorf("expected ID %s, got %s", page.ID, parsed.ID)
	}
	if parsed.Title != page.Title {
		t.Errorf("expected title %s, got %s", page.Title, parsed.Title)
	}
	if parsed.SpaceKey != page.SpaceKey {
		t.Errorf("expected spaceKey %s, got %s", page.SpaceKey, parsed.SpaceKey)
	}
	if parsed.Version != page.Version {
		t.Errorf("expected version %d, got %d", page.Version, parsed.Version)
	}
	if parsed.Body != page.Body {
		t.Errorf("expected body %s, got %s", page.Body, parsed.Body)
	}
}

func TestParseAPIResponse(t *testing.T) {
	apiResponse := `{
		"id": "123456789",
		"title": "Test Page Title",
		"spaceId": "65011",
		"version": {
			"number": 5,
			"createdAt": "2026-01-20T15:45:00.000Z"
		},
		"body": {
			"storage": {
				"value": "<p>This is the page content</p>",
				"representation": "storage"
			}
		}
	}`

	page, err := ParseAPIResponse([]byte(apiResponse))
	if err != nil {
		t.Fatalf("failed to parse API response: %v", err)
	}

	if page.ID != "123456789" {
		t.Errorf("expected ID '123456789', got '%s'", page.ID)
	}
	if page.Title != "Test Page Title" {
		t.Errorf("expected title 'Test Page Title', got '%s'", page.Title)
	}
	if page.SpaceKey != "65011" {
		t.Errorf("expected spaceKey '65011', got '%s'", page.SpaceKey)
	}
	if page.Version != 5 {
		t.Errorf("expected version 5, got %d", page.Version)
	}
	if page.Updated != "2026-01-20T15:45:00.000Z" {
		t.Errorf("expected updated '2026-01-20T15:45:00.000Z', got '%s'", page.Updated)
	}
	if page.Body != "<p>This is the page content</p>" {
		t.Errorf("expected body '<p>This is the page content</p>', got '%s'", page.Body)
	}
}

func TestParseAPIResponse_EmptyBody(t *testing.T) {
	apiResponse := `{
		"id": "987654321",
		"title": "Empty Page",
		"spaceId": "12345",
		"version": {
			"number": 1,
			"createdAt": "2026-01-15T10:00:00.000Z"
		},
		"body": {}
	}`

	page, err := ParseAPIResponse([]byte(apiResponse))
	if err != nil {
		t.Fatalf("failed to parse API response: %v", err)
	}

	if page.Body != "" {
		t.Errorf("expected empty body, got '%s'", page.Body)
	}
}

func TestPage_WriteMarkdown(t *testing.T) {
	page := &Page{
		ID:       "123456789",
		Title:    "Test Page",
		SpaceKey: "65011",
		Version:  5,
		Updated:  "2026-01-20T15:45:00.000Z",
		Body:     "<h1>Hello World</h1><p>This is a <strong>test</strong>.</p>",
	}

	var buf bytes.Buffer
	if err := page.WriteMarkdown(&buf); err != nil {
		t.Fatalf("WriteMarkdown() error = %v", err)
	}

	var result Page
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Check that metadata is preserved
	if result.ID != page.ID {
		t.Errorf("expected ID %s, got %s", page.ID, result.ID)
	}
	if result.Title != page.Title {
		t.Errorf("expected title %s, got %s", page.Title, result.Title)
	}

	// Check that body is converted to markdown
	if !strings.Contains(result.Body, "# Hello World") {
		t.Errorf("expected markdown heading, got %s", result.Body)
	}
	if !strings.Contains(result.Body, "**test**") {
		t.Errorf("expected markdown bold, got %s", result.Body)
	}
}

func TestPage_WriteBodyOnly(t *testing.T) {
	page := &Page{
		ID:       "123456789",
		Title:    "Test Page",
		SpaceKey: "65011",
		Version:  5,
		Updated:  "2026-01-20T15:45:00.000Z",
		Body:     "<h1>Hello World</h1><p>This is a <strong>test</strong>.</p>",
	}

	var buf bytes.Buffer
	if err := page.WriteBodyOnly(&buf); err != nil {
		t.Fatalf("WriteBodyOnly() error = %v", err)
	}

	result := buf.String()

	// Should NOT be JSON
	if strings.HasPrefix(result, "{") {
		t.Errorf("expected plain text, got JSON")
	}

	// Should contain markdown
	if !strings.Contains(result, "# Hello World") {
		t.Errorf("expected markdown heading, got %s", result)
	}
	if !strings.Contains(result, "**test**") {
		t.Errorf("expected markdown bold, got %s", result)
	}
}
