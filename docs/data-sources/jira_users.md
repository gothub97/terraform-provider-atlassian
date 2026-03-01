---
page_title: "atlassian_jira_users Data Source - atlassian"
subcategory: "Jira"
description: |-
  Searches for Jira users by display name or email address.
---

# atlassian_jira_users (Data Source)

Use this data source to search for Jira users by display name or email address. This is useful for discovering account IDs needed when configuring project role actors or group memberships.

## Example Usage

```hcl
data "atlassian_jira_users" "example" {
  query = "john"
}
```

## Argument Reference

* `query` - (Required) Search string matching display name or email address.

## Attributes Reference

* `users` - The list of matching users. Each user has the following attributes:
  * `account_id` - The account ID of the user.
  * `display_name` - The display name of the user.
  * `email_address` - The email address of the user.
  * `active` - Whether the user account is active.
