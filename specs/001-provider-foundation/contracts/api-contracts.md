# API Contracts: Provider Foundation

**Source**: `jira-api-doc/swagger-v3.json`

## Provider → Jira Cloud REST API v3

All calls use `Authorization: Basic base64(email:token)` and
`Content-Type: application/json`.

### Project

| Operation | Method | Path | Request Schema | Response Schema | Status |
|-----------|--------|------|---------------|-----------------|--------|
| Create | POST | `/rest/api/3/project` | `CreateProjectDetails` | `ProjectIdentifiers` (201) | |
| Read | GET | `/rest/api/3/project/{projectIdOrKey}` | — | `Project` (200) | |
| Update | PUT | `/rest/api/3/project/{projectIdOrKey}` | `UpdateProjectDetails` | `Project` (200) | |
| Delete | DELETE | `/rest/api/3/project/{projectIdOrKey}?enableUndo=true` | — | 204 | |
| Search | GET | `/rest/api/3/project/search` | query: startAt, maxResults, keys | `PageBeanProject` (200) | |

### Issue Type

| Operation | Method | Path | Request Schema | Response Schema | Status |
|-----------|--------|------|---------------|-----------------|--------|
| Create | POST | `/rest/api/3/issuetype` | `IssueTypeCreateBean` | `IssueTypeDetails` (201) | |
| Read | GET | `/rest/api/3/issuetype/{id}` | — | `IssueTypeDetails` (200) | |
| Update | PUT | `/rest/api/3/issuetype/{id}` | `IssueTypeUpdateBean` | `IssueTypeDetails` (200) | |
| Delete | DELETE | `/rest/api/3/issuetype/{id}` | query: alternativeIssueTypeId (opt) | 204 | |
| List All | GET | `/rest/api/3/issuetype` | — | `[]IssueTypeDetails` (200) | |
| List by Project | GET | `/rest/api/3/issuetype/project` | query: projectId (req), level (opt) | `[]IssueTypeDetails` (200) | |

### Priority

| Operation | Method | Path | Request Schema | Response Schema | Status |
|-----------|--------|------|---------------|-----------------|--------|
| Create | POST | `/rest/api/3/priority` | `CreatePriorityDetails` | `PriorityId` (201) | |
| Read | GET | `/rest/api/3/priority/{id}` | — | `Priority` (200) | |
| Update | PUT | `/rest/api/3/priority/{id}` | `UpdatePriorityDetails` | 204 | |
| Delete | DELETE | `/rest/api/3/priority/{id}` | — | 303 → task | Async |
| Search | GET | `/rest/api/3/priority/search` | query: startAt, maxResults | `PageBeanPriority` (200) | |
| Poll Task | GET | `/rest/api/3/task/{taskId}` | — | `TaskProgressBeanObject` | |

### Status

| Operation | Method | Path | Request Schema | Response Schema | Status |
|-----------|--------|------|---------------|-----------------|--------|
| Create | POST | `/rest/api/3/statuses` | `StatusCreateRequest` | `[]JiraStatus` (200) | Bulk |
| Read | GET | `/rest/api/3/statuses?id={id}` | — | `[]JiraStatus` (200) | Bulk |
| Update | PUT | `/rest/api/3/statuses` | `StatusUpdateRequest` | 204 | Bulk |
| Delete | DELETE | `/rest/api/3/statuses?id={id}` | — | 204 | Bulk |
| Search | GET | `/rest/api/3/statuses/search` | query: projectId, startAt, maxResults | `PageOfStatuses` (200) | |

### Myself

| Operation | Method | Path | Request Schema | Response Schema | Status |
|-----------|--------|------|---------------|-----------------|--------|
| Read | GET | `/rest/api/3/myself` | — | `User` (200) | |

## Terraform Resource Contracts

### HCL Schema — atlassian_jira_project

```hcl
resource "atlassian_jira_project" "example" {
  key                  = "PROJ"        # Required, ForceNew
  name                 = "My Project"  # Required
  project_type_key     = "software"    # Required, ForceNew
  project_template_key = "..."         # Optional, auto-derived, ForceNew
  lead_account_id      = "5b10a..."    # Required
  description          = "Description" # Optional
  assignee_type        = "PROJECT_LEAD" # Optional, Computed
}
```

### HCL Schema — atlassian_jira_issue_type

```hcl
resource "atlassian_jira_issue_type" "example" {
  name            = "Custom Bug"  # Required
  description     = "A bug type"  # Optional
  hierarchy_level = 0             # Optional (0=standard, -1=subtask), ForceNew
}
```

### HCL Schema — atlassian_jira_priority

```hcl
resource "atlassian_jira_priority" "example" {
  name         = "Critical"  # Required
  description  = "Urgent"    # Optional
  status_color = "#FF0000"   # Required
  icon_url     = "https://..." # Optional, mutually exclusive with avatar_id
}
```

### HCL Schema — atlassian_jira_status

```hcl
resource "atlassian_jira_status" "example" {
  name             = "In Review"    # Required
  description      = "Under review" # Optional
  status_category  = "IN_PROGRESS"  # Required (TODO, IN_PROGRESS, DONE)
  scope_type       = "GLOBAL"       # Required, ForceNew (GLOBAL or PROJECT)
  scope_project_id = "10001"        # Required if scope_type=PROJECT, ForceNew
}
```
