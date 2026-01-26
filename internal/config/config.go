package config

import (
	"errors"
	"os"
)

// Config holds Atlassian API configuration
type Config struct {
	Site  string // e.g., "your-site.atlassian.net"
	Email string // Atlassian account email
	Token string // API token (never log this)
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		Site:  os.Getenv("ATLASSIAN_SITE"),
		Email: os.Getenv("ATLASSIAN_EMAIL"),
		Token: os.Getenv("ATLASSIAN_TOKEN"),
	}

	return cfg, nil
}

// Validate checks that all required configuration is present
func (c *Config) Validate() error {
	if c.Site == "" {
		return errors.New("ATLASSIAN_SITE environment variable is required")
	}
	if c.Email == "" {
		return errors.New("ATLASSIAN_EMAIL environment variable is required")
	}
	if c.Token == "" {
		return errors.New("ATLASSIAN_TOKEN environment variable is required")
	}
	return nil
}

// BaseURL returns the Atlassian site base URL
func (c *Config) BaseURL() string {
	return "https://" + c.Site
}
