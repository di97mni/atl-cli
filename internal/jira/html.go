package jira

import (
	"regexp"
	"strings"
)

var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

// StripHTML removes HTML tags and decodes common HTML entities.
func StripHTML(html string) string {
	// Remove all HTML tags
	text := htmlTagRegex.ReplaceAllString(html, "")

	// Decode common HTML entities
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&apos;", "'")
	text = strings.ReplaceAll(text, "&nbsp;", " ")

	// Normalize whitespace
	text = strings.Join(strings.Fields(text), " ")

	return strings.TrimSpace(text)
}
