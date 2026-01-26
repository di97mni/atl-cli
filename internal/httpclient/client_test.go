package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_Do_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Basic ") {
			t.Error("expected Basic auth header")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := New("test@example.com", "test-token", false)
	req, _ := http.NewRequest("GET", server.URL+"/test", nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestClient_Do_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with short timeout for testing
	client := &Client{
		http: &http.Client{Timeout: 100 * time.Millisecond},
	}
	req, _ := http.NewRequest("GET", server.URL+"/slow", nil)

	_, err := client.Do(req)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestClient_Do_AuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Unauthorized"},
		})
	}))
	defer server.Close()

	client := New("test@example.com", "bad-token", false)
	req, _ := http.NewRequest("GET", server.URL+"/test", nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestClient_Do_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Issue Does Not Exist"},
		})
	}))
	defer server.Close()

	client := New("test@example.com", "test-token", false)
	req, _ := http.NewRequest("GET", server.URL+"/issue/NOTFOUND-1", nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestClient_Do_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := New("test@example.com", "test-token", false)
	req, _ := http.NewRequest("GET", server.URL+"/test", nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", resp.StatusCode)
	}

	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter != "60" {
		t.Errorf("expected Retry-After: 60, got %s", retryAfter)
	}
}

func TestClient_NewRequest_AddsAuth(t *testing.T) {
	client := New("test@example.com", "test-token", false)
	req, err := client.NewRequest(context.Background(), "GET", "https://example.com/api", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	auth := req.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Basic ") {
		t.Error("expected Basic auth header")
	}

	accept := req.Header.Get("Accept")
	if accept != "application/json" {
		t.Errorf("expected Accept: application/json, got %s", accept)
	}
}

func TestClient_HTTPSEnforcement(t *testing.T) {
	client := New("test@example.com", "test-token", false)

	// HTTP should be rejected
	_, err := client.NewRequest(context.Background(), "GET", "http://example.com/api", nil)
	if err == nil {
		t.Error("expected error for HTTP URL")
	}

	// HTTPS should work
	req, err := client.NewRequest(context.Background(), "GET", "https://example.com/api", nil)
	if err != nil {
		t.Fatalf("unexpected error for HTTPS URL: %v", err)
	}
	if req == nil {
		t.Error("expected non-nil request")
	}
}
