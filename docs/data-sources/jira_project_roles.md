---
page_title: "atlassian_jira_project_roles Data Source - atlassian"
subcategory: "Jira"
description: |-
  Retrieves all Jira project roles.
---

# atlassian_jira_project_roles (Data Source)

Use this data source to retrieve all Jira project roles. This is useful for discovering role IDs needed when configuring project role actors or permission scheme grants.

## Example Usage

```hcl
data "atlassian_jira_project_roles" "all" {
}
```

## Argument Reference

This data source has no required or optional arguments.

## Attributes Reference

* `roles` - The list of project roles. Each role has the following attributes:
  * `id` - The ID of the project role.
  * `name` - The name of the project role.
  * `description` - The description of the project role.
