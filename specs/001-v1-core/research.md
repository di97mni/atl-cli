# Research: atl-cli v1 Core Commands

**Date**: 2026-01-26 | **Feature Branch**: `001-v1-core`

## Atlassian Cloud REST API Integration

### Decision: Use Jira REST API v3 with renderedFields expansion

**Rationale**: The Jira v3 API returns description in ADF (Atlassian Document Format) by default. To extract plain text without implementing a complex ADF parser, we use `?expand=renderedFields` which provides HTML-rendered versions of rich text fields. We then strip HTML tags to get plain text. This is simpler than parsing ADF JSON and doesn't require external dependencies.

**Alternatives considered**:
- Parse ADF JSON directly → Rejected: Complex nested structure, requires custom parser or external library
- Use Jira v2 API → Rejected: Returns wiki markdown which still needs parsing, v3 is the current API
- Return raw ADF → Rejected: User clarification specified plain text only

### Decision: Use Confluence REST API v2 with storage format

**Rationale**: The Confluence v2 API (`/wiki/api/v2/pages/{id}`) is the current recommended API. The `?body-format=storage` parameter returns the page body in storage format (XHTML-like markup), which is the canonical format for Confluence content. Without this parameter, the body field is empty.

**Alternatives considered**:
- Use v1 API → Rejected: Deprecated, v2 is current
- Use atlas_doc_format → Rejected: Returns ADF which is more complex; user clarification specified storage format only

## API Endpoints

### Jira: Get Issue

```
GET https://{site}.atlassian.net/rest/api/3/issue/{issueIdOrKey}?expand=renderedFields
```

**Response fields used**:
- `key` - Issue key (e.g., "PROJ-123")
- `fields.summary` - Issue title
- `fields.status.name` - Current status
- `fields.assignee.displayName` - Assignee (null if unassigned)
- `fields.priority.name` - Priority (null if not set)
- `fields.created` - ISO 8601 timestamp
- `fields.updated` - ISO 8601 timestamp
- `renderedFields.description` - HTML description (strip to plain text)
- Constructed: `https://{site}.atlassian.net/browse/{key}` - Browse URL

### Confluence: Get Page

```
GET https://{site}.atlassian.net/wiki/api/v2/pages/{id}?body-format=storage
```

**Response fields used**:
- `id` - Page ID (string, numeric)
- `title` - Page title
- `spaceId` - Space ID (need to resolve to space key)
- `version.number` - Version number
- `version.createdAt` - Last modified timestamp
- `body.storage.value` - Body content in storage format

**Note**: The v2 API returns `spaceId` instead of space key. We either:
1. Make a second API call to get space details, or
2. Return spaceId and document this difference

**Decision**: Return spaceId as `spaceKey` field. The v2 API only provides the numeric spaceId; getting the actual key requires an additional API call which impacts performance. Document that for v1, the "space key" is actually the space ID.

### Jira: Validate Credentials (for doctor)

```
GET https://{site}.atlassian.net/rest/api/3/myself
```

Returns the authenticated user's details. A successful response confirms valid Jira credentials.

### Confluence: Validate Credentials (for doctor)

```
GET https://{site}.atlassian.net/wiki/api/v2/spaces?limit=1
```

Returns at least one space if credentials are valid. A successful response confirms valid Confluence credentials.

## Authentication

**Method**: HTTP Basic Authentication
**Format**: `Authorization: Basic <base64(email:token)>`

Go implementation:
```go
req.SetBasicAuth(email, token)
```

## Error Responses

### Jira Error Format

```json
{
  "errorMessages": ["Issue Does Not Exist"],
  "errors": {}
}
```

### Confluence Error Format

```json
{
  "errors": [
    {
      "status": 404,
      "code": "PAGE_NOT_FOUND",
      "title": "Page not found"
    }
  ]
}
```

### HTTP Status Code Mapping

| Status | Error Type | User Message |
|--------|------------|--------------|
| 401 | `auth_error` | Authentication failed - check credentials |
| 403 | `permission_error` | Access denied - check permissions |
| 404 | `not_found` | Resource not found |
| 429 | `rate_limit` | Rate limit exceeded (include retry-after) |
| 5xx | `server_error` | Atlassian service error |
| timeout | `timeout` | Request timed out after 30s |

## Rate Limiting

**Headers to capture in debug mode**:
- `X-RateLimit-Remaining` - Remaining quota
- `X-RateLimit-Reset` - Reset timestamp
- `Retry-After` - Seconds to wait (only on 429)

**Behavior**: Report error with retry-after value; no automatic retry (per clarification).

## Plain Text Extraction from HTML

For converting `renderedFields.description` (HTML) to plain text:

```go
import "regexp"

var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

func stripHTML(html string) string {
    text := htmlTagRegex.ReplaceAllString(html, "")
    // Also handle HTML entities: &amp; &lt; &gt; &quot; &nbsp;
    text = strings.ReplaceAll(text, "&amp;", "&")
    text = strings.ReplaceAll(text, "&lt;", "<")
    text = strings.ReplaceAll(text, "&gt;", ">")
    text = strings.ReplaceAll(text, "&quot;", "\"")
    text = strings.ReplaceAll(text, "&nbsp;", " ")
    return strings.TrimSpace(text)
}
```

This simple approach handles most common cases without external dependencies.

## Debug Output Format

When `--debug` flag is set, output to stderr before the JSON response:

```
DEBUG: GET https://acme.atlassian.net/rest/api/3/issue/PROJ-123
DEBUG: Authorization: Basic [REDACTED]
DEBUG: Response: 200 OK
DEBUG: X-RateLimit-Remaining: 99500
```

Token value must never appear in debug output.
