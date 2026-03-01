---
page_title: "atlassian_jira_field_configuration Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira field configuration.
---

# atlassian_jira_field_configuration (Resource)

Manages a Jira field configuration. Field configurations define the behavior of fields (visibility, requirement, description, and renderer) within a specific context.

## Example Usage

```hcl
resource "atlassian_jira_field_configuration" "bug_fields" {
  name        = "Bug Field Configuration"
  description = "Field configuration for bug issue types"

  field_item {
    field_id    = "summary"
    is_required = true
    description = "A brief summary of the bug"
  }

  field_item {
    field_id    = "description"
    is_required = true
    is_hidden   = false
    renderer    = "wiki-renderer"
    description = "Detailed description of the bug"
  }

  field_item {
    field_id    = "priority"
    is_required = true
    description = "The priority of the bug"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the field configuration. Maximum 255 characters.
* `description` - (Optional) The description of the field configuration. Maximum 255 characters.

### field_item Block

The `field_item` block defines the configuration for individual fields. You may specify zero or more `field_item` blocks.

* `field_id` - (Required) The ID of the field (e.g., `summary`, `customfield_10001`).
* `is_required` - (Optional) Whether the field is required. Defaults to `false`.
* `is_hidden` - (Optional) Whether the field is hidden. Defaults to `false`.
* `description` - (Optional) The description of the field in this configuration.
* `renderer` - (Optional) The renderer type for the field.

## Attributes Reference

* `id` - The numeric ID of the field configuration.

## Import

Import using the field configuration ID:

```shell
terraform import atlassian_jira_field_configuration.example 10100
```
