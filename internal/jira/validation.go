package jira

import (
	"fmt"
	"regexp"
)

// issueKeyPattern matches valid Jira issue keys like "PROJ-123".
// Pattern: Project key (1+ uppercase letters, optionally followed by numbers) + dash + issue number
var issueKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]*-[0-9]+$`)

// ValidateIssueKey validates that a string is a valid Jira issue key.
func ValidateIssueKey(key string) error {
	if key == "" {
		return fmt.Errorf("issue key cannot be empty")
	}

	if !issueKeyPattern.MatchString(key) {
		return fmt.Errorf("invalid issue key format: %q (expected format: PROJ-123)", key)
	}

	return nil
}
