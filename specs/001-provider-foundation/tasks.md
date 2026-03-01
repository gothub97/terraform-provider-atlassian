# Tasks: Provider Foundation

**Input**: Design documents from `/specs/001-provider-foundation/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Unit tests (httptest) and acceptance tests (terraform-plugin-testing) are required per Constitution Principle II.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Provider entry point: `main.go`
- Internal packages: `internal/provider/`, `internal/atlassian/`, `internal/jira/`
- Examples: `examples/`
- Build config: `Makefile`, `.goreleaser.yml`, `.golangci.yml`

---

## Phase 1: Setup (Project Initialization)

**Purpose**: Go module, build tooling, linting config, directory structure

- [X] T001 Initialize Go module (`go mod init`) and add terraform-plugin-framework, terraform-plugin-testing dependencies in `go.mod`
- [X] T002 [P] Create `Makefile` with targets: `lint`, `test`, `testacc`, `docs`, `build`, `fmt`
- [X] T003 [P] Create `.golangci.yml` with strict linting config (govet, staticcheck, errcheck, revive, gofmt, gosimple, unused)
- [X] T004 [P] Create `.goreleaser.yml` for cross-platform builds (linux/darwin/windows, amd64/arm64) with GPG signing
- [X] T005 Write provider entry point in `main.go` following terraform-plugin-framework pattern with `providerserver.Serve`, address `registry.terraform.io/atlassian/atlassian`, version from goreleaser ldflags
- [X] T006 Run `go mod tidy && go build -o /dev/null` to verify compilation

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Centralized HTTP client and provider skeleton — MUST complete before ANY resource or data source

**CRITICAL**: No user story work can begin until this phase is complete

- [X] T007 Implement centralized HTTP client in `internal/atlassian/client.go`: constructor accepting base URL + email + token, Basic Auth header on every request, JSON marshal/unmarshal, typed Get/Post/Put/Delete methods
- [X] T008 Implement retry with exponential backoff in `internal/atlassian/client.go`: retry on HTTP 429 and 5xx, up to 5 retries, base 1s with jitter ±20%, max 30s, respect Retry-After header
- [X] T009 Implement generic pagination helper `Paginate[T]` in `internal/atlassian/client.go`: follows Jira startAt/maxResults/total/isLast pattern, accumulates all pages into single slice
- [X] T010 Implement async task poller `WaitForTask` in `internal/atlassian/client.go`: polls `GET /rest/api/3/task/{taskId}` until COMPLETE or FAILED, configurable timeout (default 2m)
- [X] T011 Write unit tests for HTTP client in `internal/atlassian/client_test.go`: test auth header, retry logic (mock 429/5xx responses with httptest), pagination accumulation, task polling, error handling
- [X] T012 Implement provider skeleton in `internal/provider/provider.go`: schema with url/email/api_token (sensitive), env var fallback (ATLASSIAN_URL, ATLASSIAN_EMAIL, ATLASSIAN_API_TOKEN), HCL precedence, Configure method that calls GET /rest/api/3/myself to validate credentials, constructs atlassian.Client and stores in provider data
- [X] T013 Write provider acceptance test in `internal/provider/provider_test.go`: test valid auth succeeds, invalid auth fails with clear error, env var fallback works

**Checkpoint**: Provider compiles, authenticates against Jira Cloud, HTTP client handles retry/pagination. User story implementation can now begin.

---

## Phase 3: User Story 1 — Configure and Authenticate the Provider (Priority: P1) MVP

**Goal**: Provider authenticates and the `jira_myself` data source returns current user info.

**Independent Test**: `terraform plan` with provider block + `jira_myself` data source succeeds.

### Implementation for User Story 1

- [X] T014 [US1] Implement `atlassian_jira_myself` data source in `internal/jira/data_source_myself.go`: schema with all computed attributes (account_id, account_type, display_name, email_address, active, time_zone, locale, self), Read calls GET /rest/api/3/myself
- [X] T015 [US1] Register `atlassian_jira_myself` data source in `internal/provider/provider.go` DataSources method
- [X] T016 [US1] Write unit test for myself data source in `internal/jira/data_source_myself_test.go`: mock /myself response with httptest, verify all fields mapped correctly
- [X] T017 [US1] Write acceptance test in `internal/jira/data_source_myself_test.go`: TestAccDataSourceJiraMyself_basic — verify account_id and display_name are set against live instance
- [X] T018 [US1] Create example HCL in `examples/data-sources/atlassian_jira_myself/data-source.tf`
- [X] T019 [US1] Create provider example HCL in `examples/provider/provider.tf`

**Checkpoint**: Provider initializes, authenticates, `jira_myself` returns user info. MVP complete.

---

## Phase 4: User Story 2 — Manage Jira Projects (Priority: P2)

**Goal**: Full CRUD + import for `atlassian_jira_project` resource and `atlassian_jira_project` data source.

**Independent Test**: Create project, update name, import by key, detect drift, destroy.

### Implementation for User Story 2

- [X] T020 [US2] Implement `atlassian_jira_project` resource in `internal/jira/resource_project.go`: schema per data-model.md (key ForceNew, project_type_key ForceNew, project_template_key auto-derived + ForceNew, api_token Sensitive), CRUD + ImportState, project template auto-derivation from type, validators for key regex and type enum
- [X] T021 [US2] Register `atlassian_jira_project` resource in `internal/provider/provider.go` Resources method
- [X] T022 [US2] Write unit tests in `internal/jira/resource_project_test.go`: test template auto-derivation logic, key validation, request construction with httptest mocks
- [X] T023 [US2] Write acceptance tests in `internal/jira/resource_project_test.go`: TestAccJiraProject_basic (create + verify), TestAccJiraProject_update (change name/description), TestAccJiraProject_import (import by key), TestAccJiraProject_disappears (drift detection)
- [X] T024 [P] [US2] Implement `atlassian_jira_project` data source in `internal/jira/data_source_project.go`: schema with key as required input, all other project attributes computed, Read calls GET /rest/api/3/project/{key}
- [X] T025 [P] [US2] Write acceptance test for project data source in `internal/jira/data_source_project_test.go`: TestAccDataSourceJiraProject_basic — create project via resource, look up via data source, verify attributes match
- [X] T026 [P] [US2] Create example HCL in `examples/resources/atlassian_jira_project/resource.tf` and `examples/data-sources/atlassian_jira_project/data-source.tf`

**Checkpoint**: Projects can be created, updated, imported, drift-detected, and destroyed via Terraform.

---

## Phase 5: User Story 3 — Manage Issue Types (Priority: P3)

**Goal**: Full CRUD + import for `atlassian_jira_issue_type` resource and `atlassian_jira_issue_types` data source.

**Independent Test**: Create issue type, update description, import by ID, destroy.

### Implementation for User Story 3

- [X] T027 [US3] Implement `atlassian_jira_issue_type` resource in `internal/jira/resource_issue_type.go`: schema per data-model.md (hierarchy_level ForceNew, name max 60 chars validator), CRUD + ImportState
- [X] T028 [US3] Register `atlassian_jira_issue_type` resource in `internal/provider/provider.go` Resources method
- [X] T029 [US3] Write unit tests in `internal/jira/resource_issue_type_test.go`: test name validation, hierarchy_level mapping, request/response parsing
- [X] T030 [US3] Write acceptance tests in `internal/jira/resource_issue_type_test.go`: TestAccJiraIssueType_basic, TestAccJiraIssueType_update, TestAccJiraIssueType_import, TestAccJiraIssueType_subtask (hierarchy_level=-1)
- [X] T031 [P] [US3] Implement `atlassian_jira_issue_types` data source in `internal/jira/data_source_issue_types.go`: optional project_id filter, issue_types list attribute, Read calls GET /rest/api/3/issuetype (all) or GET /rest/api/3/issuetype/project?projectId=... (filtered)
- [X] T032 [P] [US3] Write acceptance test for issue types data source in `internal/jira/data_source_issue_types_test.go`: TestAccDataSourceJiraIssueTypes_basic, TestAccDataSourceJiraIssueTypes_projectFilter
- [X] T033 [P] [US3] Create example HCL in `examples/resources/atlassian_jira_issue_type/resource.tf` and `examples/data-sources/atlassian_jira_issue_types/data-source.tf`

**Checkpoint**: Issue types can be created, updated, imported, and destroyed. Data source lists/filters them.

---

## Phase 6: User Story 4 — Manage Priorities (Priority: P4)

**Goal**: Full CRUD + import for `atlassian_jira_priority` resource (including async delete) and `atlassian_jira_priorities` data source.

**Independent Test**: Create priority, update color, import by ID, delete (async poll), verify via data source.

### Implementation for User Story 4

- [X] T034 [US4] Implement `atlassian_jira_priority` resource in `internal/jira/resource_priority.go`: schema per data-model.md (status_color hex validator, icon_url/avatar_id mutual exclusion via AtLeastOneOf/ConflictsWith), CRUD + ImportState, Delete uses async task polling via client.WaitForTask
- [X] T035 [US4] Register `atlassian_jira_priority` resource in `internal/provider/provider.go` Resources method
- [X] T036 [US4] Write unit tests in `internal/jira/resource_priority_test.go`: test color validation, mutual exclusion logic, async delete task polling with httptest mock (303 → task poll → COMPLETE)
- [X] T037 [US4] Write acceptance tests in `internal/jira/resource_priority_test.go`: TestAccJiraPriority_basic, TestAccJiraPriority_update, TestAccJiraPriority_import
- [X] T038 [P] [US4] Implement `atlassian_jira_priorities` data source in `internal/jira/data_source_priorities.go`: priorities list attribute, Read calls GET /rest/api/3/priority/search with pagination
- [X] T039 [P] [US4] Write acceptance test for priorities data source in `internal/jira/data_source_priorities_test.go`: TestAccDataSourceJiraPriorities_basic
- [X] T040 [P] [US4] Create example HCL in `examples/resources/atlassian_jira_priority/resource.tf` and `examples/data-sources/atlassian_jira_priorities/data-source.tf`

**Checkpoint**: Priorities can be created, updated, imported, async-deleted, and listed.

---

## Phase 7: User Story 5 — Manage Statuses (Priority: P5)

**Goal**: Full CRUD + import for `atlassian_jira_status` resource (bulk API adaptation) and `atlassian_jira_statuses` data source.

**Independent Test**: Create status, update name, import by ID, delete, verify via data source with project filter.

### Implementation for User Story 5

- [X] T041 [US5] Implement `atlassian_jira_status` resource in `internal/jira/resource_status.go`: schema per data-model.md (scope_type/scope_project_id ForceNew, status_category enum validator, scope_project_id required-if-PROJECT validation), CRUD + ImportState wrapping bulk API as single-item operations
- [X] T042 [US5] Register `atlassian_jira_status` resource in `internal/provider/provider.go` Resources method
- [X] T043 [US5] Write unit tests in `internal/jira/resource_status_test.go`: test bulk-to-single wrapping, scope validation, status_category enum, request/response marshaling
- [X] T044 [US5] Write acceptance tests in `internal/jira/resource_status_test.go`: TestAccJiraStatus_basic (GLOBAL scope), TestAccJiraStatus_projectScope, TestAccJiraStatus_update, TestAccJiraStatus_import
- [X] T045 [P] [US5] Implement `atlassian_jira_statuses` data source in `internal/jira/data_source_statuses.go`: optional project_id filter, statuses list attribute, Read calls GET /rest/api/3/statuses/search with pagination
- [X] T046 [P] [US5] Write acceptance test for statuses data source in `internal/jira/data_source_statuses_test.go`: TestAccDataSourceJiraStatuses_basic, TestAccDataSourceJiraStatuses_projectFilter
- [X] T047 [P] [US5] Create example HCL in `examples/resources/atlassian_jira_status/resource.tf` and `examples/data-sources/atlassian_jira_statuses/data-source.tf`

**Checkpoint**: Statuses can be created, updated, imported, and destroyed. Data source lists/filters them.

---

## Phase 8: User Story 6 — End-to-End Example & Documentation (Priority: P6)

**Goal**: Sample HCL demonstrating all resources together, plus generated registry docs.

**Independent Test**: `terraform apply` on full-setup example provisions project + issue types + priorities + statuses.

### Implementation for User Story 6

- [X] T048 [US6] Create end-to-end example in `examples/full-setup/main.tf`: provider config, jira_myself data source, project resource using myself.account_id as lead, issue type resources (standard + subtask), priority resource, status resource (GLOBAL scope), outputs
- [X] T049 [US6] Run `tfplugindocs generate` to produce Terraform Registry-compatible docs in `docs/`
- [X] T050 [US6] Verify generated docs exist for: provider index, 4 resources, 5 data sources — all with HCL examples

**Checkpoint**: Full end-to-end example works. Registry docs generated for all components.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, CI readiness, cleanup

- [X] T051 [P] Run `golangci-lint run ./...` and fix all findings
- [X] T052 [P] Run `go test -race ./...` (unit tests) and fix any race conditions
- [X] T053 Run full acceptance test suite: `TF_ACC=1 go test ./... -v -timeout 30m` and fix any failures
- [X] T054 [P] Verify all exported identifiers have doc comments per Constitution Principle I
- [X] T055 [P] Update `README.md` with provider overview, authentication setup, quick start, and development instructions

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **User Stories (Phase 3–8)**: All depend on Foundational phase completion
  - US1 (Phase 3): No dependencies on other stories
  - US2 (Phase 4): No dependencies on other stories
  - US3 (Phase 5): No dependencies on other stories
  - US4 (Phase 6): No dependencies on other stories
  - US5 (Phase 7): No dependencies on other stories
  - US6 (Phase 8): Depends on US1–US5 (needs all resources to exist)
- **Polish (Phase 9)**: Depends on all user stories being complete

### Within Each User Story

- Resource implementation before data source (resource tests may create data for data source tests)
- Registration in provider.go immediately after resource/data source implementation
- Unit tests alongside implementation
- Acceptance tests after implementation
- Example HCL can be written in parallel with tests

### Parallel Opportunities

- T002, T003, T004 in Phase 1 (independent config files)
- T011 can begin once T007–T010 are complete
- Within each user story: data source [P] tasks can run parallel to resource acceptance tests
- US2, US3, US4, US5 (Phases 4–7) can all run in parallel after Foundational phase
- T051, T052, T054, T055 in Phase 9 (independent files)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: User Story 1 (provider + myself data source)
4. **STOP and VALIDATE**: `terraform plan` with provider block + `jira_myself` succeeds
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Provider compiles and authenticates
2. Add US1 → Provider health check works (MVP!)
3. Add US2 → Projects manageable via Terraform
4. Add US3 → Issue types manageable
5. Add US4 → Priorities manageable
6. Add US5 → Statuses manageable
7. Add US6 → End-to-end example + docs
8. Polish → CI-ready

### Parallel Team Strategy

With multiple developers after Foundational phase:
- Developer A: US2 (Projects)
- Developer B: US3 (Issue Types)
- Developer C: US4 (Priorities) + US5 (Statuses)
- Then: US6 (End-to-End) once all resources exist

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable
- All API shapes from `jira-api-doc/swagger-v3.json` — never guess endpoints
- Read relevant skills before implementing: `golang-pro`, `provider-resources`, `new-terraform-provider`
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
