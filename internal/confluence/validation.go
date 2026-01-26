package confluence

import (
	"fmt"
	"regexp"
)

// pageIDPattern matches valid Confluence page IDs (numeric only).
var pageIDPattern = regexp.MustCompile(`^[0-9]+$`)

// ValidatePageID validates that a string is a valid Confluence page ID.
func ValidatePageID(id string) error {
	if id == "" {
		return fmt.Errorf("page ID cannot be empty")
	}

	if !pageIDPattern.MatchString(id) {
		return fmt.Errorf("invalid page ID: %q (must be numeric)", id)
	}

	return nil
}
