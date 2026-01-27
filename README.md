# atl-cli

A lightweight, agent-friendly CLI for Atlassian Cloud that supports querying Jira issues and Confluence pages.

All output is JSON to stdout, errors to stderr.

## Features

- **Jira**: Retrieve issue details by key
- **Confluence**: Retrieve page content by ID
- **Doctor**: Validate configuration and test API connectivity
- **Debug mode**: Request/response logging with redacted credentials
- **JSON output**: Machine-readable output for scripting and AI agents

## Installation

### From source

Requires Go 1.24 or later.

```bash
git clone https://github.com/martin/atl-cli.git
cd atl-cli
go build ./cmd/atl-cli
```

This creates the `atl-cli` binary in the current directory.

### Install to GOPATH

```bash
go install ./cmd/atl-cli
```

This installs `atl-cli` to `$GOPATH/bin` (typically `~/go/bin`).

Ensure `$GOPATH/bin` is in your PATH by adding this to your `~/.bashrc` or `~/.zshrc`:

```bash
export PATH="$PATH:$HOME/go/bin"
```

Then reload your shell: `source ~/.bashrc` (or restart your terminal).

## Configuration

Set the following environment variables:

| Variable | Description |
|----------|-------------|
| `ATL_CLI_SITE` | Your Atlassian site (e.g., `acme.atlassian.net`) |
| `ATL_CLI_EMAIL` | Your Atlassian account email |
| `ATL_CLI_TOKEN` | Your Atlassian API token |

### Getting an API token

1. Go to https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Copy the token and set it as `ATL_CLI_TOKEN`

### Example setup

```bash
export ATL_CLI_SITE="acme.atlassian.net"
export ATL_CLI_EMAIL="user@example.com"
export ATL_CLI_TOKEN="your-api-token"
```

## Usage

### Validate configuration

Check that your environment is configured correctly:

```bash
atl-cli doctor
```

Output:
```json
{
  "status": "ok",
  "checks": [
    {"name": "ATL_CLI_SITE", "status": "ok"},
    {"name": "ATL_CLI_EMAIL", "status": "ok"},
    {"name": "ATL_CLI_TOKEN", "status": "ok"},
    {"name": "jira_connectivity", "status": "ok"},
    {"name": "confluence_connectivity", "status": "ok"}
  ]
}
```

### Get a Jira issue

```bash
atl-cli jira issue get PROJ-123
```

Output:
```json
{
  "key": "PROJ-123",
  "summary": "Issue title",
  "status": "In Progress",
  "type": "Task",
  "priority": "Medium",
  "assignee": "user@example.com",
  "reporter": "reporter@example.com",
  "created": "2024-01-15T10:30:00Z",
  "updated": "2024-01-16T14:22:00Z",
  "description": "Issue description in markdown...",
  "url": "https://acme.atlassian.net/browse/PROJ-123"
}
```

### Get a Confluence page

```bash
atl-cli confluence page get 12345678
```

Output:
```json
{
  "id": "12345678",
  "title": "Page Title",
  "spaceKey": "SPACE",
  "version": 5,
  "createdAt": "2024-01-10T09:00:00Z",
  "updatedAt": "2024-01-20T11:30:00Z",
  "body": "Page content...",
  "url": "https://acme.atlassian.net/wiki/spaces/SPACE/pages/12345678"
}
```

### Debug mode

Enable debug output to see HTTP requests and responses (credentials are redacted):

```bash
atl-cli --debug jira issue get PROJ-123
```

## Getting help

```bash
# General help
atl-cli --help

# Command-specific help
atl-cli jira --help
atl-cli jira issue --help
atl-cli jira issue get --help
atl-cli confluence --help
atl-cli confluence page --help
atl-cli doctor --help
```

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (config, validation, API, etc.) |

## Error output

Errors are written to stderr as JSON:

```json
{
  "error": "not_found",
  "message": "Issue PROJ-999 not found"
}
```

Error types:
- `config_error` - Missing or invalid environment variables
- `validation_error` - Invalid input (issue key, page ID)
- `auth_error` - Authentication failed (401)
- `not_found` - Resource not found (404)
- `rate_limit` - Rate limited (429)
- `timeout` - Request timed out
- `server_error` - Server error (5xx)

## Development

### Build

```bash
go build ./cmd/atl-cli
```

### Test

```bash
go test ./...
```

### Lint

```bash
go vet ./...
```

## License

MIT
