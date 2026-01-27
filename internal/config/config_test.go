package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoadFromEnv(t *testing.T) {
	// Save original env
	origSite := os.Getenv(EnvSite)
	origEmail := os.Getenv(EnvEmail)
	origToken := os.Getenv(EnvToken)
	defer func() {
		os.Setenv(EnvSite, origSite)
		os.Setenv(EnvEmail, origEmail)
		os.Setenv(EnvToken, origToken)
	}()

	// Set test values
	os.Setenv(EnvSite, "test.atlassian.net")
	os.Setenv(EnvEmail, "test@example.com")
	os.Setenv(EnvToken, "test-token-123")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Site != "test.atlassian.net" {
		t.Errorf("expected site 'test.atlassian.net', got '%s'", cfg.Site)
	}
	if cfg.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", cfg.Email)
	}
	if cfg.Token != "test-token-123" {
		t.Errorf("expected token 'test-token-123', got '%s'", cfg.Token)
	}
}

func TestConfig_Validate_AllPresent(t *testing.T) {
	cfg := &Config{
		Site:  "test.atlassian.net",
		Email: "test@example.com",
		Token: "test-token",
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestConfig_Validate_MissingSite(t *testing.T) {
	cfg := &Config{
		Site:  "",
		Email: "test@example.com",
		Token: "test-token",
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing site")
	}
	if !strings.Contains(err.Error(), EnvSite) {
		t.Errorf("expected error to mention %s, got: %v", EnvSite, err)
	}
}

func TestConfig_Validate_MissingEmail(t *testing.T) {
	cfg := &Config{
		Site:  "test.atlassian.net",
		Email: "",
		Token: "test-token",
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing email")
	}
	if !strings.Contains(err.Error(), EnvEmail) {
		t.Errorf("expected error to mention %s, got: %v", EnvEmail, err)
	}
}

func TestConfig_Validate_MissingToken(t *testing.T) {
	cfg := &Config{
		Site:  "test.atlassian.net",
		Email: "test@example.com",
		Token: "",
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing token")
	}
	if !strings.Contains(err.Error(), EnvToken) {
		t.Errorf("expected error to mention %s, got: %v", EnvToken, err)
	}
}

func TestConfig_Validate_AllMissing(t *testing.T) {
	cfg := &Config{
		Site:  "",
		Email: "",
		Token: "",
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for all missing fields")
	}
	// Should report first missing field (Site)
	if !strings.Contains(err.Error(), EnvSite) {
		t.Errorf("expected error to mention %s first, got: %v", EnvSite, err)
	}
}

func TestConfig_BaseURL(t *testing.T) {
	cfg := &Config{
		Site: "test.atlassian.net",
	}

	url := cfg.BaseURL()
	expected := "https://test.atlassian.net"
	if url != expected {
		t.Errorf("expected '%s', got '%s'", expected, url)
	}
}

func TestLoadFromEnv_EmptyEnv(t *testing.T) {
	// Save original env
	origSite := os.Getenv(EnvSite)
	origEmail := os.Getenv(EnvEmail)
	origToken := os.Getenv(EnvToken)
	defer func() {
		os.Setenv(EnvSite, origSite)
		os.Setenv(EnvEmail, origEmail)
		os.Setenv(EnvToken, origToken)
	}()

	// Clear env vars
	os.Unsetenv(EnvSite)
	os.Unsetenv(EnvEmail)
	os.Unsetenv(EnvToken)

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv should not error: %v", err)
	}

	// Values should be empty
	if cfg.Site != "" || cfg.Email != "" || cfg.Token != "" {
		t.Error("expected all fields to be empty when env vars not set")
	}

	// Validate should fail
	err = cfg.Validate()
	if err == nil {
		t.Error("expected Validate to fail with empty config")
	}
}
