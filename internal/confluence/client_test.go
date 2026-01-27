package confluence

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/martin/atl-cli/internal/config"
)

func TestClient_GetPage_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request path
		if !strings.Contains(r.URL.Path, "/wiki/api/v2/pages/123456789") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify body-format parameter
		if r.URL.Query().Get("body-format") != "storage" {
			t.Error("expected body-format=storage query parameter")
		}

		// Verify auth header
		if auth := r.Header.Get("Authorization"); !strings.HasPrefix(auth, "Basic ") {
			t.Error("expected Basic auth header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "123456789",
			"title":   "Test Page",
			"spaceId": "65011",
			"version": map[string]interface{}{
				"number":    5,
				"createdAt": "2026-01-20T15:45:00.000Z",
			},
			"body": map[string]interface{}{
				"storage": map[string]string{
					"value":          "<p>Test content</p>",
					"representation": "storage",
				},
			},
		})
	}))
	defer server.Close()

	cfg := &config.Config{
		Site:  strings.TrimPrefix(server.URL, "https://"),
		Email: "test@example.com",
		Token: "test-token",
	}

	client := NewClient(cfg, false)
	client.SetHTTPClient(server.Client())

	page, err := client.GetPage(context.Background(), "123456789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if page.ID != "123456789" {
		t.Errorf("expected ID 123456789, got %s", page.ID)
	}
	if page.Title != "Test Page" {
		t.Errorf("expected title 'Test Page', got '%s'", page.Title)
	}
	if page.Body != "<p>Test content</p>" {
		t.Errorf("expected body '<p>Test content</p>', got '%s'", page.Body)
	}
}

func TestClient_GetPage_NotFound(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]interface{}{
				{"status": 404, "code": "PAGE_NOT_FOUND", "title": "Page not found"},
			},
		})
	}))
	defer server.Close()

	cfg := &config.Config{
		Site:  strings.TrimPrefix(server.URL, "https://"),
		Email: "test@example.com",
		Token: "test-token",
	}

	client := NewClient(cfg, false)
	client.SetHTTPClient(server.Client())

	_, err := client.GetPage(context.Background(), "999999999")
	if err == nil {
		t.Error("expected error for non-existent page")
	}

	if !strings.Contains(err.Error(), "not_found") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not found error, got: %v", err)
	}
}

func TestClient_GetPage_Unauthorized(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	cfg := &config.Config{
		Site:  strings.TrimPrefix(server.URL, "https://"),
		Email: "test@example.com",
		Token: "bad-token",
	}

	client := NewClient(cfg, false)
	client.SetHTTPClient(server.Client())

	_, err := client.GetPage(context.Background(), "123456789")
	if err == nil {
		t.Error("expected error for unauthorized request")
	}

	if !strings.Contains(err.Error(), "auth") && !strings.Contains(err.Error(), "401") {
		t.Errorf("expected auth error, got: %v", err)
	}
}

func TestClient_GetPage_InvalidID(t *testing.T) {
	cfg := &config.Config{
		Site:  "test.atlassian.net",
		Email: "test@example.com",
		Token: "test-token",
	}

	client := NewClient(cfg, false)

	_, err := client.GetPage(context.Background(), "invalid")
	if err == nil {
		t.Error("expected error for invalid page ID")
	}

	if !strings.Contains(err.Error(), "numeric") {
		t.Errorf("expected validation error, got: %v", err)
	}
}
