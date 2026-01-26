package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// Error types for JSON output
const (
	ErrTypeConfig     = "config_error"
	ErrTypeValidation = "validation_error"
	ErrTypeAuth       = "auth_error"
	ErrTypePermission = "permission_error"
	ErrTypeNotFound   = "not_found"
	ErrTypeRateLimit  = "rate_limit"
	ErrTypeTimeout    = "timeout"
	ErrTypeServer     = "server_error"
	ErrTypeUnknown    = "unknown_error"
)

// ErrorResponse is the standard error format for CLI output.
type ErrorResponse struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	RetryAfter int    `json:"retryAfter,omitempty"` // Only for rate_limit errors
}

// NewErrorResponse creates an ErrorResponse from an HTTP response.
func NewErrorResponse(resp *http.Response) *ErrorResponse {
	errType := mapStatusToErrorType(resp.StatusCode)
	message := defaultMessage(errType)

	// Try to extract message from response body
	if resp.Body != nil {
		body, err := io.ReadAll(resp.Body)
		if err == nil && len(body) > 0 {
			// Try Jira error format
			var jiraErr struct {
				ErrorMessages []string `json:"errorMessages"`
			}
			if json.Unmarshal(body, &jiraErr) == nil && len(jiraErr.ErrorMessages) > 0 {
				message = jiraErr.ErrorMessages[0]
			}

			// Try Confluence error format
			var confErr struct {
				Errors []struct {
					Title string `json:"title"`
				} `json:"errors"`
			}
			if json.Unmarshal(body, &confErr) == nil && len(confErr.Errors) > 0 {
				message = confErr.Errors[0].Title
			}
		}
	}

	errResp := &ErrorResponse{
		Error:   errType,
		Message: message,
	}

	// Add retry-after for rate limit errors
	if errType == ErrTypeRateLimit {
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				errResp.RetryAfter = seconds
			}
		}
	}

	return errResp
}

// NewConfigError creates a config error response.
func NewConfigError(message string) *ErrorResponse {
	return &ErrorResponse{
		Error:   ErrTypeConfig,
		Message: message,
	}
}

// NewValidationError creates a validation error response.
func NewValidationError(message string) *ErrorResponse {
	return &ErrorResponse{
		Error:   ErrTypeValidation,
		Message: message,
	}
}

// NewTimeoutError creates a timeout error response.
func NewTimeoutError() *ErrorResponse {
	return &ErrorResponse{
		Error:   ErrTypeTimeout,
		Message: "Request timed out after 30 seconds",
	}
}

// Write writes the error response as JSON to the given writer.
func (e *ErrorResponse) Write(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(e)
}

// String returns the error as a string (for error interface).
func (e *ErrorResponse) String() string {
	return fmt.Sprintf("%s: %s", e.Error, e.Message)
}

func mapStatusToErrorType(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return ErrTypeAuth
	case http.StatusForbidden:
		return ErrTypePermission
	case http.StatusNotFound:
		return ErrTypeNotFound
	case http.StatusTooManyRequests:
		return ErrTypeRateLimit
	default:
		if status >= 500 {
			return ErrTypeServer
		}
		return ErrTypeUnknown
	}
}

func defaultMessage(errType string) string {
	switch errType {
	case ErrTypeAuth:
		return "Authentication failed - check your credentials"
	case ErrTypePermission:
		return "Access denied - check your permissions"
	case ErrTypeNotFound:
		return "Resource not found"
	case ErrTypeRateLimit:
		return "Rate limit exceeded"
	case ErrTypeServer:
		return "Atlassian service error"
	default:
		return "An unexpected error occurred"
	}
}
