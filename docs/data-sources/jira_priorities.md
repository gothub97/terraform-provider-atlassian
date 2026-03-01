---
page_title: "atlassian_jira_priorities Data Source - atlassian"
subcategory: "Jira"
description: |-
  Retrieves all Jira priorities.
---

# atlassian_jira_priorities (Data Source)

Use this data source to retrieve all Jira priorities. This is useful for discovering priority IDs and their configurations.

## Example Usage

```hcl
data "atlassian_jira_priorities" "all" {}

output "default_priority" {
  value = [for p in data.atlassian_jira_priorities.all.priorities : p if p.is_default]
}
```

## Argument Reference

This data source has no arguments.

## Attributes Reference

* `priorities` - The list of priorities. Each priority has the following attributes:
  * `id` - The ID of the priority.
  * `name` - The name of the priority.
  * `description` - The description of the priority.
  * `status_color` - The color of the priority in hex format.
  * `icon_url` - The URL of the priority icon.
  * `is_default` - Whether this is the default priority.
