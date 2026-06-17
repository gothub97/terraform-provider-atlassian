---
page_title: "atlassian_confluence_space_permissions Resource - atlassian"
subcategory: "Confluence"
description: |-
  Manages the granular Confluence space permissions granted to a single group for a single space.
---

# atlassian_confluence_space_permissions (Resource)

Manages the granular Confluence space permissions granted to a single group for a single space, through the Confluence v1 REST API. Each `permission` block is an individual operation/target grant; the resource adds and removes grants to converge on the declared set.

~> **Note:** This resource uses the classic (non-RBAC) Confluence space permission model and requires a site where that model is active:
> - On the **Free** edition the API rejects writes with `403 — "User isn't authorized to modify permissions because they are using the Free Edition."`
> - On sites running in **roles-only (RBAC) mode** the API rejects writes with `400 — "Space permission updates that are not from RBAC are not supported in roles-only mode."` On those sites, use [`atlassian_confluence_space_role_assignment`](confluence_space_role_assignment.md) instead.
>
> Reading permissions works regardless of mode.

## Example Usage

```hcl
resource "atlassian_confluence_space_permissions" "docs_team" {
  space_key = atlassian_confluence_space.docs.key
  group_id  = atlassian_jira_group.team.id

  permission {
    operation = "read"
    target    = "space"
  }

  permission {
    operation = "create"
    target    = "page"
  }

  permission {
    operation = "delete"
    target    = "page"
  }
}
```

## Argument Reference

* `space_key` - (Required) The key of the Confluence space. Changing this forces a new resource to be created.
* `group_id` - (Required) The Atlassian group ID that the permissions are granted to. Changing this forces a new resource to be created.
* `permission` - (Optional) One or more granular permission grant blocks. Each block supports:
    * `operation` - (Required) The operation key, e.g. `read`, `create`, `delete`, `export`, `administer`, `restrict_content`, `archive`.
    * `target` - (Required) The operation target, e.g. `page`, `blogpost`, `comment`, `attachment`, `space`.

## Attributes Reference

* `id` - Composite identifier in the form `{space_key}/{group_id}`.
* `permission.id` - The ID of each granular permission grant returned by Confluence.

## Import

Import using the composite ID `{space_key}/{group_id}`:

```shell
terraform import atlassian_confluence_space_permissions.example DOCS/0a1b2c3d-group-id
```
