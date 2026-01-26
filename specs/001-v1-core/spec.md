# Feature Specification: atl-cli v1 Core Commands

**Feature Branch**: `001-v1-core`
**Created**: 2026-01-26
**Status**: Draft
**Input**: User description: "atl-cli v1 core commands: jira issue get, confluence page get, and doctor"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Fetch Jira Issue Details (Priority: P1)

A developer or code agent needs to retrieve details about a specific Jira issue to understand its current status, assignee, and description. They run a single command with the issue key and receive structured data they can parse or pipe to other tools.

**Why this priority**: This is the primary use case identified in the PRD. Developers and agents need quick access to ticket information without opening a browser or using the Jira UI.

**Independent Test**: Can be fully tested by running `atl-cli jira issue get KEY-123` with valid credentials and verifying the JSON output contains expected fields.

**Acceptance Scenarios**:

1. **Given** valid credentials are configured via environment variables, **When** user runs `atl-cli jira issue get PROJ-123`, **Then** the system outputs JSON to stdout containing: key, summary, status, assignee, priority, created, updated, description, and URL.

2. **Given** valid credentials are configured, **When** user runs `atl-cli jira issue get INVALID-999` for a non-existent issue, **Then** the system outputs a JSON error to stderr with a clear message indicating the issue was not found.

3. **Given** the ATLASSIAN_TOKEN environment variable is missing, **When** user runs `atl-cli jira issue get PROJ-123`, **Then** the system outputs a JSON error to stderr indicating the missing configuration.

---

### User Story 2 - Fetch Confluence Page Content (Priority: P2)

A developer or code agent needs to retrieve the content of a specific Confluence page by its numeric ID. They receive the page title, space, version, and body content in a structured format.

**Why this priority**: Second most common use case. Enables agents and scripts to access documentation and wiki content programmatically.

**Independent Test**: Can be fully tested by running `atl-cli confluence page get 123456789` with valid credentials and verifying the JSON output contains expected fields.

**Acceptance Scenarios**:

1. **Given** valid credentials are configured via environment variables, **When** user runs `atl-cli confluence page get 123456789`, **Then** the system outputs JSON to stdout containing: id, title, space key, version number, last updated timestamp, and body (storage format).

2. **Given** valid credentials are configured, **When** user runs `atl-cli confluence page get 999999999` for a non-existent page, **Then** the system outputs a JSON error to stderr with a clear message indicating the page was not found.

3. **Given** valid credentials are configured, **When** user runs `atl-cli confluence page get abc` with an invalid (non-numeric) ID, **Then** the system outputs a JSON error to stderr indicating the ID must be numeric.

---

### User Story 3 - Validate Configuration and Connectivity (Priority: P3)

A user setting up atl-cli for the first time wants to verify their configuration is correct and they can connect to both Jira and Confluence APIs before using other commands.

**Why this priority**: Essential for onboarding and troubleshooting, but not the primary workflow. Users run this once during setup or when debugging connectivity issues.

**Independent Test**: Can be fully tested by running `atl-cli doctor` and verifying it reports the status of each configuration item and API connectivity test.

**Acceptance Scenarios**:

1. **Given** all required environment variables are set correctly, **When** user runs `atl-cli doctor`, **Then** the system outputs JSON indicating all checks passed, including: environment variable presence, Jira API connectivity (via /myself endpoint), and Confluence API connectivity.

2. **Given** ATLASSIAN_SITE is missing, **When** user runs `atl-cli doctor`, **Then** the system outputs JSON showing which environment variables are missing and skips API connectivity tests.

3. **Given** environment variables are set but the token is invalid, **When** user runs `atl-cli doctor`, **Then** the system outputs JSON showing environment variables are present but API authentication failed, with the specific error (without exposing the token value).

---

### Edge Cases

- What happens when the Atlassian API is temporarily unavailable? System outputs a JSON error with connection timeout or service unavailable message.
- What happens when the API rate limit is exceeded? System outputs a JSON error indicating rate limiting with retry-after information if provided.
- What happens when the issue key format is invalid (e.g., "123" instead of "PROJ-123")? System outputs a JSON error indicating invalid issue key format.
- What happens when --debug flag is used? System outputs request details (URL, status, headers) to stderr with Authorization header redacted, then normal output to stdout.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST read configuration from environment variables: ATLASSIAN_SITE, ATLASSIAN_EMAIL, ATLASSIAN_TOKEN
- **FR-002**: System MUST output all successful responses as JSON to stdout
- **FR-003**: System MUST output all errors as JSON to stderr with `error` and `message` fields
- **FR-004**: System MUST validate that required environment variables are present before making API calls
- **FR-005**: System MUST support the `--debug` flag to output request/response details with redacted authentication
- **FR-006**: System MUST return a non-zero exit code on any error
- **FR-007**: System MUST NEVER print, log, or expose the ATLASSIAN_TOKEN value
- **FR-008**: Jira issue get MUST return: key, summary, status, assignee (or null), priority (or null), created, updated, description (raw and text), and URL
- **FR-009**: Confluence page get MUST return: id, title, space key, version number, last updated timestamp, and body (storage format)
- **FR-010**: Doctor command MUST check: presence of all required environment variables, Jira API connectivity, Confluence API connectivity
- **FR-011**: System MUST use HTTPS for all API requests

### Key Entities

- **JiraIssue**: Represents a Jira issue with key, summary, status name, assignee display name, priority name, timestamps, description (ADF raw format and plain text), and browse URL
- **ConfluencePage**: Represents a Confluence page with numeric ID, title, space key, version number, last modified timestamp, and body content in storage format
- **DoctorResult**: Represents the health check results with environment variable status, Jira connectivity status, Confluence connectivity status, and any error messages
- **Config**: Represents the runtime configuration loaded from environment variables (site, email, token - token never exposed)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can retrieve a Jira issue's details in under 3 seconds for typical network conditions
- **SC-002**: Users can retrieve a Confluence page's content in under 3 seconds for typical network conditions
- **SC-003**: The doctor command completes all checks in under 10 seconds
- **SC-004**: 100% of error conditions produce valid JSON output to stderr with actionable error messages
- **SC-005**: Zero credential exposure in any output mode, including --debug
- **SC-006**: Output JSON schema remains stable (no breaking changes) within the v1.x release line
- **SC-007**: Users can successfully configure and validate their setup using only the doctor command output

## Assumptions

- Users have valid Atlassian Cloud accounts with API token access (not Atlassian Server/Data Center)
- Users have read permissions for the Jira projects and Confluence spaces they are querying
- Network connectivity to *.atlassian.net is available
- API tokens are generated via Atlassian account settings (https://id.atlassian.com/manage-profile/security/api-tokens)
