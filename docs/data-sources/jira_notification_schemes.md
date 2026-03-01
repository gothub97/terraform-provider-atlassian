---
page_title: "atlassian_jira_notification_schemes Data Source - atlassian"
subcategory: "Jira"
description: |-
  Retrieves all Jira notification schemes.
---

# atlassian_jira_notification_schemes (Data Source)

Use this data source to retrieve all Jira notification schemes. This is useful for discovering scheme IDs needed when assigning notification schemes to projects.

## Example Usage

```hcl
data "atlassian_jira_notification_schemes" "all" {
}
```

## Argument Reference

This data source has no required or optional arguments.

## Attributes Reference

* `schemes` - The list of notification schemes. Each scheme has the following attributes:
  * `id` - The ID of the notification scheme.
  * `name` - The name of the notification scheme.
  * `description` - The description of the notification scheme.
