# Tasks: atl-cli v1 Core Commands

**Input**: Design documents from `/specs/001-v1-core/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Required per Constitution Principle IV (Test-First Development)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Based on plan.md structure:
- **CLI commands**: `internal/cli/`
- **Jira client**: `internal/jira/`
- **Confluence client**: `internal/confluence/`
- **HTTP utilities**: `internal/httpclient/`
- **Config**: `internal/config/` (exists)
- **Tests**: Colocated `*_test.go` files

---

## Phase 1: Setup

**Purpose**: Create foundational packages and shared infrastructure

- [x] T001 Create internal/httpclient/ package directory
- [x] T002 Create tests/integration/ directory with README.md

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests for Foundational

- [x] T003 [P] Write tests for HTTP client in internal/httpclient/client_test.go (timeout, auth, debug output, error mapping)
- [x] T004 [P] Write tests for config validation in internal/config/config_test.go (env vars present/missing/empty)

### Implementation for Foundational

- [x] T005 Implement shared HTTP client in internal/httpclient/client.go (30s timeout, Basic Auth, HTTPS enforcement)
- [x] T006 Implement debug output in internal/httpclient/debug.go (request logging with redacted auth)
- [x] T007 Implement error types in internal/httpclient/errors.go (ErrorResponse struct, error type constants)
- [x] T008 Add config validation tests passing in internal/config/config.go (ensure Validate() works correctly)

**Checkpoint**: Foundation ready - run `go test ./internal/httpclient/... ./internal/config/...` - all tests must pass

---

## Phase 3: User Story 1 - Fetch Jira Issue Details (Priority: P1) 🎯 MVP

**Goal**: Users can run `atl-cli jira issue get PROJ-123` and receive JSON with issue details

**Independent Test**: `go build ./cmd/atl-cli && ./atl-cli jira issue get <real-issue-key>` returns valid JSON with all required fields

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T009 [P] [US1] Write tests for Issue struct and parsing in internal/jira/issue_test.go
- [x] T010 [P] [US1] Write tests for Jira client GetIssue in internal/jira/client_test.go (mock HTTP responses for success, 404, 401, timeout)
- [x] T011 [P] [US1] Write tests for issue key validation in internal/jira/validation_test.go (valid/invalid formats)

### Implementation for User Story 1

- [x] T012 [P] [US1] Create Issue struct in internal/jira/issue.go (per data-model.md)
- [x] T013 [P] [US1] Implement HTML stripping utility in internal/jira/html.go (for renderedFields.description)
- [x] T014 [US1] Implement Jira client in internal/jira/client.go (GetIssue method using httpclient)
- [x] T015 [US1] Implement issue key validation in internal/jira/validation.go (regex pattern from data-model.md)
- [x] T016 [US1] Implement API response parsing in internal/jira/client.go (map API response to Issue struct)
- [x] T017 [US1] Create jira command group in internal/cli/jira.go (parent command for jira subcommands)
- [x] T018 [US1] Implement `jira issue get` command in internal/cli/jira.go (load config, call client, output JSON)
- [x] T019 [US1] Add error handling for missing config in internal/cli/jira.go (JSON error to stderr, exit 1)
- [x] T020 [US1] Add --debug flag integration in internal/cli/jira.go (pass debug flag to client)

**Checkpoint**: Run `go test ./internal/jira/...` - all tests pass. Build and test: `./atl-cli jira issue get <key>` works

---

## Phase 4: User Story 2 - Fetch Confluence Page Content (Priority: P2)

**Goal**: Users can run `atl-cli confluence page get 123456789` and receive JSON with page content

**Independent Test**: `./atl-cli confluence page get <real-page-id>` returns valid JSON with all required fields

### Tests for User Story 2

- [x] T021 [P] [US2] Write tests for Page struct and parsing in internal/confluence/page_test.go
- [x] T022 [P] [US2] Write tests for Confluence client GetPage in internal/confluence/client_test.go (mock HTTP responses)
- [x] T023 [P] [US2] Write tests for page ID validation in internal/confluence/validation_test.go (numeric only)

### Implementation for User Story 2

- [x] T024 [P] [US2] Create Page struct in internal/confluence/page.go (per data-model.md)
- [x] T025 [US2] Implement Confluence client in internal/confluence/client.go (GetPage method using httpclient)
- [x] T026 [US2] Implement page ID validation in internal/confluence/validation.go (numeric check)
- [x] T027 [US2] Implement API response parsing in internal/confluence/client.go (map API response to Page struct)
- [x] T028 [US2] Create confluence command group in internal/cli/confluence.go (parent command for confluence subcommands)
- [x] T029 [US2] Implement `confluence page get` command in internal/cli/confluence.go (load config, call client, output JSON)
- [x] T030 [US2] Add error handling for missing config in internal/cli/confluence.go
- [x] T031 [US2] Add --debug flag integration in internal/cli/confluence.go

**Checkpoint**: Run `go test ./internal/confluence/...` - all tests pass. Build and test: `./atl-cli confluence page get <id>` works

---

## Phase 5: User Story 3 - Validate Configuration and Connectivity (Priority: P3)

**Goal**: Users can run `atl-cli doctor` to verify configuration and API connectivity

**Independent Test**: `./atl-cli doctor` returns JSON with status of all checks (env vars + API connectivity)

### Tests for User Story 3

- [x] T032 [P] [US3] Write tests for DoctorResult struct in internal/cli/doctor_test.go
- [x] T033 [P] [US3] Write tests for env var checks in internal/cli/doctor_test.go (all present, some missing, all missing)
- [x] T034 [P] [US3] Write tests for API connectivity checks in internal/cli/doctor_test.go (mock responses)

### Implementation for User Story 3

- [x] T035 [US3] Create DoctorResult and Check structs in internal/cli/doctor.go (per data-model.md)
- [x] T036 [US3] Implement env var presence checks in internal/cli/doctor.go
- [x] T037 [US3] Add Jira connectivity check to internal/jira/client.go (CheckConnectivity method using /myself endpoint)
- [x] T038 [US3] Add Confluence connectivity check to internal/confluence/client.go (CheckConnectivity method using /spaces?limit=1)
- [x] T039 [US3] Implement `doctor` command in internal/cli/doctor.go (run all checks, output JSON result)
- [x] T040 [US3] Handle partial failures in doctor (some checks pass, some fail, still output complete result)
- [x] T041 [US3] Add --debug flag integration in internal/cli/doctor.go

**Checkpoint**: Run `go test ./internal/cli/...` - doctor tests pass. Build and test: `./atl-cli doctor` works

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T042 [P] Add integration test README in tests/integration/README.md (setup instructions for real API testing)
- [x] T043 [P] Verify JSON output matches contracts/ schemas for all commands
- [x] T044 Run full test suite: `go test ./...` - all tests pass
- [x] T045 Run vet and build: `go vet ./... && go build ./cmd/atl-cli`
- [x] T046 Validate quickstart.md examples work end-to-end
- [x] T047 Update CLAUDE.md with any new patterns discovered during implementation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in priority order (P1 → P2 → P3)
  - Or in parallel if multiple developers available
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - No dependencies on US1
- **User Story 3 (P3)**: Depends on US1 and US2 for connectivity check methods (T037, T038)

### Within Each User Story

1. Tests MUST be written and FAIL before implementation
2. Structs/types before client implementation
3. Client implementation before CLI command
4. Core implementation before error handling/debug integration
5. Story complete before moving to next priority

### Parallel Opportunities

**Foundational Phase**:
```
T003 (httpclient tests) || T004 (config tests)
```

**User Story 1**:
```
T009 (issue tests) || T010 (client tests) || T011 (validation tests)
T012 (Issue struct) || T013 (HTML strip)
```

**User Story 2**:
```
T021 (page tests) || T022 (client tests) || T023 (validation tests)
T024 (Page struct) can start immediately
```

**User Story 3**:
```
T032 (result tests) || T033 (env tests) || T034 (connectivity tests)
```

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Write tests for Issue struct and parsing in internal/jira/issue_test.go"
Task: "Write tests for Jira client GetIssue in internal/jira/client_test.go"
Task: "Write tests for issue key validation in internal/jira/validation_test.go"

# Then launch parallel structs:
Task: "Create Issue struct in internal/jira/issue.go"
Task: "Implement HTML stripping utility in internal/jira/html.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Jira issue get)
4. **STOP and VALIDATE**: Test `atl-cli jira issue get` independently
5. Deploy/demo if ready - CLI can fetch Jira issues!

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test: `atl-cli jira issue get` works → Deploy/Demo (MVP!)
3. Add User Story 2 → Test: `atl-cli confluence page get` works → Deploy/Demo
4. Add User Story 3 → Test: `atl-cli doctor` works → Deploy/Demo
5. Each story adds value without breaking previous stories

### TDD Workflow (Constitution Principle IV)

For each user story:
1. Write test → Run test → Verify it FAILS (red)
2. Implement minimum code to pass test
3. Run test → Verify it PASSES (green)
4. Refactor if needed
5. Repeat for next test

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Tests written first per TDD requirement (Constitution Principle IV)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Token must NEVER appear in any output (Constitution Principle III)
