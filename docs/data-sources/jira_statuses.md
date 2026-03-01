---
page_title: "atlassian_jira_statuses Data Source - atlassian"
subcategory: "Jira"
description: |-
  Retrieves all Jira statuses, optionally filtered by project.
---

# atlassian_jira_statuses (Data Source)

Use this data source to retrieve all Jira statuses, optionally filtered by project. This is useful for discovering status IDs needed when configuring workflows.

## Example Usage

### All Statuses

```hcl
data "atlassian_jira_statuses" "all" {}
```

### Statuses for a Project

```hcl
data "atlassian_jira_statuses" "project" {
  project_id = atlassian_jira_project.example.id
}
```

## Argument Reference

* `project_id` - (Optional) Filter statuses by project ID.

## Attributes Reference

* `statuses` - The list of statuses. Each status has the following attributes:
  * `id` - The ID of the status.
  * `name` - The name of the status.
  * `description` - The description of the status.
  * `status_category` - The status category (`TODO`, `IN_PROGRESS`, `DONE`).
  * `scope_type` - The scope of the status (`GLOBAL` or `PROJECT`).
  * `scope_project_id` - The project ID if the status is project-scoped.
