---
page_title: "atlassian_confluence_space Data Source - atlassian"
subcategory: "Confluence"
description: |-
  Retrieves information about a Confluence space by its key.
---

# atlassian_confluence_space (Data Source)

Use this data source to retrieve information about an existing Confluence space by its key. This is useful for discovering the numeric space ID and homepage ID needed when configuring space permissions or role assignments.

## Example Usage

```hcl
data "atlassian_confluence_space" "docs" {
  key = "DOCS"
}

output "space_id" {
  value = data.atlassian_confluence_space.docs.id
}
```

## Argument Reference

* `key` - (Required) The key of the Confluence space to look up.

## Attributes Reference

* `id` - The numeric ID of the space.
* `key` - The key of the space.
* `name` - The name of the space.
* `type` - The type of the space (e.g., `global`, `personal`).
* `status` - The status of the space (e.g., `current`, `archived`).
* `homepage_id` - The ID of the space's homepage.
* `description` - The plain-text description of the space.
