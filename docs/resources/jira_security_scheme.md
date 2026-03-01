---
page_title: "atlassian_jira_security_scheme Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira issue security scheme.
---

# atlassian_jira_security_scheme (Resource)

Manages a Jira issue security scheme. Issue security schemes control which users can view individual issues within a project. Each scheme contains security levels, and each level defines members who are granted access. During updates, existing levels are deleted asynchronously and replaced with the declared set.

## Example Usage

```hcl
resource "atlassian_jira_security_scheme" "confidential" {
  name        = "Confidential Security Scheme"
  description = "Restricts access to sensitive issues"

  level {
    name        = "Internal Only"
    description = "Visible to internal team members"
    is_default  = true

    member {
      type      = "group"
      parameter = atlassian_jira_group.developers.id
    }

    member {
      type = "reporter"
    }
  }

  level {
    name        = "Executives"
    description = "Visible to executives only"

    member {
      type      = "projectrole"
      parameter = "10360"
    }
  }
}
```

## Argument Reference

* `name` - (Required) The name of the security scheme.
* `description` - (Optional) The description of the security scheme.
* `level` - (Optional) One or more security level blocks. Each block supports:
    * `name` - (Required) The name of the security level.
    * `description` - (Optional) The description of the security level.
    * `is_default` - (Optional) Whether this is the default security level.
    * `member` - (Optional) One or more member blocks. Each block supports:
        * `type` - (Required) The member type (`reporter`, `group`, `user`, `projectrole`, `applicationRole`).
        * `parameter` - (Optional) The parameter value (group ID, user account ID, role ID, etc). Not required for the `reporter` type.

## Attributes Reference

* `id` - The ID of the security scheme.
* `default_level_id` - The ID of the default security level.
* `level.id` - The ID of each security level.
* `level.member.id` - The ID of each member entry.

## Import

Import using the security scheme ID:

```shell
terraform import atlassian_jira_security_scheme.example 10300
```
