# Data Model: Issues and Workflows

**Branch**: `002-issues-and-workflows` | **Date**: 2026-03-01

## Entity Relationship Diagram

```
Project (from 001)
  ├── assigned → FieldConfigurationScheme
  │                 └── maps IssueType → FieldConfiguration
  │                                        └── contains FieldConfigurationItem[]
  │                                                       └── references Field
  ├── assigned → IssueTypeScreenScheme
  │                 └── maps IssueType → ScreenScheme
  │                                        └── maps Operation → Screen
  │                                                              └── contains Tab[]
  │                                                                         └── contains Field[] (ordered)
  ├── assigned → WorkflowScheme
  │                 └── maps IssueType → Workflow
  │                                        └── contains Status[]
  │                                        └── contains Transition[]
  │                                                       └── has Condition[]
  │                                                       └── has Validator[]
  │                                                       └── has PostFunction[]
  └── contains → Issue[] (managed outside Terraform)
```

## Entities

### Field (Custom Field)

| Attribute | Type | Required | Mutable | Notes |
|-----------|------|----------|---------|-------|
| id | string | computed | no | Format: `customfield_XXXXX` |
| name | string | yes | yes | Must be unique |
| type | string | yes | no (ForceNew) | Jira field type key (e.g., `com.atlassian.jira.plugin.system.customfieldtypes:textfield`) |
| description | string | no | yes | |
| searcher_key | string | no | yes | Searcher type key |

**Identity**: Unique by `id` (assigned by Jira)
**Lifecycle**: Created → Updated (name/description/searcher) → Trashed (delete)

### FieldConfiguration

| Attribute | Type | Required | Mutable | Notes |
|-----------|------|----------|---------|-------|
| id | string | computed | no | Numeric ID as string |
| name | string | yes | yes | Must be unique, max 255 chars |
| description | string | no | yes | Max 255 chars |
| field_items | list(object) | no | yes | Ordered list of field configuration items |

**FieldConfigurationItem** (nested):

| Attribute | Type | Required | Mutable | Notes |
|-----------|------|----------|---------|-------|
| field_id | string | yes | yes | References a field (system or custom) |
| is_required | bool | no | yes | Default: false |
| is_hidden | bool | no | yes | Default: false |
| description | string | no | yes | Description override |
| renderer | string | no | yes | Renderer type |

**Identity**: Unique by `id`
**Lifecycle**: Created → Updated (name/description/items replaced atomically) → Deleted

### FieldConfigurationScheme

| Attribute | Type | Required | Mutable | Notes |
|-----------|------|----------|---------|-------|
| id | string | computed | no | Numeric ID as string |
| name | string | yes | yes | Must be unique, max 255 chars |
| description | string | no | yes | Max 1024 chars |
| mappings | list(object) | no | yes | Issue type to field configuration mappings |

**FieldConfigurationSchemeMapping** (nested):

| Attribute | Type | Required | Mutable | Notes |
|-----------|------|----------|---------|-------|
| issue_type_id | string | yes | yes | Issue type ID or "default" |
| field_configuration_id | string | yes | yes | Field configuration ID |

**Identity**: Unique by `id`
**Lifecycle**: Created → Updated (name/description/mappings) → Deleted (if not assigned to projects)

### Screen

| Attribute | Type | Required | Mutable | Notes |
|-----------|------|----------|---------|-------|
| id | string | computed | no | Numeric ID as string |
| name | string | yes | yes | Must be unique, max 255 chars |
| description | string | no | yes | Max 255 chars |
| tabs | list(object) | no | yes | Ordered list of screen tabs |

**ScreenTab** (nested):

| Attribute | Type | Required | Mutable | Notes |
|-----------|------|----------|---------|-------|
| id | string | computed | no | Tab ID assigned by Jira |
| name | string | yes | yes | Max 255 chars |
| fields | list(string) | no | yes | Ordered list of field IDs |

**Identity**: Unique by `id`
**Lifecycle**: Created (with default tab) → Updated (tabs/fields via individual API calls) → Deleted (if not in use)
**Composite operations**: Create screen → create tabs → add fields → reorder fields/tabs

### ScreenScheme

| Attribute | Type | Required | Mutable | Notes |
|-----------|------|----------|---------|-------|
| id | string | computed | no | Numeric ID as string |
| name | string | yes | yes | |
| description | string | no | yes | |
| default_screen_id | string | yes | yes | Required default screen |
| create_screen_id | string | no | yes | Screen for create operation |
| edit_screen_id | string | no | yes | Screen for edit operation |
| view_screen_id | string | no | yes | Screen for view operation |

**Identity**: Unique by `id`
**Lifecycle**: Created → Updated → Deleted (if not in use by issue type screen schemes)

### IssueTypeScreenScheme

| Attribute | Type | Required | Mutable | Notes |
|-----------|------|----------|---------|-------|
| id | string | computed | no | Numeric ID as string |
| name | string | yes | yes | |
| description | string | no | yes | |
| mappings | list(object) | no | yes | Issue type to screen scheme mappings |

**IssueTypeScreenSchemeMapping** (nested):

| Attribute | Type | Required | Mutable | Notes |
|-----------|------|----------|---------|-------|
| issue_type_id | string | yes | yes | Issue type ID or "default" |
| screen_scheme_id | string | yes | yes | Screen scheme ID |

**Identity**: Unique by `id`
**Lifecycle**: Created → Updated → Deleted (if not assigned to projects)

### Workflow

| Attribute | Type | Required | Mutable | Notes |
|-----------|------|----------|---------|-------|
| id | string | computed | no | UUID assigned by Jira |
| name | string | yes | yes | |
| description | string | no | yes | |
| statuses | list(object) | yes | yes | Workflow statuses |
| transitions | list(object) | yes | yes | Workflow transitions |

**WorkflowStatus** (nested):

| Attribute | Type | Required | Mutable | Notes |
|-----------|------|----------|---------|-------|
| status_reference | string | yes | yes | UUID reference for this status in workflow |
| status_id | string | no | computed | Jira status ID (resolved after creation) |
| name | string | yes | yes | Display name |
| status_category | string | yes | yes | TODO, IN_PROGRESS, or DONE |

**WorkflowTransition** (nested):

| Attribute | Type | Required | Mutable | Notes |
|-----------|------|----------|---------|-------|
| name | string | yes | yes | Transition name |
| from_status_reference | string | no | yes | Source status reference (empty for initial/global) |
| to_status_reference | string | yes | yes | Target status reference |
| type | string | yes | yes | initial, directed, or global |
| conditions | list(object) | no | yes | Transition conditions |
| validators | list(object) | no | yes | Transition validators |
| post_functions | list(object) | no | yes | Transition post-functions |

**WorkflowTransitionRule** (used by conditions, validators, post_functions):

| Attribute | Type | Required | Mutable | Notes |
|-----------|------|----------|---------|-------|
| rule_key | string | yes | yes | System rule key (e.g., `system:check-permission-validator`) |
| parameters | map(string) | no | yes | Rule-specific configuration key-value pairs |

**WorkflowConditionGroup** (for hierarchical conditions):

| Attribute | Type | Required | Mutable | Notes |
|-----------|------|----------|---------|-------|
| operator | string | no | yes | AND or OR (for compound conditions) |
| rules | list(WorkflowTransitionRule) | no | yes | Simple conditions at this level |
| groups | list(WorkflowConditionGroup) | no | yes | Nested condition groups |

**Identity**: Unique by `id` (UUID)
**Lifecycle**: Created (bulk API) → Updated (bulk API) → Deleted (if inactive and not in schemes)

### WorkflowScheme

| Attribute | Type | Required | Mutable | Notes |
|-----------|------|----------|---------|-------|
| id | string | computed | no | Numeric ID as string |
| name | string | yes | yes | Must be unique, max 255 chars |
| description | string | no | yes | |
| default_workflow | string | no | yes | Workflow name, defaults to "jira" |
| mappings | list(object) | no | yes | Issue type to workflow mappings |

**WorkflowSchemeMapping** (nested):

| Attribute | Type | Required | Mutable | Notes |
|-----------|------|----------|---------|-------|
| issue_type_id | string | yes | yes | Issue type ID |
| workflow_name | string | yes | yes | Workflow name |

**Identity**: Unique by `id`
**Lifecycle**: Created → Updated (direct if inactive, draft+publish if active) → Deleted (if not assigned to projects)
**State transitions**: Inactive ↔ Active (when assigned/unassigned to projects). Active schemes require draft-based updates.

### ~~Issue~~ REMOVED

The `jira_issue` resource has been removed. Issues are ephemeral work items managed outside Terraform.

## Validation Rules

- Field `name` must be unique across the Jira instance
- Field `type` is immutable after creation (ForceNew)
- Screen `name` must be unique
- FieldConfiguration `name` must be unique, max 255 chars
- FieldConfigurationScheme `name` must be unique, max 255 chars
- WorkflowScheme `name` must be unique, max 255 chars
- Workflow `statuses` must include at least one status with category TODO (for initial status)
- Workflow `transitions` must include exactly one `initial` type transition
- WorkflowTransitionRule `rule_key` must be a valid system, connect, or forge rule key
