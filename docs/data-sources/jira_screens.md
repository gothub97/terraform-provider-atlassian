---
page_title: "atlassian_jira_screens Data Source - atlassian"
subcategory: "Jira"
description: |-
  Retrieves all Jira screens.
---

# atlassian_jira_screens (Data Source)

Use this data source to retrieve all Jira screens. Screens define the fields displayed when users create, edit, or transition issues. This data source is useful for discovering screen IDs needed when configuring screen schemes or screen tab associations.

## Example Usage

```hcl
data "atlassian_jira_screens" "all" {
}
```

### Look Up a Screen by Name

```hcl
data "atlassian_jira_screens" "all" {
}

locals {
  default_screen = [
    for s in data.atlassian_jira_screens.all.screens : s
    if s.name == "Default Screen"
  ]
}
```

## Argument Reference

This data source has no arguments.

## Attributes Reference

* `screens` - The list of screens. Each screen has the following attributes:
  * `id` - The ID of the screen.
  * `name` - The name of the screen.
  * `description` - The description of the screen.
