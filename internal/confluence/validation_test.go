package confluence

import "testing"

func TestValidatePageID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		// Valid IDs
		{"simple numeric", "123456789", false},
		{"single digit", "1", false},
		{"large number", "9999999999999999", false},
		{"zero", "0", false},

		// Invalid IDs
		{"empty string", "", true},
		{"alphabetic", "abc", true},
		{"alphanumeric", "123abc", true},
		{"negative", "-123", true},
		{"decimal", "123.456", true},
		{"with spaces", "123 456", true},
		{"with dash", "123-456", true},
		{"special chars", "123!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePageID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePageID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}
