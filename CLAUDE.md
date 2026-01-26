# atl-cli Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-01-26

## Active Technologies

- Go 1.24.1 + Cobra v1.10.2 (existing), net/http (stdlib) (001-v1-core)

## Project Structure

```text
cmd/atl-cli/          # Main entry point
internal/
  cli/                # Cobra commands (root, jira, confluence, doctor)
  config/             # Environment variable loading and validation
  httpclient/         # Shared HTTP client with auth, debug, error handling
  jira/               # Jira REST API v3 client
  confluence/         # Confluence REST API v2 client
specs/                # Feature specifications
tests/integration/    # Integration tests (require real credentials)
```

## Commands

```bash
# Build
go build ./cmd/atl-cli

# Test
go test ./...

# Vet
go vet ./...
```

## Environment Variables

All use `ATL_CLI_` prefix to avoid collisions:

- `ATL_CLI_SITE` - Atlassian site (e.g., `acme.atlassian.net`)
- `ATL_CLI_EMAIL` - User email for authentication
- `ATL_CLI_TOKEN` - API token (never exposed in output)

## Code Style

Go 1.24.1: Follow standard conventions

### Key Patterns

- **JSON Output**: All commands output JSON to stdout, errors to stderr
- **Exit Codes**: 0 for success, 1 for all errors
- **Debug Mode**: `--debug` flag enables request logging (auth redacted)
- **TDD**: Tests written first (Constitution Principle IV)
- **Token Security**: Token must NEVER appear in any output (Constitution Principle III)

### Error Handling

Use `internal/httpclient.ErrorResponse` for consistent JSON errors:
- `config_error` - Missing/invalid environment variables
- `validation_error` - Invalid input (issue key, page ID)
- `auth_error` - 401 Unauthorized
- `not_found` - 404 Not Found
- `rate_limit` - 429 Too Many Requests (includes retryAfter)
- `timeout` - Request timeout
- `server_error` - 5xx errors

### Testing with HTTP Mocks

Use `httptest.NewTLSServer` and `SetHTTPClient()` method for testing:

```go
server := httptest.NewTLSServer(handler)
client := jira.NewClient(cfg, false)
client.SetHTTPClient(server.Client())
```

## Recent Changes

- 001-v1-core: Implemented Jira, Confluence, and Doctor commands
- Changed env vars from `ATLASSIAN_*` to `ATL_CLI_*` prefix

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
