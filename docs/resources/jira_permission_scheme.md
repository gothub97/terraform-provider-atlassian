---
page_title: "atlassian_jira_permission_scheme Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira permission scheme.
---

# atlassian_jira_permission_scheme (Resource)

Manages a Jira permission scheme. Permission schemes control which users and groups have specific permissions within projects. Updates use a PUT operation that overwrites all grants, so the full set of permissions must be declared.

## Example Usage

```hcl
resource "atlassian_jira_permission_scheme" "standard" {
  name        = "Standard Permission Scheme"
  description = "Default permissions for standard projects"

  permission {
    permission  = "ADMINISTER_PROJECTS"
    holder_type = "projectLead"
  }

  permission {
    permission  = "BROWSE_PROJECTS"
    holder_type = "anyone"
  }

  permission {
    permission   = "CREATE_ISSUES"
    holder_type  = "group"
    holder_value = atlassian_jira_group.developers.id
  }

  permission {
    permission   = "EDIT_ISSUES"
    holder_type  = "projectRole"
    holder_value = "10360"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the permission scheme.
* `description` - (Optional) The description of the permission scheme.
* `permission` - (Optional) One or more permission grant blocks. Each block supports:
    * `permission` - (Required) The permission key (e.g., `ADMINISTER_PROJECTS`, `BROWSE_PROJECTS`, `CREATE_ISSUES`).
    * `holder_type` - (Required) The type of the permission holder (e.g., `group`, `user`, `projectRole`, `anyone`, `reporter`, `assignee`, `projectLead`, `currentUser`).
    * `holder_value` - (Optional) The value for the holder (group ID, user account ID, role ID, etc). Not required for holder types like `anyone`, `reporter`, `assignee`, `projectLead`, `currentUser`.

## Attributes Reference

* `id` - The numeric ID of the permission scheme.
* `self` - The URL of the permission scheme.
* `permission.id` - The ID of each permission grant.

## Import

Import using the permission scheme ID:

```shell
terraform import atlassian_jira_permission_scheme.example 10100
```
