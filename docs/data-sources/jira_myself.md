---
page_title: "atlassian_jira_myself Data Source - atlassian"
subcategory: "Jira"
description: |-
  Retrieves information about the currently authenticated Jira user.
---

# atlassian_jira_myself (Data Source)

Use this data source to retrieve information about the currently authenticated Jira user. This is commonly used to get the `account_id` needed for project lead assignments, group memberships, and role actors.

## Example Usage

```hcl
data "atlassian_jira_myself" "me" {}

resource "atlassian_jira_project" "example" {
  key              = "EX"
  name             = "Example Project"
  project_type_key = "software"
  lead_account_id  = data.atlassian_jira_myself.me.account_id
}
```

## Argument Reference

This data source has no arguments.

## Attributes Reference

* `id` - The account ID of the authenticated user (same as `account_id`).
* `account_id` - The account ID of the authenticated user.
* `account_type` - The type of account (e.g., `atlassian`, `app`, `customer`).
* `display_name` - The display name of the user.
* `email_address` - The email address of the user.
* `active` - Whether the user account is active.
* `time_zone` - The user's time zone.
* `locale` - The user's locale.
* `self` - The URL of the user profile.
