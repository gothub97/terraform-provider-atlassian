# Tasks: Issues and Workflows

**Input**: Design documents from `/specs/002-issues-and-workflows/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Required by constitution (Principle II) and spec (FR-052). Every resource and data source must have unit + acceptance tests.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

**Update (2026-03-01)**: US5 (Manage Issues) REMOVED — managing individual Jira issues from Terraform is not a good use case. Issues are ephemeral work items, not infrastructure. The `atlassian_jira_issue_type` resource is the correct abstraction.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Verify the existing foundation and prepare for new resources

- [X] T001 Verify existing test infrastructure runs against real Jira Cloud instance: `export PATH="/home/gauthier/.local/go/bin:$HOME/go/bin:$PATH" && source ~/.bashrc && TF_ACC=1 go test -v -count=1 -timeout 10m -run TestAcc ./internal/jira/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: No new foundational infrastructure needed — 001 provides HTTP client, provider, test helpers. This phase is a no-op.

**Checkpoint**: Foundation ready from 001 — user story implementation can begin.

---

## Phase 3: User Story 1 - Manage Custom Fields (Priority: P1)

**Goal**: Users can create, read, update, delete, and import custom fields. The `jira_fields` data source lists all fields with filtering.

**Independent Test**: Create a custom text field, update its name, import it, delete it. Query the fields data source with and without type filter.

### Tests for User Story 1

- [X] T004 [P] [US1] Write acceptance tests for jira_field resource in `internal/jira/resource_field_acc_test.go`. Tests: `TestAccJiraField_basic` (create text field, check attributes), `TestAccJiraField_update` (change name/description), `TestAccJiraField_import` (import state verify). Use `acctest.RandStringFromCharSet` for unique names.
- [X] T005 [P] [US1] Write acceptance tests for jira_fields data source in `internal/jira/data_source_fields_acc_test.go`. Tests: `TestAccJiraFields_all` (verify system fields returned), `TestAccJiraFields_customFilter` (create a field, filter by custom, verify it appears).

### Implementation for User Story 1

- [X] T006 [US1] Implement `jira_field` resource in `internal/jira/resource_field.go`. Define: `FieldResource` struct, `FieldResourceModel` (id, name, type, description, searcher_key), API request/response types. Schema: `type` has `RequiresReplace()`, `id` is Computed. CRUD: POST create, GET via field search, PUT update name/description/searcher_key, DELETE trash. Import via `ImportStatePassthroughID`. Drift detection: remove from state on 404.
- [X] T007 [P] [US1] Implement `jira_fields` data source in `internal/jira/data_source_fields.go`. Define: `FieldsDataSource` struct, `FieldsDataSourceModel` with optional `type` filter attribute and `fields` list output. Read: call `GET /rest/api/3/field`, filter by `custom` boolean based on type attribute. Output: id, name, custom (bool), schema type, clause_names.
- [X] T008 [US1] Register `jira.NewFieldResource` and `jira.NewFieldsDataSource` in `internal/provider/provider.go` Resources() and DataSources() methods.
- [X] T009 [US1] Run all US1 tests and verify they pass: `TF_ACC=1 go test -v -count=1 -timeout 15m -run "TestAccJiraField" ./internal/jira/`

**Checkpoint**: Custom fields fully functional — create, update, import, delete, list with filtering.

---

## Phase 4: User Story 2 - Configure Field Behavior (Priority: P2)

**Goal**: Users can create field configurations with field items, and field configuration schemes mapping issue types to field configurations.

**Independent Test**: Create a field configuration with required/hidden field items, update items, create a scheme with default and issue-type mappings, import both.

### Tests for User Story 2

- [X] T012 [P] [US2] Write acceptance tests for both resources in their respective test files. Tests: `TestAccJiraFieldConfiguration_basic`, `TestAccJiraFieldConfiguration_update`, `TestAccJiraFieldConfiguration_import`, `TestAccJiraFieldConfigurationScheme_basic`, `TestAccJiraFieldConfigurationScheme_withMappings`, `TestAccJiraFieldConfigurationScheme_import`.

### Implementation for User Story 2

- [X] T013 [US2] Implement `jira_field_configuration` resource in `internal/jira/resource_field_configuration.go`. Define model with id, name, description, field_items (ListNestedBlock with field_id, is_required, is_hidden, description, renderer). Create: POST config then PUT field items. Read: GET config + GET items. Update: PUT config + PUT items (replace-all). Delete: DELETE. Import by numeric ID.
- [X] T014 [US2] Implement `jira_field_configuration_scheme` resource in `internal/jira/resource_field_configuration_scheme.go`. Define model with id, name, description, mappings (ListNestedBlock with issue_type_id, field_configuration_id). Create: POST scheme + PUT mappings. Read: GET scheme + GET mappings (paginated via `GET /rest/api/3/fieldconfigurationscheme/mapping?fieldConfigurationSchemeId=`). Update: PUT scheme + PUT mappings. Delete: DELETE. Import by ID.
- [X] T015 [US2] Register `jira.NewFieldConfigurationResource` and `jira.NewFieldConfigurationSchemeResource` in `internal/provider/provider.go`.
- [X] T016 [US2] Run all US2 tests: `TF_ACC=1 go test -v -count=1 -timeout 15m -run "TestAccJiraFieldConfig" ./internal/jira/`

**Checkpoint**: Field configurations and schemes fully functional.

---

## Phase 5: User Story 3 - Design Screens (Priority: P3)

**Goal**: Users can create screens with tabs and ordered fields, screen schemes mapping screens to operations, issue type screen schemes mapping issue types to screen schemes, and query the screens data source.

**Independent Test**: Create a screen with two tabs and ordered fields, update field order, create a screen scheme, create an ITSS with default mapping, import all three, query screens data source.

### Tests for User Story 3

- [X] T021 [P] [US3] Write acceptance tests for all three resources and data source. Tests: `TestAccJiraScreen_basic` (create with one tab + fields), `TestAccJiraScreen_multipleTabs` (two tabs, verify ordering), `TestAccJiraScreen_update` (add field, remove field, reorder), `TestAccJiraScreen_import`, `TestAccJiraScreenScheme_basic`, `TestAccJiraScreenScheme_import`, `TestAccJiraIssueTypeScreenScheme_basic`, `TestAccJiraIssueTypeScreenScheme_import`, `TestAccJiraScreens_all`.

### Implementation for User Story 3

- [X] T022 [US3] Implement `jira_screen` resource in `internal/jira/resource_screen.go`. Define model with id, name, description, tabs (ListNestedBlock with id (computed), name, fields list(string)). Create: POST screen → for each tab: POST tab → for each field: POST add field → reorder fields. Read: GET screen + GET tabs + GET fields per tab. Update: diff current vs desired tabs (add new, remove old, rename changed) + diff fields per tab (add/remove/move). Delete: DELETE screen. Import by numeric ID.
- [X] T023 [P] [US3] Implement `jira_screen_scheme` resource in `internal/jira/resource_screen_scheme.go`. Define model with id, name, description, default_screen_id (required), create_screen_id, edit_screen_id, view_screen_id (optional). CRUD maps to `POST/GET/PUT/DELETE /rest/api/3/screenscheme`. Screens map sent as `{default: id, create: id, edit: id, view: id}`. Import by ID.
- [X] T024 [P] [US3] Implement `jira_issue_type_screen_scheme` resource in `internal/jira/resource_issue_type_screen_scheme.go`. Define model with id, name, description, mappings (issue_type_id + screen_scheme_id). Create: POST scheme. Read: GET scheme + GET mappings. Update: PUT scheme + PUT mappings. Delete: DELETE. Import by ID.
- [X] T025 [P] [US3] Implement `jira_screens` data source in `internal/jira/data_source_screens.go`. List all screens via `GET /rest/api/3/screens` with pagination. Output: screens list with id, name, description.
- [X] T026 [US3] Register `NewScreenResource`, `NewScreenSchemeResource`, `NewIssueTypeScreenSchemeResource`, `NewScreensDataSource` in `internal/provider/provider.go`.
- [X] T027 [US3] Run all US3 tests: `TF_ACC=1 go test -v -count=1 -timeout 15m -run "TestAccJiraScreen|TestAccJiraIssueTypeScreenScheme" ./internal/jira/`

**Checkpoint**: Screens, screen schemes, and issue type screen schemes fully functional.

---

## Phase 6: User Story 4 - Define Workflows (Priority: P4)

**Goal**: Users can create workflows with statuses, transitions (including conditions/validators/post-functions), and workflow schemes with draft management. The workflows data source lists workflows with project filtering.

**Independent Test**: Create a workflow with 3 statuses and 3 transitions (including one with a validator), create a workflow scheme, import both, query workflows data source.

### Tests for User Story 4

- [X] T031 [P] [US4] Write acceptance tests for all resources and data source. Tests: `TestAccJiraWorkflow_basic` (3 statuses, 3 transitions), `TestAccJiraWorkflow_withRules` (transitions with validators/conditions/post-functions), `TestAccJiraWorkflow_update` (add transition), `TestAccJiraWorkflow_import`, `TestAccJiraWorkflowScheme_basic`, `TestAccJiraWorkflowScheme_withMappings`, `TestAccJiraWorkflowScheme_import`, `TestAccJiraWorkflows_all`.

### Implementation for User Story 4

- [X] T032 [US4] Implement `jira_workflow` resource in `internal/jira/resource_workflow.go`. Define model with id, name, description, statuses (ListNestedBlock: status_reference, status_id (computed), name, status_category), transitions (ListNestedBlock: name, from_status_reference, to_status_reference, type, validators (ListNestedBlock: rule_key, parameters map(string)), condition (SingleNestedBlock: operator, rules list, groups list — recursive structure), post_functions (ListNestedBlock: rule_key, parameters map(string))). Create: POST `/workflows/create` wrapping single workflow in array. Read: POST `/workflows` with workflow ID. Update: POST `/workflows/update`. Delete: DELETE `/workflow/{entityId}`. Import by UUID.
- [X] T033 [P] [US4] Implement `jira_workflow_scheme` resource in `internal/jira/resource_workflow_scheme.go`. Define model with id, name, description, default_workflow, mappings (issue_type_id + workflow_name). Create: POST scheme. Read: GET scheme. Update: check project usages — if active, create draft + update draft + publish (async via `WaitForTask`); if inactive, PUT directly. Delete: DELETE. Import by numeric ID.
- [X] T034 [P] [US4] Implement `jira_workflows` data source in `internal/jira/data_source_workflows.go`. Read: POST `/rest/api/3/workflows` with optional project/issue type filter from `project_key` attribute. Output: workflows list with id, name, description, statuses, transitions.
- [X] T035 [US4] Register `NewWorkflowResource`, `NewWorkflowSchemeResource`, `NewWorkflowsDataSource` in `internal/provider/provider.go`.
- [X] T036 [US4] Run all US4 tests: `TF_ACC=1 go test -v -count=1 -timeout 15m -run "TestAccJiraWorkflow" ./internal/jira/`

**Checkpoint**: Workflows and workflow schemes fully functional, including transition rules and draft management.

---

## ~~Phase 7: User Story 5 - Manage Issues (Priority: P5)~~ REMOVED

**REMOVED**: The `atlassian_jira_issue` resource has been removed. Managing individual Jira issues from Terraform is not a good use case — issues are ephemeral work items, not infrastructure. The `atlassian_jira_issue_type` resource is the correct abstraction.

- ~~T037-T041~~ REMOVED

---

## Phase 8: User Story 6 - End-to-End Project Configuration (Priority: P6)

**Goal**: A single Terraform configuration provisions the complete chain: project → custom fields → field configuration → screen → screen scheme → issue type screen scheme → workflow → workflow scheme. Validates idempotency and clean destroy.

**Independent Test**: Run `terraform apply` on the end-to-end example, verify `terraform plan` shows no changes, run `terraform destroy` successfully.

### Implementation for User Story 6

- [X] T042 [US6] Create end-to-end example in `examples/002-issues-and-workflows/main.tf`. Declare: provider config, project (from 001), custom field, field configuration with field items, field configuration scheme with default mapping, screen with tabs and fields, screen scheme mapping screen to operations, issue type screen scheme with default, workflow with statuses/transitions, workflow scheme with issue type mapping. Use proper `depends_on` or reference chaining for dependency ordering.
- [ ] T043 [US6] Write end-to-end acceptance test in `internal/jira/e2e_test.go`. Test: `TestAccE2E_fullChain` — apply the full chain config, verify all resources created, run plan (expect no changes), import each resource and verify no diff, destroy all resources. Use unique random names for all resources.
- [ ] T044 [US6] Run the end-to-end test: `TF_ACC=1 go test -v -count=1 -timeout 30m -run TestAccE2E ./internal/jira/`

**Checkpoint**: Full configuration chain works end-to-end: apply → plan-no-changes → destroy.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, linting, and final quality verification

- [X] T045 [P] Generate Terraform Registry documentation for all 8 resources in `docs/resources/`: `jira_field.md`, `jira_field_configuration.md`, `jira_field_configuration_scheme.md`, `jira_screen.md`, `jira_screen_scheme.md`, `jira_issue_type_screen_scheme.md`, `jira_workflow.md`, `jira_workflow_scheme.md`. Each doc must include description, example HCL (from contracts/terraform-schemas.md), and full attribute reference.
- [X] T046 [P] Generate Terraform Registry documentation for all 3 data sources in `docs/data-sources/`: `jira_fields.md`, `jira_workflows.md`, `jira_screens.md`. Each doc must include description, example HCL, and attribute reference.
- [ ] T047 Run `go vet ./...` and fix any findings across all new files in `internal/jira/`.
- [ ] T048 Run full test suite: `TF_ACC=1 go test -v -count=1 -timeout 30m ./internal/jira/` and confirm all tests pass.
- [ ] T049 Final verification: `go build ./...` clean build with no warnings.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — verify existing foundation ✅
- **US1 (Phase 3)**: Depends on Setup — complete ✅
- **US2 (Phase 4)**: Depends on Setup — complete ✅
- **US3 (Phase 5)**: Depends on Setup — complete ✅
- **US4 (Phase 6)**: Depends on Setup — complete ✅
- ~~**US5 (Phase 7)**: REMOVED~~
- **US6 (Phase 8)**: Depends on US1-US4 being complete — E2E test pending
- **Polish (Phase 9)**: Depends on US6 — verification pending

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Resource implementation before data source implementation
- Core CRUD before advanced features (e.g., screen tab diff logic)
- Register in provider.go after implementation
- Run story tests to verify

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing (TDD per constitution Principle II)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- The most complex resources are: `jira_screen` (composite tab/field operations) and `jira_workflow` (transition rules with hierarchical conditions)
- All acceptance tests require: `export PATH="/home/gauthier/.local/go/bin:$HOME/go/bin:$PATH" && source ~/.bashrc && TF_ACC=1`
