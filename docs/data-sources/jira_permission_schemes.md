---
page_title: "atlassian_jira_permission_schemes Data Source - atlassian"
subcategory: "Jira"
description: |-
  Retrieves all Jira permission schemes.
---

# atlassian_jira_permission_schemes (Data Source)

Use this data source to retrieve all Jira permission schemes. This is useful for discovering scheme IDs needed when assigning permission schemes to projects.

## Example Usage

```hcl
data "atlassian_jira_permission_schemes" "all" {
}
```

## Argument Reference

This data source has no required or optional arguments.

## Attributes Reference

* `schemes` - The list of permission schemes. Each scheme has the following attributes:
  * `id` - The ID of the permission scheme.
  * `name` - The name of the permission scheme.
  * `description` - The description of the permission scheme.
