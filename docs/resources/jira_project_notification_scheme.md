---
page_title: "atlassian_jira_project_notification_scheme Resource - atlassian"
subcategory: "Jira"
description: |-
  Assigns a notification scheme to a Jira project.
---

# atlassian_jira_project_notification_scheme (Resource)

Assigns a notification scheme to a Jira project. When this resource is destroyed, the project reverts to the default notification scheme.

## Example Usage

```hcl
resource "atlassian_jira_notification_scheme" "standard" {
  name = "Standard Notifications"
}

resource "atlassian_jira_project_notification_scheme" "example" {
  project_key = "PROJ"
  scheme_id   = atlassian_jira_notification_scheme.standard.id
}
```

## Argument Reference

* `project_key` - (Required) The key of the project.
* `scheme_id` - (Required) The ID of the notification scheme to assign.

## Attributes Reference

No additional attributes are exported.

## Import

Import using the project key:

```shell
terraform import atlassian_jira_project_notification_scheme.example PROJ
```
