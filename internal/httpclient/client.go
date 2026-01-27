// Package httpclient provides a shared HTTP client for Atlassian API requests.
package httpclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// DefaultTimeout is the default request timeout (30 seconds per spec)
	DefaultTimeout = 30 * time.Second
)

// Client is an HTTP client configured for Atlassian API requests.
type Client struct {
	http  *http.Client
	email string
	token string
	debug bool
}

// New creates a new HTTP client with the given credentials.
func New(email, token string, debug bool) *Client {
	return &Client{
		http: &http.Client{
			Timeout: DefaultTimeout,
		},
		email: email,
		token: token,
		debug: debug,
	}
}

// NewRequest creates a new HTTP request with authentication headers.
// Returns an error if the URL is not HTTPS.
func (c *Client) NewRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	// Enforce HTTPS
	if !strings.HasPrefix(url, "https://") {
		return nil, errors.New("HTTPS is required for all API requests")
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	// Set authentication
	req.SetBasicAuth(c.email, c.token)

	// Set standard headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// Do executes an HTTP request and returns the response.
// Debug output is written to stderr if debug mode is enabled.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// Ensure auth is set (in case request was created externally)
	if req.Header.Get("Authorization") == "" {
		req.SetBasicAuth(c.email, c.token)
	}

	// Debug output before request
	if c.debug {
		DebugRequest(req)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		if c.debug {
			DebugError(err)
		}
		return nil, err
	}

	// Debug output after response
	if c.debug {
		DebugResponse(resp)
	}

	return resp, nil
}

// IsDebug returns whether debug mode is enabled.
func (c *Client) IsDebug() bool {
	return c.debug
}

// SetHTTPClient sets the underlying HTTP client (for testing).
func (c *Client) SetHTTPClient(httpClient *http.Client) {
	c.http = httpClient
}
