package jira

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/martin/atl-cli/internal/config"
	"github.com/martin/atl-cli/internal/httpclient"
)

// Client is a Jira REST API client.
type Client struct {
	cfg        *config.Config
	httpClient *httpclient.Client
}

// NewClient creates a new Jira client.
func NewClient(cfg *config.Config, debug bool) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: httpclient.New(cfg.Email, cfg.Token, debug),
	}
}

// GetIssue retrieves a Jira issue by its key.
func (c *Client) GetIssue(ctx context.Context, key string) (*Issue, error) {
	// Validate issue key format
	if err := ValidateIssueKey(key); err != nil {
		return nil, err
	}

	// Build request URL
	url := fmt.Sprintf("%s/rest/api/3/issue/%s?expand=renderedFields", c.cfg.BaseURL(), key)

	req, err := c.httpClient.NewRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Check for timeout
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("request timed out")
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		return nil, c.handleError(resp)
	}

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	issue, err := ParseAPIResponse(body, c.cfg.Site)
	if err != nil {
		return nil, err
	}

	return issue, nil
}

// CheckConnectivity verifies the Jira API connection using the /myself endpoint.
func (c *Client) CheckConnectivity(ctx context.Context) error {
	url := fmt.Sprintf("%s/rest/api/3/myself", c.cfg.BaseURL())

	req, err := c.httpClient.NewRequest(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleError(resp)
	}

	return nil
}

func (c *Client) handleError(resp *http.Response) error {
	errResp := httpclient.NewErrorResponse(resp)
	return fmt.Errorf("%s: %s", errResp.Error, errResp.Message)
}

// SetHTTPClient sets the underlying HTTP client (for testing).
func (c *Client) SetHTTPClient(httpClient *http.Client) {
	c.httpClient.SetHTTPClient(httpClient)
}
