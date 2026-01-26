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

// Environment variable names
const (
	EnvSite  = "ATL_CLI_SITE"
	EnvEmail = "ATL_CLI_EMAIL"
	EnvToken = "ATL_CLI_TOKEN"
)

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		Site:  os.Getenv(EnvSite),
		Email: os.Getenv(EnvEmail),
		Token: os.Getenv(EnvToken),
	}

	return cfg, nil
}

// Validate checks that all required configuration is present
func (c *Config) Validate() error {
	if c.Site == "" {
		return errors.New(EnvSite + " environment variable is required")
	}
	if c.Email == "" {
		return errors.New(EnvEmail + " environment variable is required")
	}
	if c.Token == "" {
		return errors.New(EnvToken + " environment variable is required")
	}
	return nil
}

// BaseURL returns the Atlassian site base URL
func (c *Config) BaseURL() string {
	return "https://" + c.Site
}
