# Implementation Plan: Issues and Workflows

**Branch**: `002-issues-and-workflows` | **Date**: 2026-03-01 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-issues-and-workflows/spec.md`

## Summary

Implement 8 Terraform resources and 3 data sources for managing Jira Cloud workflows, fields, screens, and configuration schemes. All resources support full CRUD + import. The implementation builds on the provider foundation (001) reusing the existing HTTP client, provider registration patterns, and test infrastructure. An end-to-end example demonstrates the complete configuration chain.

## Technical Context

**Language/Version**: Go 1.24.1
**Primary Dependencies**: hashicorp/terraform-plugin-framework v1.18.0, hashicorp/terraform-plugin-testing v1.14.0
**Storage**: N/A (state managed by Terraform, data in Jira Cloud)
**Testing**: `go test` with `TF_ACC=1` for acceptance tests, `httptest.NewServer` for unit tests
**Target Platform**: Cross-platform Terraform provider plugin (linux, darwin, windows)
**Project Type**: Terraform provider (Go plugin)
**Performance Goals**: N/A (CLI tool, no specific latency targets)
**Constraints**: Must conform to Terraform Plugin Framework patterns; constitution mandates strongly-typed attributes (no `types.Dynamic`); all API interactions via centralized `atlassian.Client`
**Scale/Scope**: 8 resources, 3 data sources, ~18 new Go files, ~11 test files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Idiomatic Go & Code Quality | PASS | All new code follows existing patterns; `golangci-lint` required |
| II. Test-First Discipline | PASS | Unit + acceptance tests for all 12 resources/data sources |
| III. API Fidelity | PASS | All schemas from swagger-v3.json; no guessed endpoints |
| IV. API Integration Discipline | PASS | Reuses centralized `atlassian.Client`; drift detection in all Read ops |
| V. Provider Design Consistency | PASS | Full CRUD + ImportState; strongly-typed attributes; naming: `atlassian_jira_<entity>` |
| VI. Documentation Standards | PASS | tfplugindocs for all resources/data sources; HCL examples included |
| VII. Multi-Product Extensibility | PASS | All resources in `internal/jira/`; shared client in `internal/atlassian/` |
| VIII. Release Quality | PASS | Atomic commits; quality gates enforced |

**Post-design re-check**: `custom_fields` uses `map(string)` not `types.Dynamic` — PASS. Workflow rule parameters use `map(string)` — PASS. No constitution violations.

## Project Structure

### Documentation (this feature)

```text
specs/002-issues-and-workflows/
├── plan.md              # This file
├── research.md          # Phase 0 output (API research, decision log)
├── data-model.md        # Phase 1 output (entity definitions)
├── quickstart.md        # Phase 1 output (getting started guide)
├── contracts/           # Phase 1 output (Terraform schema contracts)
│   └── terraform-schemas.md
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
internal/jira/
├── resource_field.go                       # Custom field resource
├── resource_field_configuration.go         # Field configuration resource
├── resource_field_configuration_scheme.go  # Field configuration scheme resource
├── resource_screen.go                      # Screen resource (composite: tabs + fields)
├── resource_screen_scheme.go               # Screen scheme resource
├── resource_issue_type_screen_scheme.go    # Issue type screen scheme resource
├── resource_workflow.go                    # Workflow resource (statuses + transitions + rules)
├── resource_workflow_scheme.go             # Workflow scheme resource (with draft management)
├── data_source_fields.go                   # Fields data source
├── data_source_workflows.go               # Workflows data source
├── data_source_screens.go                  # Screens data source
├── resource_field_test.go                  # Unit + acceptance tests
├── resource_field_configuration_test.go
├── resource_field_configuration_scheme_test.go
├── resource_screen_test.go
├── resource_screen_scheme_test.go
├── resource_issue_type_screen_scheme_test.go
├── resource_workflow_test.go
├── resource_workflow_scheme_test.go
├── data_source_fields_test.go
├── data_source_workflows_test.go
└── data_source_screens_test.go

internal/provider/
└── provider.go                             # Updated: register 9 new resources + 3 data sources

examples/002-issues-and-workflows/
└── main.tf                                 # End-to-end example

docs/resources/
├── jira_field.md
├── jira_field_configuration.md
├── jira_field_configuration_scheme.md
├── jira_screen.md
├── jira_screen_scheme.md
├── jira_issue_type_screen_scheme.md
├── jira_workflow.md
└── jira_workflow_scheme.md

docs/data-sources/
├── jira_fields.md
├── jira_workflows.md
└── jira_screens.md
```

**Structure Decision**: Follows the established pattern from 001 — all Jira resources in `internal/jira/`, one file per resource/data source, tests alongside source. No new packages needed.

## Implementation Phases

### Phase 1: Custom Fields (P1)

**Resources**: `jira_field`, `jira_fields` data source

**API Endpoints**:
- `POST /rest/api/3/field` — create custom field
- `GET /rest/api/3/field` — list all fields (for data source)
- `GET /rest/api/3/field/search` — search fields (paginated, for reading individual fields)
- `PUT /rest/api/3/field/{fieldId}` — update custom field
- `DELETE /rest/api/3/field/{fieldId}` — trash custom field

**Key design decisions**:
- `type` is ForceNew (immutable after creation)
- Delete uses trash mechanism
- Data source supports `type` filter: "system" or "custom"
- Import by field ID (e.g., `customfield_10001`)

**Files**: `resource_field.go`, `resource_field_test.go`, `data_source_fields.go`, `data_source_fields_test.go`

### Phase 2: Field Configurations (P2)

**Resources**: `jira_field_configuration`, `jira_field_configuration_scheme`

**API Endpoints**:
- Field Configuration: `POST/GET/PUT/DELETE /rest/api/3/fieldconfiguration`
- Field Items: `GET/PUT /rest/api/3/fieldconfiguration/{id}/fields` (PUT replaces all)
- Scheme: `POST/GET/PUT/DELETE /rest/api/3/fieldconfigurationscheme`
- Mappings: `GET /rest/api/3/fieldconfigurationscheme/mapping`, `PUT /rest/api/3/fieldconfigurationscheme/{id}/mapping`, `POST /rest/api/3/fieldconfigurationscheme/{id}/mapping/delete`

**Key design decisions**:
- Field items use PUT replace-all semantics (no diff needed)
- Scheme mappings support "default" issue type ID for fallback
- Both support import by numeric ID

**Files**: `resource_field_configuration.go`, `resource_field_configuration_test.go`, `resource_field_configuration_scheme.go`, `resource_field_configuration_scheme_test.go`

### Phase 3: Screens (P3)

**Resources**: `jira_screen`, `jira_screen_scheme`, `jira_issue_type_screen_scheme`, `jira_screens` data source

**API Endpoints**:
- Screen: `POST/GET/PUT/DELETE /rest/api/3/screens`
- Tabs: `POST/GET/PUT/DELETE /rest/api/3/screens/{screenId}/tabs`, move: `POST .../tabs/{tabId}/move/{pos}`
- Fields: `POST/GET/DELETE /rest/api/3/screens/{screenId}/tabs/{tabId}/fields`, move: `POST .../fields/{id}/move`
- Screen Scheme: `POST/GET/PUT/DELETE /rest/api/3/screenscheme`
- ITSS: `POST/GET/PUT/DELETE /rest/api/3/issuetypescreenscheme`

**Key design decisions**:
- Screen is a composite resource: create screen → create tabs → add fields → reorder
- Update performs diff: compare current vs desired tabs/fields, add/remove/reorder as needed
- Tab `id` is computed (assigned by Jira API)
- Field ordering uses the `move` endpoint with `First`/`Last` positioning
- Tab ordering uses `move/{pos}` with 0-based index
- Screen scheme maps operations to screen IDs: `default` (required), `create`, `edit`, `view` (optional)
- ITSS mappings support "default" issue type for fallback

**Files**: `resource_screen.go`, `resource_screen_test.go`, `resource_screen_scheme.go`, `resource_screen_scheme_test.go`, `resource_issue_type_screen_scheme.go`, `resource_issue_type_screen_scheme_test.go`, `data_source_screens.go`, `data_source_screens_test.go`

### Phase 4: Workflows (P4)

**Resources**: `jira_workflow`, `jira_workflow_scheme`, `jira_workflows` data source

**API Endpoints**:
- Create: `POST /rest/api/3/workflows/create` (bulk, creates statuses + workflow atomically)
- Read: `POST /rest/api/3/workflows` (bulk read by name/ID)
- Update: `POST /rest/api/3/workflows/update` (bulk update)
- Delete: `DELETE /rest/api/3/workflow/{entityId}` (single, inactive only)
- Scheme: `POST/GET/PUT/DELETE /rest/api/3/workflowscheme`
- Draft: `POST .../createdraft`, `GET/PUT/DELETE .../draft`, `POST .../draft/publish`
- Project Usages: `GET /rest/api/3/workflowscheme/{id}/projectUsages`

**Key design decisions**:
- Workflow uses bulk create/update API (wraps single workflow in array)
- Statuses use `status_reference` (local UUID) for internal cross-referencing within transitions
- `status_id` is computed after creation (resolved by Jira)
- Transitions support full rule model: conditions (hierarchical AND/OR), validators, post-functions
- Rules use `rule_key` + `parameters` map(string) pattern for all 21+ rule types
- Workflow scheme detects active status via project usages endpoint
- Active scheme updates: create draft → update draft → publish draft (async via task polling)
- Data source supports project filtering

**Files**: `resource_workflow.go`, `resource_workflow_test.go`, `resource_workflow_scheme.go`, `resource_workflow_scheme_test.go`, `data_source_workflows.go`, `data_source_workflows_test.go`

### ~~Phase 5: Issues (P5)~~ REMOVED

**REMOVED**: The `jira_issue` resource has been removed. Managing individual Jira issues from Terraform is not a good use case — issues are ephemeral work items, not infrastructure.

### Phase 6: Integration & Documentation (P6)

**Deliverables**:
- Register all 8 resources + 3 data sources in `provider.go`
- End-to-end example in `examples/002-issues-and-workflows/main.tf`
- Terraform Registry docs for all resources and data sources
- Acceptance tests for the full chain (apply → plan-no-changes → destroy)

**Files**: `internal/provider/provider.go` (updated), `examples/002-issues-and-workflows/main.tf`, `docs/resources/*.md`, `docs/data-sources/*.md`

## Risk Register

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Workflow bulk API has undocumented behaviors | Medium | High | Test each rule type individually; fall back to swagger spec |
| Screen composite operations partially fail (e.g., tab created but fields fail) | Medium | Medium | Implement Read to recover current state; re-apply on next plan/apply |
| Workflow scheme draft publish is slow (async) | Low | Medium | Use existing `WaitForTask` with 5-minute timeout |
| Custom field delete is not truly permanent (trash) | Low | Low | Document behavior; treat trash as deleted in provider |
| ADF format changes or has edge cases | Low | Low | Store as opaque JSON string; let Jira validate |

## Complexity Tracking

No constitution violations requiring justification.
