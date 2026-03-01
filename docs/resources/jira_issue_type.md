---
page_title: "atlassian_jira_issue_type Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira issue type.
---

# atlassian_jira_issue_type (Resource)

Manages a Jira issue type. Issue types distinguish different types of work (e.g., Bug, Story, Task) and can be standard issues or subtasks.

## Example Usage

### Standard Issue Type

```hcl
resource "atlassian_jira_issue_type" "feature" {
  name        = "Feature Request"
  description = "A request for a new feature"
}
```

### Subtask Issue Type

```hcl
resource "atlassian_jira_issue_type" "subtask" {
  name            = "Technical Subtask"
  description     = "A technical subtask"
  hierarchy_level = -1
}
```

## Argument Reference

* `name` - (Required) The name of the issue type. Maximum 60 characters.
* `description` - (Optional) The description of the issue type.
* `hierarchy_level` - (Optional, ForceNew) The hierarchy level. Use `0` for standard issue types (default) or `-1` for subtask types.
* `avatar_id` - (Optional) The ID of the avatar to use for this issue type.

## Attributes Reference

* `id` - The ID of the issue type.
* `icon_url` - The URL of the issue type icon.
* `subtask` - Whether this is a subtask issue type.
* `self` - The URL of the issue type.

## Import

Import using the issue type ID:

```shell
terraform import atlassian_jira_issue_type.example 10001
```
