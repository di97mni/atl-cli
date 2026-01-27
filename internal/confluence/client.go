package confluence

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/martin/atl-cli/internal/config"
	"github.com/martin/atl-cli/internal/httpclient"
)

// Client is a Confluence REST API client.
type Client struct {
	cfg        *config.Config
	httpClient *httpclient.Client
}

// NewClient creates a new Confluence client.
func NewClient(cfg *config.Config, debug bool) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: httpclient.New(cfg.Email, cfg.Token, debug),
	}
}

// GetPage retrieves a Confluence page by its ID.
func (c *Client) GetPage(ctx context.Context, id string) (*Page, error) {
	// Validate page ID format
	if err := ValidatePageID(id); err != nil {
		return nil, err
	}

	// Build request URL - Confluence v2 API
	url := fmt.Sprintf("%s/wiki/api/v2/pages/%s?body-format=storage", c.cfg.BaseURL(), id)

	req, err := c.httpClient.NewRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
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

	page, err := ParseAPIResponse(body)
	if err != nil {
		return nil, err
	}

	return page, nil
}

// CheckConnectivity verifies the Confluence API connection.
func (c *Client) CheckConnectivity(ctx context.Context) error {
	url := fmt.Sprintf("%s/wiki/api/v2/spaces?limit=1", c.cfg.BaseURL())

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
