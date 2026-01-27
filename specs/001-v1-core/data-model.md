# Data Model: atl-cli v1 Core Commands

**Date**: 2026-01-26 | **Feature Branch**: `001-v1-core`

## Entities

### JiraIssue

Represents a Jira issue returned by `atl-cli jira issue get`.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| key | string | Yes | Issue key (e.g., "PROJ-123") |
| summary | string | Yes | Issue title/summary |
| status | string | Yes | Current status name |
| assignee | string | No | Assignee display name (null if unassigned) |
| priority | string | No | Priority name (null if not set) |
| created | string | Yes | ISO 8601 creation timestamp |
| updated | string | Yes | ISO 8601 last update timestamp |
| description | string | Yes | Plain text description (empty string if none) |
| url | string | Yes | Browse URL for the issue |

**Go struct**:
```go
type Issue struct {
    Key         string  `json:"key"`
    Summary     string  `json:"summary"`
    Status      string  `json:"status"`
    Assignee    *string `json:"assignee"`    // null if unassigned
    Priority    *string `json:"priority"`    // null if not set
    Created     string  `json:"created"`
    Updated     string  `json:"updated"`
    Description string  `json:"description"`
    URL         string  `json:"url"`
}
```

### ConfluencePage

Represents a Confluence page returned by `atl-cli confluence page get`.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string | Yes | Page ID (numeric string) |
| title | string | Yes | Page title |
| spaceKey | string | Yes | Space ID (numeric, see note) |
| version | int | Yes | Version number |
| updated | string | Yes | ISO 8601 last modified timestamp |
| body | string | Yes | Body content in storage format |

**Note**: The Confluence v2 API returns `spaceId` (numeric) rather than the traditional space key. For v1, we return this as `spaceKey` and document that it's the numeric ID.

**Go struct**:
```go
type Page struct {
    ID       string `json:"id"`
    Title    string `json:"title"`
    SpaceKey string `json:"spaceKey"`  // Actually spaceId from API
    Version  int    `json:"version"`
    Updated  string `json:"updated"`
    Body     string `json:"body"`
}
```

### DoctorResult

Represents the health check results from `atl-cli doctor`.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| status | string | Yes | Overall status: "ok" or "error" |
| checks | []Check | Yes | Array of individual check results |

**Check struct**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | Yes | Check name (e.g., "ATLASSIAN_SITE") |
| status | string | Yes | "ok", "missing", or "error" |
| message | string | No | Additional details (only on error/missing) |

**Go structs**:
```go
type DoctorResult struct {
    Status string  `json:"status"` // "ok" or "error"
    Checks []Check `json:"checks"`
}

type Check struct {
    Name    string  `json:"name"`
    Status  string  `json:"status"`  // "ok", "missing", "error"
    Message *string `json:"message,omitempty"`
}
```

### ErrorResponse

Standard error response format for all commands.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| error | string | Yes | Error type (e.g., "not_found", "auth_error") |
| message | string | Yes | Human-readable error message |

**Go struct**:
```go
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
}
```

**Error types**:
- `config_error` - Missing or invalid configuration
- `validation_error` - Invalid input (e.g., bad issue key format)
- `auth_error` - Authentication failed
- `permission_error` - Access denied
- `not_found` - Resource not found
- `rate_limit` - Rate limit exceeded
- `timeout` - Request timed out
- `server_error` - Atlassian service error
- `unknown_error` - Unexpected error

### Config

Runtime configuration (internal, not output).

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Site | string | Yes | Atlassian site (e.g., "acme.atlassian.net") |
| Email | string | Yes | User email for authentication |
| Token | string | Yes | API token (NEVER exposed in output) |

**Note**: This struct already exists in `internal/config/config.go`.

## Validation Rules

### Issue Key Format
- Pattern: `^[A-Z][A-Z0-9]+-[0-9]+$`
- Examples: "PROJ-123", "ABC-1", "TEST123-99999"
- Invalid: "123", "proj-123", "PROJ", "PROJ-"

### Page ID Format
- Pattern: `^[0-9]+$`
- Must be numeric string
- Invalid: "abc", "12.34", "-1"

### Environment Variables
- ATLASSIAN_SITE: Required, non-empty
- ATLASSIAN_EMAIL: Required, non-empty
- ATLASSIAN_TOKEN: Required, non-empty

## State Transitions

N/A - All entities are read-only snapshots. No state transitions in v1.
