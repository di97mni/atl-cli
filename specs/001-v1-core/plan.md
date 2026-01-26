# Implementation Plan: atl-cli v1 Core Commands

**Branch**: `001-v1-core` | **Date**: 2026-01-26 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-v1-core/spec.md`

## Summary

Implement three read-only CLI commands for atl-cli v1: `jira issue get`, `confluence page get`, and `doctor`. All commands output JSON to stdout, errors to stderr, and use environment variables for Atlassian Cloud authentication. The implementation follows TDD and the existing Cobra CLI patterns established in the codebase.

## Technical Context

**Language/Version**: Go 1.24.1
**Primary Dependencies**: Cobra v1.10.2 (existing), net/http (stdlib)
**Storage**: N/A (no local storage)
**Testing**: Go testing package with table-driven tests
**Target Platform**: Linux (primary), macOS (best-effort)
**Project Type**: Single binary CLI
**Performance Goals**: <3s for issue/page retrieval, <10s for doctor
**Constraints**: 30s API timeout, no retries, HTTPS only
**Scale/Scope**: Single user CLI tool

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Single Binary Distribution | ✅ PASS | Using Go stdlib + Cobra only |
| II. JSON-First Output | ✅ PASS | All output is JSON to stdout/stderr |
| III. Security by Default | ✅ PASS | Token never logged, HTTPS enforced, --debug redacts auth |
| IV. Test-First Development | ✅ PASS | TDD workflow planned, coverage targets defined |
| V. Read-Only v1 | ✅ PASS | Only GET operations: issue get, page get, doctor |
| VI. Modular Clients | ✅ PASS | Separate internal/jira and internal/confluence packages |
| VII. Environment-Based Config | ✅ PASS | Using existing config.LoadFromEnv() pattern |

**Gate Status**: ALL PASSED - No violations requiring justification

## Project Structure

### Documentation (this feature)

```text
specs/001-v1-core/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   ├── jira-issue.json  # JiraIssue JSON schema
│   ├── confluence-page.json  # ConfluencePage JSON schema
│   ├── doctor-result.json    # DoctorResult JSON schema
│   └── error.json       # Error response JSON schema
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
cmd/atl-cli/
└── main.go                    # Entry point (exists)

internal/
├── cli/
│   ├── root.go               # Root command (exists)
│   ├── version.go            # Version command (exists)
│   ├── jira.go               # jira command group (NEW)
│   ├── confluence.go         # confluence command group (NEW)
│   └── doctor.go             # doctor command (NEW)
├── config/
│   └── config.go             # Config loading (exists)
├── jira/
│   ├── client.go             # Jira HTTP client (NEW)
│   ├── client_test.go        # Client tests (NEW)
│   ├── issue.go              # Issue types and parsing (NEW)
│   └── issue_test.go         # Issue tests (NEW)
├── confluence/
│   ├── client.go             # Confluence HTTP client (NEW)
│   ├── client_test.go        # Client tests (NEW)
│   ├── page.go               # Page types and parsing (NEW)
│   └── page_test.go          # Page tests (NEW)
└── httpclient/
    ├── client.go             # Shared HTTP utilities (NEW)
    └── client_test.go        # HTTP utilities tests (NEW)

tests/
└── integration/
    └── README.md             # Integration test instructions (NEW)
```

**Structure Decision**: Single project with feature-based internal packages. The `internal/httpclient` package provides shared HTTP utilities (timeout, auth, debug output) used by both Jira and Confluence clients without circular dependencies.
