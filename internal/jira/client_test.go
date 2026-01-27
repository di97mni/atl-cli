package jira

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/martin/atl-cli/internal/config"
)

func TestClient_GetIssue_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request path
		if !strings.Contains(r.URL.Path, "/rest/api/3/issue/TEST-123") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify expand parameter
		if r.URL.Query().Get("expand") != "renderedFields" {
			t.Error("expected expand=renderedFields query parameter")
		}

		// Verify auth header
		if auth := r.Header.Get("Authorization"); !strings.HasPrefix(auth, "Basic ") {
			t.Error("expected Basic auth header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"key": "TEST-123",
			"fields": map[string]interface{}{
				"summary":  "Test issue",
				"status":   map[string]string{"name": "Open"},
				"assignee": map[string]string{"displayName": "Test User"},
				"priority": map[string]string{"name": "High"},
				"created":  "2026-01-15T10:30:00.000+0000",
				"updated":  "2026-01-25T14:45:00.000+0000",
			},
			"renderedFields": map[string]string{
				"description": "<p>Test description</p>",
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
	// Use test server's client to handle self-signed cert
	client.SetHTTPClient(server.Client())

	issue, err := client.GetIssue(context.Background(), "TEST-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if issue.Key != "TEST-123" {
		t.Errorf("expected key TEST-123, got %s", issue.Key)
	}
	if issue.Summary != "Test issue" {
		t.Errorf("expected summary 'Test issue', got '%s'", issue.Summary)
	}
	if issue.Description != "Test description" {
		t.Errorf("expected description 'Test description', got '%s'", issue.Description)
	}
}

func TestClient_GetIssue_NotFound(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Issue Does Not Exist"},
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

	_, err := client.GetIssue(context.Background(), "NOTFOUND-999")
	if err == nil {
		t.Error("expected error for non-existent issue")
	}

	// Should be a not_found error
	if !strings.Contains(err.Error(), "not_found") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not found error, got: %v", err)
	}
}

func TestClient_GetIssue_Unauthorized(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Unauthorized"},
		})
	}))
	defer server.Close()

	cfg := &config.Config{
		Site:  strings.TrimPrefix(server.URL, "https://"),
		Email: "test@example.com",
		Token: "bad-token",
	}

	client := NewClient(cfg, false)
	client.SetHTTPClient(server.Client())

	_, err := client.GetIssue(context.Background(), "TEST-123")
	if err == nil {
		t.Error("expected error for unauthorized request")
	}

	if !strings.Contains(err.Error(), "auth") && !strings.Contains(err.Error(), "401") {
		t.Errorf("expected auth error, got: %v", err)
	}
}

func TestClient_GetIssue_RateLimit(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	cfg := &config.Config{
		Site:  strings.TrimPrefix(server.URL, "https://"),
		Email: "test@example.com",
		Token: "test-token",
	}

	client := NewClient(cfg, false)
	client.SetHTTPClient(server.Client())

	_, err := client.GetIssue(context.Background(), "TEST-123")
	if err == nil {
		t.Error("expected error for rate limited request")
	}

	if !strings.Contains(err.Error(), "rate") && !strings.Contains(err.Error(), "429") {
		t.Errorf("expected rate limit error, got: %v", err)
	}
}

func TestClient_GetIssue_InvalidKey(t *testing.T) {
	cfg := &config.Config{
		Site:  "test.atlassian.net",
		Email: "test@example.com",
		Token: "test-token",
	}

	client := NewClient(cfg, false)

	_, err := client.GetIssue(context.Background(), "invalid")
	if err == nil {
		t.Error("expected error for invalid issue key")
	}

	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected validation error, got: %v", err)
	}
}
