---
page_title: "atlassian_jira_priority Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira priority.
---

# atlassian_jira_priority (Resource)

Manages a Jira priority. Priorities indicate the relative importance of an issue (e.g., High, Medium, Low).

## Example Usage

```hcl
resource "atlassian_jira_priority" "critical" {
  name         = "Critical"
  description  = "Blocks release, must be fixed immediately"
  status_color = "#FF0000"
}

resource "atlassian_jira_priority" "low" {
  name         = "Low"
  description  = "Nice to have, no immediate impact"
  status_color = "#009900"
}
```

## Argument Reference

* `name` - (Required) The name of the priority.
* `status_color` - (Required) The color of the priority in hex format (e.g., `#FF0000` or `#FFF`).
* `description` - (Optional) The description of the priority.
* `icon_url` - (Optional) The URL of an icon for the priority. Conflicts with `avatar_id`.
* `avatar_id` - (Optional) The ID of an avatar for the priority. Conflicts with `icon_url`.

## Attributes Reference

* `id` - The ID of the priority.
* `is_default` - Whether this is the default priority.
* `self` - The URL of the priority.

## Import

Import using the priority ID:

```shell
terraform import atlassian_jira_priority.example 10001
```
