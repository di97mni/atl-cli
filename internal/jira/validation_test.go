package jira

import "testing"

func TestValidateIssueKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		// Valid keys
		{"simple key", "PROJ-123", false},
		{"single letter project", "A-1", false},
		{"long project key", "VERYLONGPROJECT-99999", false},
		{"project with numbers", "TEST123-456", false},
		{"two letter project", "AB-1", false},

		// Invalid keys
		{"empty string", "", true},
		{"missing dash", "PROJ123", true},
		{"lowercase project", "proj-123", true},
		{"missing issue number", "PROJ-", true},
		{"missing project", "-123", true},
		{"just number", "123", true},
		{"number first in project", "1PROJ-123", true},
		{"space in key", "PROJ 123", true},
		{"special characters", "PROJ-123!", true},
		{"double dash", "PROJ--123", true},
		{"zero issue number", "PROJ-0", false}, // 0 is technically valid per regex
		{"negative number", "PROJ--1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIssueKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIssueKey(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
			}
		})
	}
}
