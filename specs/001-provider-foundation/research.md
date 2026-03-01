# Research: Provider Foundation

**Phase**: 0 — Outline & Research
**Date**: 2026-03-01

## R-001: Terraform Plugin Framework Provider Pattern

**Decision**: Use `terraform-plugin-framework` (not SDKv2) for all
resources and data sources.

**Rationale**: The Plugin Framework is HashiCorp's recommended path
forward. It provides strongly typed schemas, native plan modifiers,
built-in validators, and first-class support for `ImportState`. SDKv2 is
in maintenance mode.

**Alternatives considered**:
- SDKv2: More examples available but deprecated for new providers.
- terraform-plugin-mux: Not needed — we're a pure Framework provider.

## R-002: Jira Cloud REST API v3 Authentication

**Decision**: HTTP Basic Auth with `email:api_token` encoded as Base64
in the `Authorization` header.

**Rationale**: Jira Cloud REST API v3 supports Basic Auth with API
tokens (not passwords). This is the simplest and most widely documented
auth method for Terraform providers. OAuth 2.0 (3LO) is available but
adds complexity (token refresh, browser redirect) that is inappropriate
for headless automation.

**Alternatives considered**:
- OAuth 2.0 (3LO): Requires interactive browser flow; not suitable for
  CI/CD. Can be added later as an optional auth method.
- Personal Access Tokens: Not supported by Jira Cloud (only Jira
  Data Center).

## R-003: HTTP Client Retry Strategy

**Decision**: Exponential backoff with jitter on HTTP 429 and 5xx. Up
to 5 retries. Base interval 1 second, max 30 seconds. Respect
`Retry-After` header when present.

**Rationale**: Jira Cloud enforces rate limits and returns 429 with an
optional `Retry-After` header. Exponential backoff with jitter is the
standard approach to avoid thundering herd. Five retries with max 30s
provides adequate coverage for transient issues without hanging
indefinitely.

**Alternatives considered**:
- Fixed delay: Simple but wastes time on short blips and doesn't
  prevent thundering herd.
- hashicorp/go-retryablehttp: Adds a dependency for what is ~50 lines
  of custom code. Our client needs Jira-specific pagination and auth
  integration anyway, so a standalone client is cleaner.

## R-004: Pagination Pattern

**Decision**: Generic `Paginate[T]` function that follows Jira's
`startAt`/`maxResults`/`total` pattern.

**Rationale**: Jira Cloud paginated endpoints return a common envelope:
`{ startAt, maxResults, total, isLast, values: [...] }`. A generic
function can iterate until `isLast == true` or `startAt >= total`,
accumulating all `values` into a single slice. This keeps resource code
clean — callers just get the full result set.

**Alternatives considered**:
- Per-resource pagination: Duplicates logic across every list operation.
- Iterator pattern: More idiomatic in some languages but Go's
  generics make a simple accumulator function equally clean and easier
  to test.

## R-005: Project Template Key Mapping

**Decision**: Auto-derive a default `projectTemplateKey` from
`projectTypeKey`; expose optional override.

**Rationale**: The Jira API requires `projectTemplateKey` at project
creation time, and it must be compatible with the `projectTypeKey`.
Requiring users to know the full template key string
(`com.pyxis.greenhopper.jira:gh-simplified-agility-kanban`) would be
terrible UX. A sensible default per type keeps the schema simple.

**Default mappings** (derived from Jira Cloud documentation):
| projectTypeKey | Default projectTemplateKey |
|----------------|----------------------------|
| `software` | `com.pyxis.greenhopper.jira:gh-simplified-agility-kanban` |
| `business` | `com.atlassian.jira-core-project-templates:jira-core-simplified-task-tracking` |
| `service_desk` | `com.atlassian.servicedesk:simplified-it-service-management` |

**Alternatives considered**:
- Require explicit template key: Bad UX; most users don't know these.
- Hard-code without override: Inflexible for teams with custom
  templates.

## R-006: Status API Bulk-to-Single Adaptation

**Decision**: Wrap the bulk status API as single-item Terraform CRUD
operations.

**Rationale**: The Jira v3 status API only provides bulk endpoints
(`POST /rest/api/3/statuses` creates multiple, `PUT` updates multiple,
`DELETE` deletes multiple by query param). Terraform resources operate
on individual items. The adapter sends arrays with exactly one element.
This is a clean abstraction that preserves the Terraform resource model.

**Alternatives considered**:
- Batch resource: A `jira_statuses` (plural) resource that manages
  multiple statuses. This breaks the standard Terraform resource model
  and complicates state management.

## R-007: Priority Async Delete Handling

**Decision**: Follow the 303 redirect, poll the task endpoint until
completion or failure.

**Rationale**: `DELETE /rest/api/3/priority/{id}` returns 303 with a
`Location` header pointing to `/rest/api/3/task/{taskId}`. The provider
must poll this task until it reaches status `COMPLETE` or `FAILED`. A
helper function `waitForTask` with configurable timeout (default 2
minutes) handles this pattern. The same helper can be reused if other
endpoints adopt async patterns in the future.

**Alternatives considered**:
- Fire-and-forget: Terraform would report success before the delete
  completes, leading to stale state.

## R-008: Testing Strategy

**Decision**: Two-tier testing — unit tests with `httptest` mocks and
acceptance tests with `terraform-plugin-testing` against a live Jira
Cloud instance.

**Rationale**:
- Unit tests validate request construction, response parsing, error
  handling, retry logic, and pagination without network calls.
- Acceptance tests validate end-to-end CRUD + import against a real
  Jira instance, gated by `TF_ACC=1`.
- This aligns with Constitution Principle II and standard Terraform
  provider testing practices.

**Alternatives considered**:
- Recorded HTTP fixtures (go-vcr): Adds complexity and fixture
  maintenance burden. For a new provider, live acceptance tests are
  more reliable and simpler to start with.

## R-009: Documentation Generation

**Decision**: Use `tfplugindocs` with example files in `examples/`
directory following the standard tfplugindocs layout.

**Rationale**: `tfplugindocs` generates Terraform Registry-compatible
documentation from schema definitions and example files. The standard
directory layout (`examples/resources/<name>/resource.tf`,
`examples/data-sources/<name>/data-source.tf`) is auto-discovered.

**Alternatives considered**:
- Hand-written docs: Error-prone and falls out of sync with schema.

## R-010: Build and Release

**Decision**: GoReleaser for cross-platform builds with GPG signing.
Makefile for local development workflows.

**Rationale**: GoReleaser is the standard tool for Terraform provider
releases. It handles cross-compilation, checksums, GPG signing, and
GitHub release creation. The Makefile provides local shortcuts (lint,
test, testacc, docs, build).

**Alternatives considered**:
- Plain `go build` scripts: Lacks checksum manifests and signing
  needed for Terraform Registry publishing.
