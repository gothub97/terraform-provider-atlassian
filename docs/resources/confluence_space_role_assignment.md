---
page_title: "atlassian_confluence_space_role_assignment Resource - atlassian"
subcategory: "Confluence"
description: |-
  Manages a Confluence space role assignment for a single principal.
---

# atlassian_confluence_space_role_assignment (Resource)

Manages a Confluence space role assignment for a single principal (user, group, or access class), through the Confluence v2 REST API. Use the `atlassian_confluence_space_roles` data source to discover the role IDs available for a space.

~> **Note:** This resource requires the Confluence **space roles** capability to be enabled on the site. On sites where it is not provisioned, the `/wiki/api/v2/spaces/{id}/role-assignments` and roles endpoints return `404 NOT_FOUND`. If your site predates space roles, manage group access with `atlassian_confluence_space_permissions` instead.

## Example Usage

```hcl
data "atlassian_confluence_space_roles" "docs" {
  space_id = atlassian_confluence_space.docs.id
}

resource "atlassian_confluence_space_role_assignment" "docs_team" {
  space_id       = atlassian_confluence_space.docs.id
  principal_type = "GROUP"
  principal_id   = atlassian_jira_group.team.id
  role_id        = data.atlassian_confluence_space_roles.docs.roles[0].id
}
```

## Argument Reference

* `space_id` - (Required) The ID of the Confluence space. Changing this forces a new resource to be created.
* `principal_id` - (Required) The ID of the principal (e.g. the group ID). Changing this forces a new resource to be created.
* `role_id` - (Required) The ID of the space role to assign to the principal.
* `principal_type` - (Optional) The type of the principal: `USER`, `GROUP`, or `ACCESS_CLASS`. Defaults to `GROUP`.

## Attributes Reference

* `id` - Composite identifier in the form `{space_id}/{principal_id}`.

## Import

Import using the composite ID `{space_id}/{principal_id}`:

```shell
terraform import atlassian_confluence_space_role_assignment.example 123456/0a1b2c3d-group-id
```
