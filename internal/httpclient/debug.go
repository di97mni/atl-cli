package httpclient

import (
	"fmt"
	"net/http"
	"os"
)

// DebugRequest prints request details to stderr with redacted auth.
func DebugRequest(req *http.Request) {
	fmt.Fprintf(os.Stderr, "DEBUG: %s %s\n", req.Method, req.URL.String())
	fmt.Fprintf(os.Stderr, "DEBUG: Authorization: Basic [REDACTED]\n")

	// Print other relevant headers (not auth-related)
	for key, values := range req.Header {
		if key == "Authorization" {
			continue // Already handled above
		}
		for _, v := range values {
			fmt.Fprintf(os.Stderr, "DEBUG: %s: %s\n", key, v)
		}
	}
}

// DebugResponse prints response details to stderr.
func DebugResponse(resp *http.Response) {
	fmt.Fprintf(os.Stderr, "DEBUG: Response: %s\n", resp.Status)

	// Print rate limit headers if present
	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		fmt.Fprintf(os.Stderr, "DEBUG: X-RateLimit-Remaining: %s\n", remaining)
	}
	if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
		fmt.Fprintf(os.Stderr, "DEBUG: X-RateLimit-Reset: %s\n", reset)
	}
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		fmt.Fprintf(os.Stderr, "DEBUG: Retry-After: %s\n", retryAfter)
	}
}

// DebugError prints error details to stderr.
func DebugError(err error) {
	fmt.Fprintf(os.Stderr, "DEBUG: Error: %v\n", err)
}
