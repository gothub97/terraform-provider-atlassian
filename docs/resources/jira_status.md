---
page_title: "atlassian_jira_status Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira status.
---

# atlassian_jira_status (Resource)

Manages a Jira status. Statuses represent the state of an issue at a given point in a workflow (e.g., "To Do", "In Progress", "Done"). Statuses can be scoped globally or to a specific project.

## Example Usage

### Global Status

```hcl
resource "atlassian_jira_status" "in_review" {
  name            = "In Review"
  description     = "Issue is being reviewed"
  status_category = "IN_PROGRESS"
  scope_type      = "GLOBAL"
}
```

### Project-scoped Status

```hcl
resource "atlassian_jira_status" "testing" {
  name             = "Testing"
  status_category  = "IN_PROGRESS"
  scope_type       = "PROJECT"
  scope_project_id = atlassian_jira_project.example.id
}
```

## Argument Reference

* `name` - (Required) The name of the status.
* `status_category` - (Required) The category of the status. Must be one of: `TODO`, `IN_PROGRESS`, `DONE`.
* `scope_type` - (Required, ForceNew) The scope of the status. Must be one of: `GLOBAL`, `PROJECT`.
* `description` - (Optional) The description of the status.
* `scope_project_id` - (Optional, ForceNew) The project ID for project-scoped statuses. Required when `scope_type` is `PROJECT`.

## Attributes Reference

* `id` - The ID of the status.

## Import

Import using the status ID:

```shell
terraform import atlassian_jira_status.example 10001
```
