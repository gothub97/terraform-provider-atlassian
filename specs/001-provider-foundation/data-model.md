# Data Model: Provider Foundation

**Phase**: 1 — Design & Contracts
**Date**: 2026-03-01
**Source**: `jira-api-doc/swagger-v3.json` + spec.md Key Entities

## Provider Configuration

| Field | Type | Required | Sensitive | Env Var | Notes |
|-------|------|----------|-----------|---------|-------|
| `url` | string | Yes | No | `ATLASSIAN_URL` | Base URL of Jira Cloud instance (e.g., `https://mysite.atlassian.net`) |
| `email` | string | Yes | No | `ATLASSIAN_EMAIL` | Email for Basic Auth |
| `api_token` | string | Yes | Yes | `ATLASSIAN_API_TOKEN` | API token for Basic Auth |

**Precedence**: HCL value > environment variable. All three are required
at configure time; missing values produce a diagnostic error.

## Entity: Project

**Resource**: `atlassian_jira_project`
**API**: `POST /rest/api/3/project`, `GET /rest/api/3/project/{projectIdOrKey}`, `PUT /rest/api/3/project/{projectIdOrKey}`, `DELETE /rest/api/3/project/{projectIdOrKey}`

### Attributes

| Attribute | Terraform Type | API Field | Required | Computed | ForceNew | Sensitive | Notes |
|-----------|---------------|-----------|----------|----------|----------|-----------|-------|
| `id` | `types.String` | `id` | — | Yes | — | No | Read-only. Project ID. |
| `key` | `types.String` | `key` | Yes | — | Yes | No | Immutable. 2-10 uppercase letters. |
| `name` | `types.String` | `name` | Yes | — | — | No | |
| `project_type_key` | `types.String` | `projectTypeKey` | Yes | — | Yes | No | One of: `software`, `business`, `service_desk` |
| `project_template_key` | `types.String` | `projectTemplateKey` | No | Yes | Yes | No | Auto-derived from `project_type_key` if unset |
| `lead_account_id` | `types.String` | `leadAccountId` | Yes | — | — | No | Account ID of project lead |
| `description` | `types.String` | `description` | No | — | — | No | |
| `assignee_type` | `types.String` | `assigneeType` | No | Yes | — | No | `PROJECT_LEAD` or `UNASSIGNED` |
| `self` | `types.String` | `self` | — | Yes | — | No | Read-only. URL of the project. |

### Validation Rules

- `key`: regex `^[A-Z][A-Z0-9]{1,9}$`, unique per instance.
- `project_type_key`: one of `software`, `business`, `service_desk`.
- `assignee_type`: one of `PROJECT_LEAD`, `UNASSIGNED` (if set).

### Import

Import by project key or numeric ID: `terraform import atlassian_jira_project.example MYKEY`

## Entity: Issue Type

**Resource**: `atlassian_jira_issue_type`
**API**: `POST /rest/api/3/issuetype`, `GET /rest/api/3/issuetype/{id}`, `PUT /rest/api/3/issuetype/{id}`, `DELETE /rest/api/3/issuetype/{id}`

### Attributes

| Attribute | Terraform Type | API Field | Required | Computed | ForceNew | Sensitive | Notes |
|-----------|---------------|-----------|----------|----------|----------|-----------|-------|
| `id` | `types.String` | `id` | — | Yes | — | No | Read-only. |
| `name` | `types.String` | `name` | Yes | — | — | No | Max 60 chars, unique. |
| `description` | `types.String` | `description` | No | — | — | No | |
| `hierarchy_level` | `types.Int64` | `hierarchyLevel` | No | Yes | Yes | No | `0` = standard (default), `-1` = subtask |
| `avatar_id` | `types.Int64` | `avatarId` | No | Yes | — | No | |
| `icon_url` | `types.String` | `iconUrl` | — | Yes | — | No | Read-only. Computed from avatar. |
| `subtask` | `types.Bool` | `subtask` | — | Yes | — | No | Read-only. Derived from hierarchy_level. |
| `self` | `types.String` | `self` | — | Yes | — | No | Read-only. |

### Validation Rules

- `name`: length 1-60 characters.
- `hierarchy_level`: one of `0`, `-1`.

### Import

Import by ID: `terraform import atlassian_jira_issue_type.example 10001`

## Entity: Priority

**Resource**: `atlassian_jira_priority`
**API**: `POST /rest/api/3/priority`, `GET /rest/api/3/priority/{id}`, `PUT /rest/api/3/priority/{id}`, `DELETE /rest/api/3/priority/{id}`

### Attributes

| Attribute | Terraform Type | API Field | Required | Computed | ForceNew | Sensitive | Notes |
|-----------|---------------|-----------|----------|----------|----------|-----------|-------|
| `id` | `types.String` | `id` | — | Yes | — | No | Read-only. |
| `name` | `types.String` | `name` | Yes | — | — | No | Must be unique. |
| `description` | `types.String` | `description` | No | — | — | No | |
| `status_color` | `types.String` | `statusColor` | Yes | — | — | No | Hex color, e.g. `#FF0000` |
| `icon_url` | `types.String` | `iconUrl` | No | Yes | — | No | Mutually exclusive with `avatar_id` |
| `avatar_id` | `types.Int64` | `avatarId` | No | Yes | — | No | Mutually exclusive with `icon_url` |
| `is_default` | `types.Bool` | `isDefault` | — | Yes | — | No | Read-only. |
| `self` | `types.String` | `self` | — | Yes | — | No | Read-only. |

### Validation Rules

- `status_color`: regex `^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$`.
- `icon_url` and `avatar_id` are mutually exclusive — at most one set.

### Special: Async Delete

Delete returns 303 → poll `GET /rest/api/3/task/{taskId}` until
`COMPLETE` or `FAILED`.

### Import

Import by ID: `terraform import atlassian_jira_priority.example 10100`

## Entity: Status

**Resource**: `atlassian_jira_status`
**API**: `POST /rest/api/3/statuses`, `GET /rest/api/3/statuses?id=...`, `PUT /rest/api/3/statuses`, `DELETE /rest/api/3/statuses?id=...`

### Attributes

| Attribute | Terraform Type | API Field | Required | Computed | ForceNew | Sensitive | Notes |
|-----------|---------------|-----------|----------|----------|----------|-----------|-------|
| `id` | `types.String` | `id` | — | Yes | — | No | Read-only. |
| `name` | `types.String` | `name` | Yes | — | — | No | |
| `description` | `types.String` | `description` | No | — | — | No | |
| `status_category` | `types.String` | `statusCategory` | Yes | — | — | No | `TODO`, `IN_PROGRESS`, or `DONE` |
| `scope_type` | `types.String` | `scope.type` | Yes | — | Yes | No | `GLOBAL` or `PROJECT` |
| `scope_project_id` | `types.String` | `scope.project.id` | No | — | Yes | No | Required when `scope_type` is `PROJECT` |

### Validation Rules

- `status_category`: one of `TODO`, `IN_PROGRESS`, `DONE`.
- `scope_type`: one of `GLOBAL`, `PROJECT`.
- `scope_project_id`: required if `scope_type == PROJECT`, forbidden
  otherwise.

### Bulk API Adaptation

All CRUD operations use bulk endpoints with single-element arrays.
- Create: `POST /rest/api/3/statuses` with `{ scope: {...}, statuses: [{ name, statusCategory, description }] }`
- Update: `PUT /rest/api/3/statuses` with `{ statuses: [{ id, name, statusCategory, description }] }`
- Delete: `DELETE /rest/api/3/statuses?id=<id>`
- Read: `GET /rest/api/3/statuses?id=<id>`

### Import

Import by ID: `terraform import atlassian_jira_status.example 10200`

## Entity: Myself (Data Source Only)

**Data Source**: `atlassian_jira_myself`
**API**: `GET /rest/api/3/myself`

### Attributes

| Attribute | Terraform Type | API Field | Computed | Notes |
|-----------|---------------|-----------|----------|-------|
| `account_id` | `types.String` | `accountId` | Yes | Unique Atlassian ID |
| `account_type` | `types.String` | `accountType` | Yes | `atlassian`, `app`, or `customer` |
| `display_name` | `types.String` | `displayName` | Yes | |
| `email_address` | `types.String` | `emailAddress` | Yes | May be null per privacy settings |
| `active` | `types.Bool` | `active` | Yes | |
| `time_zone` | `types.String` | `timeZone` | Yes | |
| `locale` | `types.String` | `locale` | Yes | May be null |
| `self` | `types.String` | `self` | Yes | URL of the user |

## Data Source: Project

**Data Source**: `atlassian_jira_project`
**API**: `GET /rest/api/3/project/{projectIdOrKey}`

### Attributes

Same as the Project resource attributes, all Computed. Plus:
- `key` (Required input): the project key to look up.

## Data Source: Issue Types

**Data Source**: `atlassian_jira_issue_types`
**API**: `GET /rest/api/3/issuetype` (all) or `GET /rest/api/3/issuetype/project?projectId=...` (filtered)

### Attributes

| Attribute | Terraform Type | Required | Computed | Notes |
|-----------|---------------|----------|----------|-------|
| `project_id` | `types.String` | No | — | Optional filter |
| `issue_types` | `types.List` of objects | — | Yes | List of issue type objects |

Each issue type object: `id`, `name`, `description`, `hierarchy_level`,
`subtask`, `icon_url`.

## Data Source: Priorities

**Data Source**: `atlassian_jira_priorities`
**API**: `GET /rest/api/3/priority/search`

### Attributes

| Attribute | Terraform Type | Required | Computed | Notes |
|-----------|---------------|----------|----------|-------|
| `priorities` | `types.List` of objects | — | Yes | List of priority objects |

Each priority object: `id`, `name`, `description`, `status_color`,
`icon_url`, `is_default`.

## Data Source: Statuses

**Data Source**: `atlassian_jira_statuses`
**API**: `GET /rest/api/3/statuses/search`

### Attributes

| Attribute | Terraform Type | Required | Computed | Notes |
|-----------|---------------|----------|----------|-------|
| `project_id` | `types.String` | No | — | Optional filter |
| `statuses` | `types.List` of objects | — | Yes | List of status objects |

Each status object: `id`, `name`, `description`, `status_category`,
`scope_type`, `scope_project_id`.

## Relationships

```text
Provider Configuration
  └── authenticates via → GET /rest/api/3/myself
  └── constructs → atlassian.Client

atlassian.Client
  └── used by → all resources and data sources

Project (atlassian_jira_project)
  ├── has many → Issue Types (via issue type scheme, out of scope)
  ├── has many → Statuses (when scope_type = PROJECT)
  └── no direct FK to Priority (priorities are global)

Issue Type (atlassian_jira_issue_type)
  └── global entity, not scoped to a project

Priority (atlassian_jira_priority)
  └── global entity, not scoped to a project

Status (atlassian_jira_status)
  ├── scope_type = GLOBAL → available to all company-managed projects
  └── scope_type = PROJECT → scoped to a specific project (by scope_project_id)
```
