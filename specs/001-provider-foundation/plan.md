# Implementation Plan: Provider Foundation

**Branch**: `001-provider-foundation` | **Date**: 2026-03-01 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-provider-foundation/spec.md`

## Summary

Build the foundational Terraform provider for Atlassian Jira Cloud: a Go
project using the Terraform Plugin Framework that authenticates via
Basic Auth, wraps the Jira REST API v3 through a centralized HTTP client
with retry and pagination, and implements four managed resources
(project, issue type, priority, status) plus five data sources. All API
interactions are derived from the OpenAPI spec at
`jira-api-doc/swagger-v3.json`.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**:
  - `github.com/hashicorp/terraform-plugin-framework` — provider, resources, data sources
  - `github.com/hashicorp/terraform-plugin-testing` — acceptance test harness
  - `github.com/hashicorp/terraform-plugin-docs` — registry doc generation (tfplugindocs)
**Storage**: N/A (state managed by Terraform Core)
**Testing**: `go test` with `terraform-plugin-testing` for acceptance; `net/http/httptest` for unit mocks
**Target Platform**: Cross-platform binary (linux, darwin, windows — amd64 + arm64)
**Project Type**: Terraform provider (plugin)
**Build Tool**: GoReleaser with GPG signing
**Linting**: golangci-lint with strict config
**Performance Goals**: N/A (API-bound; backoff/retry handles throughput)
**Constraints**: All API shapes derived from `jira-api-doc/swagger-v3.json`; no inferred endpoints
**Scale/Scope**: 4 resources, 5 data sources, 1 provider configuration

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Idiomatic Go & Code Quality | PASS | Go 1.22+, golangci-lint strict, no panic, diag.Diagnostics |
| II. Test-First Discipline | PASS | Unit + acceptance tests for every resource/data source, TF_ACC gating |
| III. API Fidelity | PASS | swagger-v3.json is authoritative source; no guessing |
| IV. API Integration Discipline | PASS | Centralized client, retry/backoff, transparent pagination, drift detection |
| V. Provider Design Consistency | PASS | Full CRUD + ImportState, strongly typed schemas, sensitive marking, `atlassian_jira_<entity>` naming |
| VI. Documentation Standards | PASS | tfplugindocs, HCL examples, provider auth docs |
| VII. Multi-Product Extensibility | PASS | `internal/jira/` for product-specific; `internal/atlassian/` for shared client/auth |
| VIII. Release Quality | PASS | Semver, GoReleaser signed builds, CI gates |

No violations. Gate passed.

## Project Structure

### Documentation (this feature)

```text
specs/001-provider-foundation/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (API contracts)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
main.go                              # Provider entry point (providerserver.Serve)
go.mod
go.sum
Makefile                             # lint, test, testacc, docs, build targets
.goreleaser.yml                      # GoReleaser config
.golangci.yml                        # golangci-lint strict config

internal/
├── provider/
│   ├── provider.go                  # Provider schema, Configure, Resources, DataSources
│   └── provider_test.go             # Provider acceptance tests
│
├── atlassian/                       # Shared across all Atlassian products
│   ├── client.go                    # HTTP client: auth, retry, backoff, pagination
│   └── client_test.go               # Unit tests (httptest mocks)
│
└── jira/                            # Jira-specific resources and data sources
    ├── resource_project.go
    ├── resource_project_test.go
    ├── resource_issue_type.go
    ├── resource_issue_type_test.go
    ├── resource_priority.go
    ├── resource_priority_test.go
    ├── resource_status.go
    ├── resource_status_test.go
    ├── data_source_project.go
    ├── data_source_project_test.go
    ├── data_source_issue_types.go
    ├── data_source_issue_types_test.go
    ├── data_source_priorities.go
    ├── data_source_priorities_test.go
    ├── data_source_statuses.go
    ├── data_source_statuses_test.go
    ├── data_source_myself.go
    └── data_source_myself_test.go

templates/                           # tfplugindocs templates (optional overrides)

examples/
├── provider/
│   └── provider.tf                  # Provider configuration example
├── resources/
│   ├── atlassian_jira_project/
│   │   └── resource.tf
│   ├── atlassian_jira_issue_type/
│   │   └── resource.tf
│   ├── atlassian_jira_priority/
│   │   └── resource.tf
│   └── atlassian_jira_status/
│       └── resource.tf
├── data-sources/
│   ├── atlassian_jira_project/
│   │   └── data-source.tf
│   ├── atlassian_jira_issue_types/
│   │   └── data-source.tf
│   ├── atlassian_jira_priorities/
│   │   └── data-source.tf
│   ├── atlassian_jira_statuses/
│   │   └── data-source.tf
│   └── atlassian_jira_myself/
│       └── data-source.tf
└── full-setup/
    └── main.tf                      # End-to-end: project + issue types + priorities + statuses

docs/                                # Generated by tfplugindocs (not hand-written)
```

**Structure Decision**: Single Go module with `internal/` packages
separated by concern: `provider/` (top-level wiring), `atlassian/`
(shared HTTP client, auth, pagination), `jira/` (Jira-specific resources
and data sources). This follows Constitution Principle VII — new products
(Confluence, Bitbucket) get their own `internal/<product>/` package
without restructuring.

## Complexity Tracking

No constitution violations to justify.

## Key Design Decisions

### Provider Configuration

The provider accepts three attributes: `url`, `email`, `api_token`.
Each falls back to an environment variable (ATLASSIAN_URL,
ATLASSIAN_EMAIL, ATLASSIAN_API_TOKEN) when unset in HCL. The
`api_token` attribute is marked `Sensitive: true`. During `Configure`,
the provider calls `GET /rest/api/3/myself` to validate credentials
before constructing the shared client.

### Centralized HTTP Client (`internal/atlassian/client.go`)

- Accepts base URL + Basic Auth credentials at construction.
- Every request sets `Authorization: Basic base64(email:token)` and
  `Content-Type: application/json`.
- Retry policy: on HTTP 429 or 5xx, retry up to 5 times with
  exponential backoff (base 1s, jitter ±20%, max 30s). Respects
  `Retry-After` header from 429 responses when present.
- Pagination: a generic `Paginate[T]` helper that follows the Jira
  `startAt`/`maxResults`/`total` pattern, accumulating results until
  `isLast` is true or `startAt >= total`.
- Exposes typed methods: `Get`, `Post`, `Put`, `Delete` that
  marshal/unmarshal JSON and return `(result, error)`.

### Resource Pattern (every resource follows this)

1. Struct embeds `resource.Resource` from the Plugin Framework.
2. `Metadata` returns `atlassian_jira_<entity>`.
3. `Schema` defines typed attributes derived from swagger-v3.json.
4. `Create` calls the Jira API, then reads back the created resource
   to populate all computed fields.
5. `Read` fetches remote state, compares with Terraform state, updates
   state (drift detection).
6. `Update` sends only changed fields to the API, then reads back.
7. `Delete` calls the delete endpoint; handles 404 gracefully.
8. `ImportState` calls `Read` with the imported ID to populate state.
9. All API errors are surfaced via `resp.Diagnostics.AddError()`.

### Status API Adaptation

The Jira status API uses bulk endpoints (`POST /rest/api/3/statuses`,
`PUT /rest/api/3/statuses`, `DELETE /rest/api/3/statuses?id=...`). The
`jira_status` resource wraps these as single-item operations:
- Create: sends a `StatusCreateRequest` with a single-element `statuses`
  array.
- Update: sends a `StatusUpdateRequest` with a single-element `statuses`
  array.
- Delete: sends `DELETE` with `id` query param for the single status.
- Read: uses `GET /rest/api/3/statuses?id=<id>` to fetch a single
  status by ID.

### Priority Async Delete

`DELETE /rest/api/3/priority/{id}` returns 303 with a `Location` header
pointing to a task. The resource must:
1. Follow the redirect to get the task ID.
2. Poll `GET /rest/api/3/task/{taskId}` until status is `COMPLETE` or
   `FAILED`.
3. Return success or error based on task outcome.

### Project Template Auto-Derivation

When `project_template_key` is not specified, the provider maps
`project_type_key` to a default template:
- `software` → `com.pyxis.greenhopper.jira:gh-simplified-agility-kanban`
- `business` → `com.atlassian.jira-core-project-templates:jira-core-simplified-task-tracking`
- `service_desk` → `com.atlassian.servicedesk:simplified-it-service-management`

The user can override by setting `project_template_key` explicitly.

### Project Key Immutability

The `key` attribute uses `stringplanmodifier.RequiresReplace()` — any
change to the project key forces destroy + recreate. This was a
deliberate clarification decision (see spec Clarifications section).
