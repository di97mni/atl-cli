package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/martin/atl-cli/internal/config"
)

func TestTextToADF_Empty(t *testing.T) {
	result := TextToADF("")
	if result != nil {
		t.Error("expected nil for empty text")
	}
}

func TestTextToADF_SingleParagraph(t *testing.T) {
	result := TextToADF("Hello world")
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.Type != "doc" {
		t.Errorf("expected type 'doc', got %q", result.Type)
	}
	if result.Version != 1 {
		t.Errorf("expected version 1, got %d", result.Version)
	}
	if len(result.Content) != 1 {
		t.Errorf("expected 1 paragraph, got %d", len(result.Content))
	}

	para := result.Content[0]
	if para.Type != "paragraph" {
		t.Errorf("expected paragraph type, got %q", para.Type)
	}
	if len(para.Content) != 1 {
		t.Errorf("expected 1 text node, got %d", len(para.Content))
	}
	if para.Content[0].Text != "Hello world" {
		t.Errorf("expected 'Hello world', got %q", para.Content[0].Text)
	}
}

func TestTextToADF_MultipleParagraphs(t *testing.T) {
	result := TextToADF("First paragraph\n\nSecond paragraph")
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.Content) != 2 {
		t.Errorf("expected 2 paragraphs, got %d", len(result.Content))
	}
	if result.Content[0].Content[0].Text != "First paragraph" {
		t.Errorf("expected 'First paragraph', got %q", result.Content[0].Content[0].Text)
	}
	if result.Content[1].Content[0].Text != "Second paragraph" {
		t.Errorf("expected 'Second paragraph', got %q", result.Content[1].Content[0].Text)
	}
}

func TestTextToADF_SingleNewlines(t *testing.T) {
	// Single newlines within a paragraph should become spaces
	result := TextToADF("Line one\nLine two")
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.Content) != 1 {
		t.Errorf("expected 1 paragraph, got %d", len(result.Content))
	}
	if result.Content[0].Content[0].Text != "Line one Line two" {
		t.Errorf("expected 'Line one Line two', got %q", result.Content[0].Content[0].Text)
	}
}

func TestNormalizeIssueType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"story", "Story"},
		{"Story", "Story"},
		{"STORY", "Story"},
		{"subtask", "Sub-task"},
		{"task", "Task"},
		{"bug", "Bug"},
		{"invalid", ""},
		{"", ""},
	}

	for _, tc := range tests {
		result := NormalizeIssueType(tc.input)
		if result != tc.expected {
			t.Errorf("NormalizeIssueType(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestValidIssueTypes(t *testing.T) {
	types := ValidIssueTypes()
	if len(types) != 4 {
		t.Errorf("expected 4 valid types, got %d", len(types))
	}

	expected := map[string]bool{"story": true, "subtask": true, "task": true, "bug": true}
	for _, typ := range types {
		if !expected[typ] {
			t.Errorf("unexpected type %q in ValidIssueTypes", typ)
		}
	}
}

func TestCreatedIssue_Write(t *testing.T) {
	issue := &CreatedIssue{
		Key: "TEST-123",
		URL: "https://test.atlassian.net/browse/TEST-123",
	}

	var buf bytes.Buffer
	if err := issue.Write(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if result["key"] != "TEST-123" {
		t.Errorf("expected key 'TEST-123', got %q", result["key"])
	}
	if result["url"] != "https://test.atlassian.net/browse/TEST-123" {
		t.Errorf("unexpected url: %q", result["url"])
	}
}

func TestClient_CreateIssue_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/rest/api/3/issue") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify request body
		var req CreateIssueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Fields.Project.Key != "TEST" {
			t.Errorf("expected project TEST, got %s", req.Fields.Project.Key)
		}
		if req.Fields.IssueType.Name != "Story" {
			t.Errorf("expected issue type Story, got %s", req.Fields.IssueType.Name)
		}
		if req.Fields.Summary != "Test story" {
			t.Errorf("expected summary 'Test story', got %s", req.Fields.Summary)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(CreateIssueResponse{
			ID:   "10001",
			Key:  "TEST-100",
			Self: "https://test.atlassian.net/rest/api/3/issue/10001",
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

	req := &CreateIssueRequest{
		Fields: CreateIssueFields{
			Project:   ProjectRef{Key: "TEST"},
			IssueType: IssueType{Name: "Story"},
			Summary:   "Test story",
		},
	}

	created, err := client.CreateIssue(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created.Key != "TEST-100" {
		t.Errorf("expected key TEST-100, got %s", created.Key)
	}
}

func TestClient_CreateIssue_WithDescription(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateIssueRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Fields.Description == nil {
			t.Error("expected description to be set")
		} else {
			if req.Fields.Description.Type != "doc" {
				t.Errorf("expected doc type, got %s", req.Fields.Description.Type)
			}
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(CreateIssueResponse{Key: "TEST-101"})
	}))
	defer server.Close()

	cfg := &config.Config{
		Site:  strings.TrimPrefix(server.URL, "https://"),
		Email: "test@example.com",
		Token: "test-token",
	}

	client := NewClient(cfg, false)
	client.SetHTTPClient(server.Client())

	req := &CreateIssueRequest{
		Fields: CreateIssueFields{
			Project:     ProjectRef{Key: "TEST"},
			IssueType:   IssueType{Name: "Story"},
			Summary:     "Test story",
			Description: TextToADF("This is the description"),
		},
	}

	_, err := client.CreateIssue(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_CreateIssue_SubTask(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateIssueRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Fields.Parent == nil {
			t.Error("expected parent to be set for subtask")
		} else if req.Fields.Parent.Key != "TEST-100" {
			t.Errorf("expected parent TEST-100, got %s", req.Fields.Parent.Key)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(CreateIssueResponse{Key: "TEST-102"})
	}))
	defer server.Close()

	cfg := &config.Config{
		Site:  strings.TrimPrefix(server.URL, "https://"),
		Email: "test@example.com",
		Token: "test-token",
	}

	client := NewClient(cfg, false)
	client.SetHTTPClient(server.Client())

	req := &CreateIssueRequest{
		Fields: CreateIssueFields{
			Project:   ProjectRef{Key: "TEST"},
			IssueType: IssueType{Name: "Sub-task"},
			Summary:   "Test subtask",
			Parent:    &ParentRef{Key: "TEST-100"},
		},
	}

	_, err := client.CreateIssue(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_CreateIssue_WithLabels(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateIssueRequest
		json.NewDecoder(r.Body).Decode(&req)

		if len(req.Fields.Labels) != 2 {
			t.Errorf("expected 2 labels, got %d", len(req.Fields.Labels))
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(CreateIssueResponse{Key: "TEST-103"})
	}))
	defer server.Close()

	cfg := &config.Config{
		Site:  strings.TrimPrefix(server.URL, "https://"),
		Email: "test@example.com",
		Token: "test-token",
	}

	client := NewClient(cfg, false)
	client.SetHTTPClient(server.Client())

	req := &CreateIssueRequest{
		Fields: CreateIssueFields{
			Project:   ProjectRef{Key: "TEST"},
			IssueType: IssueType{Name: "Story"},
			Summary:   "Test story",
			Labels:    []string{"label1", "label2"},
		},
	}

	_, err := client.CreateIssue(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_CreateIssue_InvalidProject(t *testing.T) {
	cfg := &config.Config{
		Site:  "test.atlassian.net",
		Email: "test@example.com",
		Token: "test-token",
	}

	client := NewClient(cfg, false)

	req := &CreateIssueRequest{
		Fields: CreateIssueFields{
			Project:   ProjectRef{Key: "invalid-project"},
			IssueType: IssueType{Name: "Story"},
			Summary:   "Test story",
		},
	}

	_, err := client.CreateIssue(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid project key")
	}
	if !strings.Contains(err.Error(), "invalid project key") {
		t.Errorf("expected validation error, got: %v", err)
	}
}

func TestClient_CreateIssue_InvalidParent(t *testing.T) {
	cfg := &config.Config{
		Site:  "test.atlassian.net",
		Email: "test@example.com",
		Token: "test-token",
	}

	client := NewClient(cfg, false)

	req := &CreateIssueRequest{
		Fields: CreateIssueFields{
			Project:   ProjectRef{Key: "TEST"},
			IssueType: IssueType{Name: "Sub-task"},
			Summary:   "Test subtask",
			Parent:    &ParentRef{Key: "invalid"},
		},
	}

	_, err := client.CreateIssue(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid parent key")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected validation error, got: %v", err)
	}
}

func TestClient_CreateIssue_APIError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Field 'summary' is required"},
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

	req := &CreateIssueRequest{
		Fields: CreateIssueFields{
			Project:   ProjectRef{Key: "TEST"},
			IssueType: IssueType{Name: "Story"},
			Summary:   "",
		},
	}

	_, err := client.CreateIssue(context.Background(), req)
	if err == nil {
		t.Error("expected error for API error response")
	}
}
