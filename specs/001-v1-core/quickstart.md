# Quickstart: atl-cli v1 Core Commands

## Prerequisites

1. Go 1.24+ installed
2. Atlassian Cloud account with API token access
3. Read permissions for target Jira projects and Confluence spaces

## Setup

### 1. Generate Atlassian API Token

1. Go to https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Give it a label (e.g., "atl-cli")
4. Copy the token (you won't see it again)

### 2. Configure Environment Variables

```bash
export ATL_CLI_SITE="your-company.atlassian.net"
export ATL_CLI_EMAIL="your-email@example.com"
export ATL_CLI_TOKEN="your-api-token"
```

Add these to your shell profile (`.bashrc`, `.zshrc`, etc.) for persistence.

### 3. Build the CLI

```bash
go build -o atl-cli ./cmd/atl-cli
```

## Usage

### Validate Configuration

Run the doctor command to verify your setup:

```bash
./atl-cli doctor
```

Expected output (all checks pass):
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

### Fetch Jira Issue

```bash
./atl-cli jira issue get PROJ-123
```

Example output:
```json
{
  "key": "PROJ-123",
  "summary": "Implement user authentication",
  "status": "In Progress",
  "assignee": "John Doe",
  "priority": "High",
  "created": "2026-01-15T10:30:00.000Z",
  "updated": "2026-01-25T14:45:00.000Z",
  "description": "Add OAuth2 login flow with Google provider...",
  "url": "https://your-company.atlassian.net/browse/PROJ-123"
}
```

### Fetch Confluence Page

```bash
./atl-cli confluence page get 123456789
```

Example output:
```json
{
  "id": "123456789",
  "title": "Architecture Overview",
  "spaceKey": "65011",
  "version": 5,
  "updated": "2026-01-20T15:45:00.000Z",
  "body": "<p>This document describes the system architecture...</p>"
}
```

### Debug Mode

Add `--debug` to any command to see request details:

```bash
./atl-cli --debug jira issue get PROJ-123
```

Debug output goes to stderr:
```
DEBUG: GET https://your-company.atlassian.net/rest/api/3/issue/PROJ-123
DEBUG: Authorization: Basic [REDACTED]
DEBUG: Response: 200 OK
```

## Error Handling

All errors output JSON to stderr and exit with code 1:

```json
{
  "error": "not_found",
  "message": "Issue PROJ-999 not found"
}
```

### Common Errors

| Error Type | Cause | Resolution |
|------------|-------|------------|
| `config_error` | Missing environment variable | Set all required env vars |
| `auth_error` | Invalid credentials | Regenerate API token |
| `permission_error` | No access to resource | Request project/space access |
| `not_found` | Resource doesn't exist | Verify issue key or page ID |
| `rate_limit` | API quota exceeded | Wait and retry later |
| `timeout` | Request took >30s | Check network, retry |

## Development

### Run Tests

```bash
go test ./...
```

### Build with Version Info

```bash
go build -ldflags "-X github.com/martin/atl-cli/internal/cli.Version=1.0.0" ./cmd/atl-cli
```
