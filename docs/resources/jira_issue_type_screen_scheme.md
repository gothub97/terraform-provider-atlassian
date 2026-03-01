---
page_title: "atlassian_jira_issue_type_screen_scheme Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira issue type screen scheme.
---

# atlassian_jira_issue_type_screen_scheme (Resource)

Manages a Jira issue type screen scheme. An issue type screen scheme maps issue types to screen schemes, allowing different issue types to display different screens during create, edit, and view operations.

## Example Usage

```hcl
resource "atlassian_jira_screen_scheme" "default_scheme" {
  name              = "Default Screen Scheme"
  default_screen_id = "1"
}

resource "atlassian_jira_screen_scheme" "bug_scheme" {
  name              = "Bug Screen Scheme"
  default_screen_id = "2"
}

resource "atlassian_jira_issue_type_screen_scheme" "example" {
  name        = "My Issue Type Screen Scheme"
  description = "Maps issue types to screen schemes"

  mapping {
    issue_type_id    = "default"
    screen_scheme_id = atlassian_jira_screen_scheme.default_scheme.id
  }

  mapping {
    issue_type_id    = "10001"
    screen_scheme_id = atlassian_jira_screen_scheme.bug_scheme.id
  }
}
```

## Argument Reference

* `name` - (Required) The name of the issue type screen scheme.
* `description` - (Optional) The description of the issue type screen scheme.

### mapping Block

The `mapping` block defines the association between an issue type and a screen scheme. You may specify zero or more `mapping` blocks. A `"default"` mapping should be included to define the screen scheme used for unmapped issue types.

* `issue_type_id` - (Required) The ID of the issue type, or `"default"` for the default mapping.
* `screen_scheme_id` - (Required) The ID of the screen scheme to associate with this issue type.

## Attributes Reference

* `id` - The numeric ID of the issue type screen scheme.

## Import

Import using the issue type screen scheme ID:

```shell
terraform import atlassian_jira_issue_type_screen_scheme.example 10500
```
