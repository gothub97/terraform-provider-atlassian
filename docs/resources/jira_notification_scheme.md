---
page_title: "atlassian_jira_notification_scheme Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira notification scheme.
---

# atlassian_jira_notification_scheme (Resource)

Manages a Jira notification scheme. Notification schemes control which users receive email notifications for specific project events. During updates, existing notifications are removed and replaced with the declared set.

## Example Usage

```hcl
resource "atlassian_jira_notification_scheme" "standard" {
  name        = "Standard Notifications"
  description = "Notifications for standard projects"

  notification {
    event_id          = "1"
    notification_type = "CurrentAssignee"
  }

  notification {
    event_id          = "1"
    notification_type = "Reporter"
  }

  notification {
    event_id          = "2"
    notification_type = "AllWatchers"
  }

  notification {
    event_id          = "1"
    notification_type = "Group"
    parameter         = "12345678-abcd-efgh-ijkl-123456789012"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the notification scheme.
* `description` - (Optional) The description of the notification scheme.
* `notification` - (Optional) One or more notification blocks. Each block supports:
    * `event_id` - (Required) The event ID (e.g., `1` for Issue Created, `2` for Issue Updated).
    * `notification_type` - (Required) The notification recipient type (e.g., `CurrentAssignee`, `Reporter`, `User`, `Group`, `ProjectRole`, `EmailAddress`, `AllWatchers`).
    * `parameter` - (Optional) The parameter value (user account ID, group ID, role ID, email address, etc). Not required for types like `CurrentAssignee`, `Reporter`, `CurrentUser`, `AllWatchers`.

## Attributes Reference

* `id` - The ID of the notification scheme.
* `notification.id` - The ID of each notification entry.

## Import

Import using the notification scheme ID:

```shell
terraform import atlassian_jira_notification_scheme.example 10200
```
