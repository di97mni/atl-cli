package jira

import (
	"encoding/json"
	"testing"
)

func TestIssue_JSON(t *testing.T) {
	issue := &Issue{
		Key:         "TEST-123",
		Summary:     "Test issue",
		Status:      "In Progress",
		Assignee:    stringPtr("John Doe"),
		Priority:    stringPtr("High"),
		Created:     "2026-01-15T10:30:00.000Z",
		Updated:     "2026-01-25T14:45:00.000Z",
		Description: "This is the issue description",
		URL:         "https://test.atlassian.net/browse/TEST-123",
	}

	data, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("failed to marshal issue: %v", err)
	}

	var parsed Issue
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal issue: %v", err)
	}

	if parsed.Key != issue.Key {
		t.Errorf("expected key %s, got %s", issue.Key, parsed.Key)
	}
	if parsed.Summary != issue.Summary {
		t.Errorf("expected summary %s, got %s", issue.Summary, parsed.Summary)
	}
	if parsed.Status != issue.Status {
		t.Errorf("expected status %s, got %s", issue.Status, parsed.Status)
	}
	if parsed.Assignee == nil || *parsed.Assignee != *issue.Assignee {
		t.Errorf("expected assignee %v, got %v", issue.Assignee, parsed.Assignee)
	}
	if parsed.Priority == nil || *parsed.Priority != *issue.Priority {
		t.Errorf("expected priority %v, got %v", issue.Priority, parsed.Priority)
	}
}

func TestIssue_NullableFields(t *testing.T) {
	// Test with null assignee and priority
	issue := &Issue{
		Key:         "TEST-456",
		Summary:     "Unassigned issue",
		Status:      "Open",
		Assignee:    nil,
		Priority:    nil,
		Created:     "2026-01-15T10:30:00.000Z",
		Updated:     "2026-01-25T14:45:00.000Z",
		Description: "",
		URL:         "https://test.atlassian.net/browse/TEST-456",
	}

	data, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("failed to marshal issue: %v", err)
	}

	// Verify null fields appear as null in JSON
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	if raw["assignee"] != nil {
		t.Error("expected assignee to be null")
	}
	if raw["priority"] != nil {
		t.Error("expected priority to be null")
	}
}

func TestParseAPIResponse(t *testing.T) {
	apiResponse := `{
		"key": "PROJ-123",
		"fields": {
			"summary": "Test issue summary",
			"status": {"name": "In Progress"},
			"assignee": {"displayName": "Jane Doe"},
			"priority": {"name": "Medium"},
			"created": "2026-01-15T10:30:00.000+0000",
			"updated": "2026-01-25T14:45:00.000+0000"
		},
		"renderedFields": {
			"description": "<p>Issue description with <strong>bold</strong> text</p>"
		}
	}`

	issue, err := ParseAPIResponse([]byte(apiResponse), "test.atlassian.net")
	if err != nil {
		t.Fatalf("failed to parse API response: %v", err)
	}

	if issue.Key != "PROJ-123" {
		t.Errorf("expected key PROJ-123, got %s", issue.Key)
	}
	if issue.Summary != "Test issue summary" {
		t.Errorf("expected summary 'Test issue summary', got '%s'", issue.Summary)
	}
	if issue.Status != "In Progress" {
		t.Errorf("expected status 'In Progress', got '%s'", issue.Status)
	}
	if issue.Assignee == nil || *issue.Assignee != "Jane Doe" {
		t.Errorf("expected assignee 'Jane Doe', got %v", issue.Assignee)
	}
	if issue.Priority == nil || *issue.Priority != "Medium" {
		t.Errorf("expected priority 'Medium', got %v", issue.Priority)
	}
	if issue.URL != "https://test.atlassian.net/browse/PROJ-123" {
		t.Errorf("expected URL 'https://test.atlassian.net/browse/PROJ-123', got '%s'", issue.URL)
	}
	// Description should have HTML stripped
	if issue.Description != "Issue description with bold text" {
		t.Errorf("expected stripped description, got '%s'", issue.Description)
	}
}

func TestParseAPIResponse_NullFields(t *testing.T) {
	apiResponse := `{
		"key": "PROJ-456",
		"fields": {
			"summary": "Unassigned issue",
			"status": {"name": "Open"},
			"assignee": null,
			"priority": null,
			"created": "2026-01-15T10:30:00.000+0000",
			"updated": "2026-01-25T14:45:00.000+0000"
		},
		"renderedFields": {
			"description": null
		}
	}`

	issue, err := ParseAPIResponse([]byte(apiResponse), "test.atlassian.net")
	if err != nil {
		t.Fatalf("failed to parse API response: %v", err)
	}

	if issue.Assignee != nil {
		t.Errorf("expected assignee to be nil, got %v", issue.Assignee)
	}
	if issue.Priority != nil {
		t.Errorf("expected priority to be nil, got %v", issue.Priority)
	}
	if issue.Description != "" {
		t.Errorf("expected empty description, got '%s'", issue.Description)
	}
}

func stringPtr(s string) *string {
	return &s
}
