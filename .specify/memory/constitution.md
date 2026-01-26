<!--
Sync Impact Report
==================
Version change: N/A (new) → 1.0.0
Modified principles: N/A (initial constitution)
Added sections:
  - Core Principles (7 principles)
  - Security Requirements
  - Development Workflow
  - Governance
Removed sections: N/A
Templates requiring updates:
  - .specify/templates/plan-template.md ✅ compatible (no changes needed)
  - .specify/templates/spec-template.md ✅ compatible (no changes needed)
  - .specify/templates/tasks-template.md ✅ compatible (no changes needed)
Follow-up TODOs: None
-->

# atl-cli Constitution

## Core Principles

### I. Single Binary Distribution

atl-cli MUST compile to a single, statically-linked Go binary with zero runtime dependencies.
Distribution is via direct binary download or `go install`. No package managers, containers,
or interpreters required. Cross-compilation targets Linux (primary) and macOS (best-effort).

### II. JSON-First Output

All commands MUST output structured JSON to stdout by default. Errors MUST go to stderr as
JSON objects with `error` and `message` fields. Human-readable formats are optional (`--format`)
but JSON is the canonical output. Output schemas MUST be stable within major versions.

### III. Security by Default (NON-NEGOTIABLE)

- Tokens and credentials MUST NEVER be printed, logged, or written to disk
- `--debug` mode MUST redact Authorization headers and token values
- Environment variables are the only supported credential source in v1
- No credential caching or storage mechanisms
- All HTTP requests MUST use HTTPS

### IV. Test-First Development

TDD is mandatory. The workflow is: write test → verify it fails → implement → verify it passes.
Use Go's standard `testing` package with table-driven tests. Minimum coverage targets:
- Config loading: 100%
- API clients: 90% (mock HTTP responses)
- CLI commands: 80% (integration tests)

### V. Read-Only v1 (NON-NEGOTIABLE)

Version 1.x MUST NOT include any mutation operations. No creating issues, updating tickets,
editing pages, or modifying any Atlassian data. Read-only operations only: get issue, get page,
validate credentials (doctor). Mutation operations are deferred to v2.

### VI. Modular Clients

Jira and Confluence clients MUST be separate internal packages (`internal/jira`, `internal/confluence`).
Each client is independently testable with its own mock server tests. Shared HTTP utilities
live in a common package but clients MUST NOT import each other.

### VII. Environment-Based Configuration

v1 configuration comes exclusively from environment variables:
- `ATLASSIAN_SITE` (required): Site hostname (e.g., `acme.atlassian.net`)
- `ATLASSIAN_EMAIL` (required): User email for authentication
- `ATLASSIAN_TOKEN` (required): API token (PAT)

CLI flags MAY override environment variables. Config files are deferred to v1.1.

## Security Requirements

These requirements are derived from Principle III and are NON-NEGOTIABLE:

1. **Credential Handling**: The `Config` struct MUST NOT implement `String()` or `GoString()`
   methods that could leak tokens. Use explicit getter methods.

2. **Debug Output**: When `--debug` is enabled, log request URLs, response status codes,
   and Atlassian trace headers. Replace Authorization header values with `[REDACTED]`.

3. **Error Messages**: Never include token values in error messages. Use placeholders like
   "token starting with 'ABC...'" if token identification is needed.

4. **Dependency Audit**: Before release, run `go mod verify` and review dependencies
   for known vulnerabilities.

## Development Workflow

### Code Organization

```
cmd/atl-cli/main.go     # Entry point only - minimal code
internal/cli/           # Cobra commands and flag handling
internal/config/        # Environment loading and validation
internal/jira/          # Jira REST API client
internal/confluence/    # Confluence REST API client
```

### Testing Requirements

1. **Unit Tests**: Colocated with source (`*_test.go` files)
2. **Integration Tests**: In `tests/` directory, require test credentials
3. **Contract Tests**: Verify JSON output schema stability

### Build and Release

- Build: `go build -ldflags "-X main.Version=..." ./cmd/atl-cli`
- Test: `go test ./...`
- Lint: `golangci-lint run`
- All tests MUST pass before merge

## Governance

This constitution supersedes all other development practices for atl-cli.

**Amendment Process**:
1. Propose change with rationale in a GitHub issue
2. Discuss impact on existing code and tests
3. Update constitution with version bump
4. Migrate existing code if principles change

**Compliance**:
- All code reviews MUST verify compliance with these principles
- NON-NEGOTIABLE principles cannot be violated under any circumstances
- Other principles may be relaxed with documented justification in the PR

**Version Policy**:
- MAJOR: Backward-incompatible principle changes or removals
- MINOR: New principles or significant clarifications
- PATCH: Typo fixes, minor wording improvements

**Version**: 1.0.0 | **Ratified**: 2026-01-26 | **Last Amended**: 2026-01-26
