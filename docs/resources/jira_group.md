---
page_title: "atlassian_jira_group Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira group.
---

# atlassian_jira_group (Resource)

Manages a Jira group. This resource allows you to create and delete groups in your Jira instance. Group names are immutable; changing the name will destroy and recreate the group.

## Example Usage

```hcl
resource "atlassian_jira_group" "developers" {
  name = "developers"
}

resource "atlassian_jira_group" "qa_team" {
  name = "qa-team"
}
```

## Argument Reference

* `name` - (Required, Forces new resource) The name of the group. Cannot be changed after creation.

## Attributes Reference

* `id` - The ID of the group (UUID).
* `self` - The URL of the group.

## Import

Import using the group ID (UUID):

```shell
terraform import atlassian_jira_group.example 12345678-abcd-efgh-ijkl-123456789012
```
