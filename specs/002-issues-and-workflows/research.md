# Research: Issues and Workflows

**Branch**: `002-issues-and-workflows` | **Date**: 2026-03-01

## Decision Log

### D-001: Workflow Rule Modeling Strategy
**Decision**: Model all system rule types as structured attributes keyed by `rule_key`. Each rule type has a `parameters` map(string) for configuration, matching the Jira API's storage format (all parameters are string key-value pairs).

**Rationale**: The Jira API stores all rule parameters as string key-value pairs regardless of logical type. Modeling each rule as a block with `rule_key` + `parameters` map provides full coverage without needing per-type Go structs for 21+ rule types. Terraform validators can enforce parameter requirements per rule_key.

**Alternatives considered**:
- Per-type nested blocks (e.g., `permission_condition {}`, `field_validator {}`) — would require 21+ block types, massive schema, and constant updates as Jira adds rule types.
- Raw JSON string — loses all schema validation.

### D-002: Screen Resource Composite Operations
**Decision**: The screen resource manages tabs and fields as nested blocks within a single resource. Create/update performs diff-based operations (add/remove/reorder tabs and fields individually).

**Rationale**: The Jira Screen API requires multiple sequential calls (create screen → create tabs → add fields → reorder). A single composite resource avoids orphaned tabs/fields and matches how users think about screens. Tab and field ordering follow declaration order in HCL.

**Alternatives considered**:
- Separate `jira_screen_tab` and `jira_screen_tab_field` resources — too granular, complex dependency management, poor UX.
- Replace-all strategy (delete all tabs/fields on update, recreate) — risky, loses tab IDs, may break references.

### D-003: Field Configuration Item Management
**Decision**: Use PUT replace-all semantics matching the Jira API. The resource sends the complete list of field items on every update.

**Rationale**: The Jira API's `PUT /fieldconfiguration/{id}/fields` replaces all items atomically. This simplifies the implementation — no diff logic needed. The Terraform state holds the full list.

**Alternatives considered**:
- Diff-based add/remove of individual items — API doesn't support individual item operations.

### D-004: Issue Description Format
**Decision**: Store ADF as a JSON string in Terraform state. Users provide ADF JSON in their HCL configuration.

**Rationale**: The Jira API only accepts ADF (Atlassian Document Format) for rich text fields. Plain text is not supported. Storing as JSON string is the simplest representation that preserves the full ADF structure.

**Alternatives considered**:
- Markdown-to-ADF conversion — adds complexity, lossy conversion, maintenance burden.
- Structured ADF blocks in HCL — too verbose for document content.

### D-005: Workflow Scheme Draft Management
**Decision**: The provider handles drafts transparently. On update, if the scheme is active (has project usages), the provider creates/updates a draft and publishes it. The publish operation is async (303 redirect → task polling).

**Rationale**: Users should not need to manage drafts manually. The provider checks project usages to determine if draft flow is needed. The existing `WaitForTask` client method handles async polling.

**Alternatives considered**:
- Expose draft as separate resource — too complex for users.
- Require manual draft management — poor UX, error-prone.

### D-006: Issue Status Transitions
**Decision**: When the `status` attribute changes, the provider queries available transitions and executes the appropriate one. If no direct transition exists, the provider returns an error listing available transitions.

**Rationale**: Jira enforces workflow transitions — you cannot directly set a status. The provider must use `GET /issue/{key}/transitions` to find valid paths and `POST /issue/{key}/transitions` to execute. Multi-step transitions (requiring intermediate statuses) are out of scope — the user must specify reachable statuses.

**Alternatives considered**:
- Multi-step pathfinding (BFS through workflow graph) — too complex, unpredictable side effects from intermediate transitions.
- Ignore status on update — contradicts the clarification decision.

### D-007: Custom Field Values on Issues
**Decision**: Use `custom_fields` attribute of type `map(string)` where keys are field IDs and values are JSON-encoded strings.

**Rationale**: Jira custom fields have heterogeneous types (text, number, select, multi-select, user, date). A map(string) with JSON-encoded values provides universal coverage. Users can use `jsonencode()` in HCL for complex values.

**Alternatives considered**:
- Typed attributes per field — impossible since custom fields are user-defined.
- Dynamic types — constitution prohibits `types.Dynamic`.

### D-008: Workflow Conditions Hierarchy
**Decision**: Model conditions as a recursive structure supporting AND/OR compound conditions with nested simple conditions. Use a `condition` block with optional `operator` (AND/OR) and nested `conditions` list, or a flat `rule_key` + `parameters` for simple conditions.

**Rationale**: The Jira API supports hierarchical conditions with AND/OR operators combining simple conditions. This must be representable in Terraform schema.

**Alternatives considered**:
- Flat condition list only — would lose AND/OR grouping capability.

## Rule Type Inventory

### Validators (6 types)
| Rule Key | Parameters |
|----------|-----------|
| `system:check-permission-validator` | `permissionKey` (required) |
| `system:parent-or-child-blocking-validator` | `blocker` (required: "PARENT"), `statusIds` (required, comma-separated) |
| `system:previous-status-validator` | `previousStatusIds` (required), `mostRecentStatusOnly` (required: "true"/"false") |
| `system:validate-field-value` | `ruleType` (required: fieldRequired/fieldChanged/fieldHasSingleValue/fieldMatchesRegularExpression/dateFieldComparison/windowDateComparison), plus sub-type specific params |
| `system:proforma-forms-attached` | (none) |
| `system:proforma-forms-submitted` | (none) |

### Conditions (9 types)
| Rule Key | Parameters |
|----------|-----------|
| `system:check-field-value` | `fieldId`, `fieldValue` (JSON array), `comparator`, `comparisonType` |
| `system:restrict-issue-transition` | `accountIds`, `roleIds`, `groupIds`, `permissionKeys`, `groupCustomFields`, `allowUserCustomFields`, `denyUserCustomFields` (all optional) |
| `system:previous-status-condition` | `previousStatusIds`, `not`, `mostRecentStatusOnly`, `includeCurrentStatus`, `ignoreLoopTransitions` |
| `system:parent-or-child-blocking-condition` | `blocker` ("CHILD"), `statusIds` |
| `system:separation-of-duties` | `fromStatusId`, `toStatusId` |
| `system:restrict-from-all-users` | `restrictMode` ("users"/"usersAndAPI") |
| `system:jsd-approvals-block-until-approved` | `approvalConfigurationJson` |
| `system:jsd-approvals-block-until-rejected` | `approvalConfigurationJson` |
| `system:block-in-progress-approval` | (none) |

### Post Functions (4 types)
| Rule Key | Parameters |
|----------|-----------|
| `system:change-assignee` | `type` (required: to-selected-user/to-unassigned/to-current-user/to-default-user), `accountId` (conditional) |
| `system:copy-value-from-other-field` | `sourceFieldKey`, `targetFieldKey`, `issueSource` (optional: SAME/PARENT) |
| `system:update-field` | `field`, `value`, `mode` (append/replace) |
| `system:trigger-webhook` | `webhookId` |

### Screen Rules (2 types)
| Rule Key | Parameters |
|----------|-----------|
| `system:remind-people-to-update-fields` | `remindingFieldIds`, `remindingMessage`, `remindingAlwaysAsk` |
| `system:transition-screen` | `screenId` |

### Ecosystem Rules
- Connect: `connect:expression-validator`, `connect:expression-condition`, `connect:remote-workflow-function`
- Forge: `forge:expression-validator`, `forge:expression-condition`, `forge:workflow-post-function`
- Both use: `ruleKey`, `appKey`/`key`, `config`, `id`, `disabled`, `tag`

## API Patterns

### Pagination
All list endpoints use standard pagination: `startAt`, `maxResults`, `total`, `isLast`. The existing `atlassian.Paginate[T]()` generic handles this.

### Async Operations
- Workflow scheme draft publish: 303 redirect → task polling via `WaitForTask()`
- Custom field deletion: May be async (trash operation)

### Replacement Semantics
- Field configuration items: PUT replaces all items
- Field configuration scheme mappings: PUT replaces all mappings
- Workflow scheme issue type mappings: Included in full scheme PUT
- Screen tabs/fields: Individual add/remove/reorder operations (no bulk replace)

### ADF Format
Rich text fields (description, environment, custom rich text) use ADF only. Structure:
```json
{
  "version": 1,
  "type": "doc",
  "content": [{"type": "paragraph", "content": [{"type": "text", "text": "..."}]}]
}
```
