---
page_title: "atlassian_jira_issue_types Data Source - atlassian"
subcategory: "Jira"
description: |-
  Retrieves all Jira issue types, optionally filtered by project.
---

# atlassian_jira_issue_types (Data Source)

Use this data source to retrieve all Jira issue types, optionally filtered by project. This is useful for discovering issue type IDs needed when configuring workflows, screens, and field configurations.

## Example Usage

### All Issue Types

```hcl
data "atlassian_jira_issue_types" "all" {}
```

### Issue Types for a Project

```hcl
data "atlassian_jira_issue_types" "project" {
  project_id = atlassian_jira_project.example.id
}
```

## Argument Reference

* `project_id` - (Optional) Filter issue types by project ID.

## Attributes Reference

* `issue_types` - The list of issue types. Each issue type has the following attributes:
  * `id` - The ID of the issue type.
  * `name` - The name of the issue type.
  * `description` - The description of the issue type.
  * `hierarchy_level` - The hierarchy level (`0` for standard, `-1` for subtask).
  * `subtask` - Whether this is a subtask issue type.
  * `icon_url` - The URL of the issue type icon.
