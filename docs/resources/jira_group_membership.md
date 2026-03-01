---
page_title: "atlassian_jira_group_membership Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages membership of a user in a Jira group.
---

# atlassian_jira_group_membership (Resource)

Manages membership of a user in a Jira group. This resource adds a user to a group and removes them when destroyed. Both attributes are immutable; changing either will destroy and recreate the membership.

## Example Usage

```hcl
resource "atlassian_jira_group" "developers" {
  name = "developers"
}

resource "atlassian_jira_group_membership" "john" {
  group_id   = atlassian_jira_group.developers.id
  account_id = "5b10ac8d82e05b22cc7d4ef5"
}
```

## Argument Reference

* `group_id` - (Required, Forces new resource) The ID of the group (UUID).
* `account_id` - (Required, Forces new resource) The account ID of the user.

## Attributes Reference

* `id` - The composite ID of the group membership (`groupId/accountId`).

## Import

Import using the format `groupId/accountId`:

```shell
terraform import atlassian_jira_group_membership.example 12345678-abcd-efgh-ijkl-123456789012/5b10ac8d82e05b22cc7d4ef5
```
