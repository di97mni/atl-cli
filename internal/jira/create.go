package jira

import (
	"encoding/json"
	"strings"
)

// CreateIssueRequest represents the request body for creating a Jira issue.
type CreateIssueRequest struct {
	Fields CreateIssueFields `json:"fields"`
}

// CreateIssueFields contains the fields for creating an issue.
type CreateIssueFields struct {
	Project     ProjectRef  `json:"project"`
	IssueType   IssueType   `json:"issuetype"`
	Summary     string      `json:"summary"`
	Description *ADFDoc     `json:"description,omitempty"`
	Parent      *ParentRef  `json:"parent,omitempty"`
	Labels      []string    `json:"labels,omitempty"`
}

// ProjectRef is a reference to a project by key.
type ProjectRef struct {
	Key string `json:"key"`
}

// IssueType specifies the type of issue.
type IssueType struct {
	Name string `json:"name"`
}

// ParentRef is a reference to a parent issue (for sub-tasks).
type ParentRef struct {
	Key string `json:"key"`
}

// CreateIssueResponse represents the API response from creating an issue.
type CreateIssueResponse struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Self string `json:"self"`
}

// CreatedIssue is the CLI output format for a created issue.
type CreatedIssue struct {
	Key string `json:"key"`
	URL string `json:"url"`
}

// Write writes the created issue as JSON to the given writer.
func (c *CreatedIssue) Write(w interface{ Write([]byte) (int, error) }) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(c)
}

// ADFDoc represents an Atlassian Document Format document.
type ADFDoc struct {
	Type    string    `json:"type"`
	Version int       `json:"version"`
	Content []ADFNode `json:"content"`
}

// ADFMark represents an inline formatting mark on a text node.
type ADFMark struct {
	Type  string                 `json:"type"`
	Attrs map[string]interface{} `json:"attrs,omitempty"`
}

// ADFNode represents a node in an ADF document.
type ADFNode struct {
	Type    string                 `json:"type"`
	Attrs   map[string]interface{} `json:"attrs,omitempty"`
	Content []ADFNode              `json:"content,omitempty"`
	Text    string                 `json:"text,omitempty"`
	Marks   []ADFMark              `json:"marks,omitempty"`
}

// TextToADF converts markdown text to an Atlassian Document Format document.
// It supports headings, bold, italic, lists, code blocks, and links.
func TextToADF(text string) *ADFDoc {
	if text == "" {
		return nil
	}

	nodes := ParseMarkdownToADFNodes(text)
	if len(nodes) == 0 {
		return nil
	}

	return &ADFDoc{
		Type:    "doc",
		Version: 1,
		Content: nodes,
	}
}

// IssueTypeNameMap maps CLI type names to Jira issue type names.
var IssueTypeNameMap = map[string]string{
	"story":   "Story",
	"subtask": "Sub-task",
	"task":    "Task",
	"bug":     "Bug",
}

// ValidIssueTypes returns a list of valid issue type names for CLI.
func ValidIssueTypes() []string {
	return []string{"story", "subtask", "task", "bug"}
}

// NormalizeIssueType converts a CLI issue type to the Jira API name.
// Returns empty string if the type is invalid.
func NormalizeIssueType(issueType string) string {
	return IssueTypeNameMap[strings.ToLower(issueType)]
}
