# PRD: `atl-cli` — Minimal Atlassian Cloud CLI (Jira + Confluence)

## 1) Overview
`atl-cli` is a lightweight, agent-friendly CLI for Atlassian Cloud that supports a small set of high-value workflows for Jira and Confluence. It is designed for automated usage (e.g., Claude code agent) and for engineers who want deterministic JSON output without requiring Atlassian admin access or marketplace apps.

## 2) Goals
- Provide a single CLI tool that can query **Jira** and **Confluence** in Atlassian Cloud.
- Support authentication using **Atlassian API tokens (PATs)** owned by the user (no admin required).
- Produce **stable, minimal JSON output** suitable for agent parsing.
- Keep scope intentionally small: expose only the use cases required by the team.

## 3) Non-Goals
- Writing/mutating data (no create/update transitions, no page edits) in v1.
- Full Jira/Confluence API surface area coverage.
- OAuth browser login flow (PAT-only in v1).
- Admin-only operations (user management, global permissions, app install, etc.).

## 4) Target Users
- Developers and SREs who need quick terminal access to tickets and pages.
- Code agents that need scripted access to Jira/Confluence content.
- Users without Atlassian site-admin permissions.

## 5) Supported Platforms
- Ubuntu Linux (primary)
- macOS (best-effort, post-v1)
- Windows (not targeted for v1)

## 6) Authentication & Configuration
**Auth mechanism:** HTTP Basic Auth using:
- `ATLASSIAN_EMAIL`
- `ATLASSIAN_TOKEN` (API token)
- `ATLASSIAN_SITE` (e.g., `your-site.atlassian.net`)

**Config precedence:**
1. CLI flags (if provided)
2. Environment variables
3. Config file (optional v1.1)

**Required environment variables (v1):**
- `ATLASSIAN_SITE`
- `ATLASSIAN_EMAIL`
- `ATLASSIAN_TOKEN`

**Security requirements:**
- Never print token values.
- `--debug` must redact Authorization header.
- Do not write secrets to disk.

## 7) Core Use Cases (v1)

### Jira
1. **Get ticket details**
   - Input: Issue key (e.g., `KEY-123`)
   - Output: minimal structured JSON
   - Must include: key, summary, status, assignee (if any), priority (if any), created, updated, description (raw + optional rendered/plain)

### Confluence
2. **Get page content**
   - Input: page ID (numeric)
   - Output: minimal structured JSON
   - Must include: id, title, space key, version number, last updated timestamp, body (storage format)

## 8) CLI Commands (v1)
Command names are optimized for low ambiguity and agent use.

### Jira
- `atl-cli jira issue get KEY-123 [--json] [--fields ...]`

### Confluence
- `atl-cli confluence page get 123456789 [--json] [--expand ...]`

### Global
- `atl-cli version`
- `atl-cli doctor` (validates env vars + can call `/myself` and Confluence “current user” endpoint if available)
- `--debug` (prints request URL, status, Atlassian trace/request id headers; redacts auth)

## 9) Output Format
Default output is JSON to stdout.

### Jira issue get (normalized output)
```json
{
  "key": "KEY-123",
  "summary": "...",
  "status": "In Progress",
  "assignee": "Jane Doe",
  "priority": "High",
  "created": "2025-01-01T12:00:00.000+0000",
  "updated": "2025-01-03T09:30:00.000+0000",
  "description": {
    "raw": { "...": "ADF or string" },
    "text": "optional plain text representation"
  },
  "url": "https://your-site.atlassian.net/browse/KEY-123"
}
