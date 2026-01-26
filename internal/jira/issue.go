// Package jira provides a client for the Jira REST API.
package jira

import (
	"encoding/json"
	"fmt"
)

// Issue represents a Jira issue returned by atl-cli.
type Issue struct {
	Key         string  `json:"key"`
	Summary     string  `json:"summary"`
	Status      string  `json:"status"`
	Assignee    *string `json:"assignee"`    // null if unassigned
	Priority    *string `json:"priority"`    // null if not set
	Created     string  `json:"created"`
	Updated     string  `json:"updated"`
	Description string  `json:"description"`
	URL         string  `json:"url"`
}

// apiIssueResponse represents the Jira API response structure.
type apiIssueResponse struct {
	Key    string `json:"key"`
	Fields struct {
		Summary  string `json:"summary"`
		Status   struct {
			Name string `json:"name"`
		} `json:"status"`
		Assignee *struct {
			DisplayName string `json:"displayName"`
		} `json:"assignee"`
		Priority *struct {
			Name string `json:"name"`
		} `json:"priority"`
		Created string `json:"created"`
		Updated string `json:"updated"`
	} `json:"fields"`
	RenderedFields struct {
		Description *string `json:"description"`
	} `json:"renderedFields"`
}

// ParseAPIResponse parses a Jira API response into an Issue.
func ParseAPIResponse(data []byte, site string) (*Issue, error) {
	var resp apiIssueResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse Jira response: %w", err)
	}

	issue := &Issue{
		Key:     resp.Key,
		Summary: resp.Fields.Summary,
		Status:  resp.Fields.Status.Name,
		Created: resp.Fields.Created,
		Updated: resp.Fields.Updated,
		URL:     fmt.Sprintf("https://%s/browse/%s", site, resp.Key),
	}

	// Handle nullable fields
	if resp.Fields.Assignee != nil {
		issue.Assignee = &resp.Fields.Assignee.DisplayName
	}
	if resp.Fields.Priority != nil {
		issue.Priority = &resp.Fields.Priority.Name
	}

	// Handle description - strip HTML from renderedFields
	if resp.RenderedFields.Description != nil {
		issue.Description = StripHTML(*resp.RenderedFields.Description)
	} else {
		issue.Description = ""
	}

	return issue, nil
}

// Write writes the issue as JSON to the given writer.
func (i *Issue) Write(w interface{ Write([]byte) (int, error) }) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(i)
}
