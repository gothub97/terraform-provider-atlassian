---
page_title: "atlassian_confluence_space_roles Data Source - atlassian"
subcategory: "Confluence"
description: |-
  Retrieves the space roles available for a Confluence space.
---

# atlassian_confluence_space_roles (Data Source)

Use this data source to retrieve the space roles available for a Confluence space. This is useful for discovering the role IDs needed when configuring an `atlassian_confluence_space_role_assignment`.

## Example Usage

```hcl
data "atlassian_confluence_space" "docs" {
  key = "DOCS"
}

data "atlassian_confluence_space_roles" "docs" {
  space_id = data.atlassian_confluence_space.docs.id
}

output "role_names" {
  value = [for r in data.atlassian_confluence_space_roles.docs.roles : r.name]
}
```

## Argument Reference

* `space_id` - (Required) The ID of the Confluence space to list available roles for.

## Attributes Reference

* `roles` - The list of roles available for the space. Each role has the following attributes:
  * `id` - The ID of the space role.
  * `name` - The name of the space role.
  * `type` - The type of the space role (e.g., `USER`, `GROUP`).
  * `description` - The description of the space role.
