---
page_title: "atlassian_jira_screen_scheme Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira screen scheme.
---

# atlassian_jira_screen_scheme (Resource)

Manages a Jira screen scheme. A screen scheme maps issue operations (create, edit, view) to screens, controlling which fields are shown for each operation.

## Example Usage

```hcl
resource "atlassian_jira_screen" "create_screen" {
  name = "Create Screen"

  tab {
    name   = "Details"
    fields = ["summary", "description", "priority"]
  }
}

resource "atlassian_jira_screen" "edit_screen" {
  name = "Edit Screen"

  tab {
    name   = "Details"
    fields = ["summary", "description", "priority", "assignee"]
  }
}

resource "atlassian_jira_screen" "view_screen" {
  name = "View Screen"

  tab {
    name   = "Details"
    fields = ["summary", "description", "priority", "assignee", "status"]
  }
}

resource "atlassian_jira_screen_scheme" "example" {
  name              = "My Screen Scheme"
  description       = "Custom screen scheme for the project"
  default_screen_id = atlassian_jira_screen.view_screen.id
  create_screen_id  = atlassian_jira_screen.create_screen.id
  edit_screen_id    = atlassian_jira_screen.edit_screen.id
  view_screen_id    = atlassian_jira_screen.view_screen.id
}
```

## Argument Reference

* `name` - (Required) The name of the screen scheme.
* `description` - (Optional) The description of the screen scheme.
* `default_screen_id` - (Required) The ID of the default screen. This screen is used for any operation that does not have a specific screen assigned.
* `create_screen_id` - (Optional) The ID of the screen used for creating issues. If not specified, the default screen is used.
* `edit_screen_id` - (Optional) The ID of the screen used for editing issues. If not specified, the default screen is used.
* `view_screen_id` - (Optional) The ID of the screen used for viewing issues. If not specified, the default screen is used.

## Attributes Reference

* `id` - The numeric ID of the screen scheme.

## Import

Import using the screen scheme ID:

```shell
terraform import atlassian_jira_screen_scheme.example 10400
```
