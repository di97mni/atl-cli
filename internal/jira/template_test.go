package jira

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTemplate_Valid(t *testing.T) {
	content := `---
version: 1
issueType: story
project: CST
summary: "{{.title}}"
labels:
  - team-alpha
  - feature
---

## Overview

{{.description}}
`
	tmpl, err := ParseTemplate(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tmpl.Frontmatter.Version != 1 {
		t.Errorf("expected version 1, got %d", tmpl.Frontmatter.Version)
	}
	if tmpl.Frontmatter.IssueType != "story" {
		t.Errorf("expected issueType 'story', got %q", tmpl.Frontmatter.IssueType)
	}
	if tmpl.Frontmatter.Project != "CST" {
		t.Errorf("expected project 'CST', got %q", tmpl.Frontmatter.Project)
	}
	if tmpl.Frontmatter.Summary != "{{.title}}" {
		t.Errorf("expected summary '{{.title}}', got %q", tmpl.Frontmatter.Summary)
	}
	if len(tmpl.Frontmatter.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(tmpl.Frontmatter.Labels))
	}
}

func TestParseTemplate_NoFrontmatter(t *testing.T) {
	content := `Just a body without frontmatter`
	_, err := ParseTemplate(content)
	if err == nil {
		t.Error("expected error for missing frontmatter")
	}
}

func TestParseTemplate_UnclosedFrontmatter(t *testing.T) {
	content := `---
version: 1
issueType: story
`
	_, err := ParseTemplate(content)
	if err == nil {
		t.Error("expected error for unclosed frontmatter")
	}
}

func TestParseTemplate_InvalidVersion(t *testing.T) {
	content := `---
version: 2
issueType: story
project: CST
summary: "Test"
---

Body
`
	_, err := ParseTemplate(content)
	if err == nil {
		t.Error("expected error for unsupported version")
	}
}

func TestTemplate_Apply(t *testing.T) {
	content := `---
version: 1
issueType: story
project: CST
summary: "{{.title}}"
labels:
  - "{{.team}}"
---

## Overview

{{.description}}
`
	tmpl, err := ParseTemplate(content)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	vars := map[string]string{
		"title":       "New Feature",
		"team":        "team-beta",
		"description": "This is the feature description.",
	}

	parsed, err := tmpl.Apply(vars)
	if err != nil {
		t.Fatalf("failed to apply template: %v", err)
	}

	if parsed.Project != "CST" {
		t.Errorf("expected project 'CST', got %q", parsed.Project)
	}
	if parsed.IssueType != "story" {
		t.Errorf("expected issueType 'story', got %q", parsed.IssueType)
	}
	if parsed.Summary != "New Feature" {
		t.Errorf("expected summary 'New Feature', got %q", parsed.Summary)
	}
	if len(parsed.Labels) != 1 || parsed.Labels[0] != "team-beta" {
		t.Errorf("expected labels ['team-beta'], got %v", parsed.Labels)
	}
	if parsed.Description != "## Overview\n\nThis is the feature description." {
		t.Errorf("unexpected description: %q", parsed.Description)
	}
}

func TestTemplate_Apply_MissingVar(t *testing.T) {
	content := `---
version: 1
issueType: story
project: CST
summary: "{{.title}}"
---

Body
`
	tmpl, err := ParseTemplate(content)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	// Missing 'title' variable - Go templates render missing keys as <no value>
	vars := map[string]string{}

	parsed, err := tmpl.Apply(vars)
	if err != nil {
		t.Fatalf("failed to apply template: %v", err)
	}

	// Go templates replace missing keys with <no value>
	if parsed.Summary != "<no value>" {
		t.Errorf("expected '<no value>' for missing var, got %q", parsed.Summary)
	}
}

func TestTemplate_Apply_EmptyLabel(t *testing.T) {
	content := `---
version: 1
issueType: story
project: CST
summary: "Test"
labels:
  - ""
  - valid-label
---

Body
`
	tmpl, err := ParseTemplate(content)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	parsed, err := tmpl.Apply(map[string]string{})
	if err != nil {
		t.Fatalf("failed to apply template: %v", err)
	}

	// Empty labels should be filtered out
	if len(parsed.Labels) != 1 {
		t.Errorf("expected 1 label (empty filtered), got %d", len(parsed.Labels))
	}
	if parsed.Labels[0] != "valid-label" {
		t.Errorf("expected 'valid-label', got %q", parsed.Labels[0])
	}
}

func TestParseVarFlags_Valid(t *testing.T) {
	flags := []string{"title=Hello World", "team=alpha", "count=42"}
	vars, err := ParseVarFlags(flags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if vars["title"] != "Hello World" {
		t.Errorf("expected title 'Hello World', got %q", vars["title"])
	}
	if vars["team"] != "alpha" {
		t.Errorf("expected team 'alpha', got %q", vars["team"])
	}
	if vars["count"] != "42" {
		t.Errorf("expected count '42', got %q", vars["count"])
	}
}

func TestParseVarFlags_ValueWithEquals(t *testing.T) {
	// Value containing equals sign
	flags := []string{"equation=a=b+c"}
	vars, err := ParseVarFlags(flags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if vars["equation"] != "a=b+c" {
		t.Errorf("expected 'a=b+c', got %q", vars["equation"])
	}
}

func TestParseVarFlags_NoEquals(t *testing.T) {
	flags := []string{"invalid"}
	_, err := ParseVarFlags(flags)
	if err == nil {
		t.Error("expected error for missing equals")
	}
}

func TestParseVarFlags_EmptyKey(t *testing.T) {
	flags := []string{"=value"}
	_, err := ParseVarFlags(flags)
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestParseVarFlags_EmptyValue(t *testing.T) {
	flags := []string{"key="}
	vars, err := ParseVarFlags(flags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if vars["key"] != "" {
		t.Errorf("expected empty value, got %q", vars["key"])
	}
}

func TestLoadTemplate(t *testing.T) {
	// Create a temp file
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	content := `---
version: 1
issueType: task
project: TEST
summary: "Test task"
---

Task body
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	tmpl, err := LoadTemplate(path)
	if err != nil {
		t.Fatalf("failed to load template: %v", err)
	}

	if tmpl.Frontmatter.IssueType != "task" {
		t.Errorf("expected issueType 'task', got %q", tmpl.Frontmatter.IssueType)
	}
}

func TestLoadTemplate_NotFound(t *testing.T) {
	_, err := LoadTemplate("/nonexistent/path/template.md")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}
