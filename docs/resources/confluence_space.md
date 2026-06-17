---
page_title: "atlassian_confluence_space Resource - atlassian"
subcategory: "Confluence"
description: |-
  Manages a Confluence space.
---

# atlassian_confluence_space (Resource)

Manages a Confluence space. Spaces are created and read through the Confluence v2 REST API, while updates and deletes use the v1 REST API (the v2 API does not expose those operations).

## Example Usage

```hcl
resource "atlassian_confluence_space" "docs" {
  key         = "DOCS"
  name        = "Documentation"
  description = "Team documentation space managed by Terraform"
}
```

## Argument Reference

* `key` - (Required) The key of the space. Changing this forces a new space to be created.
* `name` - (Required) The name of the space.
* `description` - (Optional) The plain-text description of the space.

## Attributes Reference

* `id` - The numeric ID of the space.
* `type` - The type of the space (e.g., `global`).
* `status` - The status of the space (`current` or `archived`).
* `homepage_id` - The ID of the space's homepage.
* `url` - The URL of the space.

## Import

Import using the space key:

```shell
terraform import atlassian_confluence_space.example DOCS
```
