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

// ADFNode represents a node in an ADF document.
type ADFNode struct {
	Type    string    `json:"type"`
	Content []ADFNode `json:"content,omitempty"`
	Text    string    `json:"text,omitempty"`
}

// TextToADF converts plain text to an Atlassian Document Format document.
// It splits the text by double newlines into paragraphs.
func TextToADF(text string) *ADFDoc {
	if text == "" {
		return nil
	}

	doc := &ADFDoc{
		Type:    "doc",
		Version: 1,
		Content: []ADFNode{},
	}

	// Split into paragraphs by double newlines
	paragraphs := strings.Split(text, "\n\n")
	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// Replace single newlines with space within paragraphs
		para = strings.ReplaceAll(para, "\n", " ")

		doc.Content = append(doc.Content, ADFNode{
			Type: "paragraph",
			Content: []ADFNode{
				{
					Type: "text",
					Text: para,
				},
			},
		})
	}

	if len(doc.Content) == 0 {
		return nil
	}

	return doc
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
