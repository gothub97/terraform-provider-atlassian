<!--
  Sync Impact Report
  ==================
  Version change: N/A → 1.0.0 (MAJOR — initial adoption)

  Added principles:
    I.   Idiomatic Go & Code Quality
    II.  Test-First Discipline
    III. API Fidelity
    IV.  API Integration Discipline
    V.   Provider Design Consistency
    VI.  Documentation Standards
    VII. Multi-Product Extensibility
    VIII. Release Quality

  Added sections:
    - Skill-Driven Development
    - Development Workflow & Quality Gates
    - Governance

  Removed sections: none (initial version)

  Templates requiring updates:
    - .specify/templates/plan-template.md        ✅ compatible (Constitution Check section generic)
    - .specify/templates/spec-template.md         ✅ compatible (no constitution-specific references)
    - .specify/templates/tasks-template.md        ✅ compatible (phases align with testing/polish principles)
    - .specify/templates/commands/               ✅ no command files exist

  Follow-up TODOs: none
-->

# Atlassian Terraform Provider Constitution

## Core Principles

### I. Idiomatic Go & Code Quality

- All code MUST be idiomatic Go — follow Effective Go, the Go Code
  Review Comments wiki, and `go vet`/`staticcheck` without exceptions.
- No `panic()` in production code. All errors MUST be surfaced through
  Terraform diagnostics (`diag.Diagnostics`), never swallowed or logged
  silently.
- Exported identifiers MUST have doc comments. Internal helpers SHOULD
  have comments when the intent is non-obvious.
- `golangci-lint` MUST pass with zero findings on every commit. The
  linter configuration is authoritative; do not disable rules inline
  without a justifying comment and reviewer approval.

### II. Test-First Discipline

- Every resource and data source MUST have both unit tests and
  acceptance tests before it is considered complete.
- Acceptance tests MUST run against a real Atlassian Cloud instance
  (not mocked). They are gated by the `TF_ACC` environment variable.
- All tests MUST be executed against the real environment (`TF_ACC=1`)
  as the default validation step. Mock-based unit tests are
  supplementary; they do NOT replace real-environment validation.
- Unit tests MUST cover schema validation, plan-time logic, and any
  helper/utility functions in isolation.
- Test names MUST follow the convention `TestAcc<Resource>_<scenario>`
  for acceptance tests and `Test<Unit>_<scenario>` for unit tests.
- Flaky tests MUST be fixed or quarantined immediately — never ignored.

### III. API Fidelity

- The OpenAPI specification at `jira-api-doc/swagger-v3.json` is the
  single source of truth for every Jira API interaction.
- All endpoint paths, HTTP methods, request bodies, query parameters,
  and response schemas MUST match exactly what the spec defines.
- Never guess, infer, or hallucinate API fields, endpoints, or
  behaviors. When in doubt, read the spec.
- When the upstream spec changes, affected resources MUST be updated to
  match before the next release.

### IV. API Integration Discipline

- A single, centralized HTTP client MUST be used for all Atlassian API
  calls. No ad-hoc `http.Client` usage.
- The client MUST implement retry with exponential backoff and jitter
  for rate-limit responses (HTTP 429) and transient server errors
  (5xx).
- Paginated endpoints MUST be consumed transparently — callers receive
  complete result sets, never partial pages.
- Every `Read` operation MUST perform drift detection by comparing
  remote state against Terraform state and updating the state file
  accordingly.

### V. Provider Design Consistency

- Every managed resource MUST implement full CRUD (Create, Read,
  Update, Delete) plus `ImportState`.
- Schemas MUST use strongly typed attributes (`types.String`,
  `types.Int64`, `types.Bool`, etc.) — never `types.Dynamic` or
  untyped maps as a shortcut.
- Sensitive fields (API tokens, secrets) MUST be marked `Sensitive:
  true` in the schema and MUST NOT appear in logs or plan output.
- Computed fields MUST be marked `Computed: true` and MUST be
  populated during `Read`.
- Resource and data source type names MUST follow the pattern
  `atlassian_jira_<entity>` (e.g., `atlassian_jira_project`).

### VI. Documentation Standards

- Every resource and data source MUST have Terraform Registry-
  compatible documentation generated with `tfplugindocs`.
- Each doc page MUST include at least one complete HCL usage example
  that is valid and copy-pasteable.
- Schema descriptions MUST be user-facing quality — concise, accurate,
  and referencing Jira terminology where appropriate.
- The provider-level documentation MUST explain authentication setup,
  required permissions, and rate-limit considerations.

### VII. Multi-Product Extensibility

- The package architecture MUST support adding Confluence, Bitbucket,
  and other Atlassian products without restructuring existing code.
- Product-specific logic MUST live in dedicated packages (e.g.,
  `internal/jira/`, `internal/confluence/`). Shared concerns (HTTP
  client, auth, pagination) MUST live in a common package.
- The provider configuration schema MUST allow product-specific
  options without breaking changes when new products are added.
- Resource type naming (`atlassian_<product>_<entity>`) MUST
  namespace by product to prevent collisions.

### VIII. Release Quality

- Releases MUST follow semantic versioning (semver). Breaking changes
  to resources or provider configuration MUST increment the major
  version.
- Release binaries MUST be signed and published through a CI pipeline
  — no manual builds.
- The CI pipeline MUST enforce all quality gates before release:
  linting, unit tests, acceptance tests, and documentation generation.
- Changelog entries MUST accompany every release and follow the
  Keep a Changelog format.

## Skill-Driven Development

- The project maintains custom skills at `.claude/skills/`. Before
  performing any task, the relevant skill(s) MUST be read and
  followed.
- Available skills and their triggers:
  - **golang-pro**: All Go code generation and review.
  - **new-terraform-provider**: Initializing the provider or adding
    new top-level components.
  - **provider-resources**: Every time a resource or data source is
    created, modified, or reviewed.
  - **provider-actions**: When implementing imperative operations
    at lifecycle events.
  - **refactor-module**: Restructuring packages, extracting shared
    logic, or cleaning up code.
  - **terraform-stacks**: Designing how resources compose together.
  - **speckit-analyze**: During `/speckit.analyze` phases.
  - **speckit-constitution**: During constitution creation.
  - **speckit-specify**: During `/speckit.specify` phases.
- When multiple skills apply (e.g., creating a new resource involves
  both `golang-pro` and `provider-resources`), ALL applicable skills
  MUST be read and followed. Never skip a skill because the task
  seems simple.

## Development Workflow & Quality Gates

- Every change MUST pass the following gates before merge:
  1. `golangci-lint run` — zero findings.
  2. `go test ./...` — all unit tests pass.
  3. `TF_ACC=1 go test ./...` — all acceptance tests pass (CI only).
  4. `tfplugindocs generate` — docs regenerated without errors.
- Commits MUST be atomic and focused. One logical change per commit.
- Pull requests MUST include a description of what changed, why, and
  how it was tested.
- The `main` branch MUST always be in a releasable state.

## Governance

- This constitution is the supreme governing document for the
  Atlassian Terraform Provider project. It supersedes all other
  practices, conventions, or ad-hoc decisions.
- Amendments require: (1) a written proposal describing the change
  and rationale, (2) update to this document with version increment,
  (3) propagation of changes to all dependent templates and
  workflows.
- Version increments follow semver:
  - MAJOR: Principle removed or fundamentally redefined.
  - MINOR: New principle or section added, or existing guidance
    materially expanded.
  - PATCH: Clarifications, wording fixes, non-semantic refinements.
- Compliance review: every PR and code review MUST verify adherence
  to these principles. Deviations MUST be justified in writing and
  approved before merge.
- Runtime development guidance lives in `.claude/skills/` and MUST
  remain consistent with this constitution at all times.

**Version**: 1.0.0 | **Ratified**: 2026-03-01 | **Last Amended**: 2026-03-01
