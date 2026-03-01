---
page_title: "atlassian_jira_screen Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira screen with tabs and fields.
---

# atlassian_jira_screen (Resource)

Manages a Jira screen with tabs and fields. Screens define which fields are displayed when creating, editing, or viewing issues, and organize them into tabs.

## Example Usage

```hcl
resource "atlassian_jira_screen" "bug_screen" {
  name        = "Bug Screen"
  description = "Screen for creating and editing bugs"

  tab {
    name   = "General"
    fields = ["summary", "description", "priority", "assignee"]
  }

  tab {
    name   = "Details"
    fields = ["components", "fixVersions", "labels"]
  }
}

resource "atlassian_jira_screen" "simple_screen" {
  name = "Simple Screen"

  tab {
    name   = "Default"
    fields = ["summary", "description"]
  }
}
```

## Argument Reference

* `name` - (Required) The name of the screen.
* `description` - (Optional) The description of the screen.

### tab Block

The `tab` block defines a tab on the screen. Tabs are ordered by their position in the configuration. You may specify zero or more `tab` blocks.

* `name` - (Required) The name of the tab.
* `fields` - (Optional) Ordered list of field IDs on this tab. Fields are displayed in the order specified.

## Attributes Reference

* `id` - The numeric ID of the screen.
* `tab.id` - The numeric ID of each tab, assigned by Jira after creation.

## Import

Import using the screen ID:

```shell
terraform import atlassian_jira_screen.example 10300
```
