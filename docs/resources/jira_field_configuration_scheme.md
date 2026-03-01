---
page_title: "atlassian_jira_field_configuration_scheme Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira field configuration scheme.
---

# atlassian_jira_field_configuration_scheme (Resource)

Manages a Jira field configuration scheme. A field configuration scheme maps issue types to field configurations, allowing different issue types to use different field behaviors within the same project.

## Example Usage

```hcl
resource "atlassian_jira_field_configuration" "bug_fields" {
  name = "Bug Field Configuration"
}

resource "atlassian_jira_field_configuration" "story_fields" {
  name = "Story Field Configuration"
}

resource "atlassian_jira_field_configuration_scheme" "example" {
  name        = "My Field Configuration Scheme"
  description = "Maps issue types to field configurations"

  mapping {
    issue_type_id          = "default"
    field_configuration_id = atlassian_jira_field_configuration.bug_fields.id
  }

  mapping {
    issue_type_id          = "10001"
    field_configuration_id = atlassian_jira_field_configuration.story_fields.id
  }
}
```

## Argument Reference

* `name` - (Required) The name of the field configuration scheme. Maximum 255 characters.
* `description` - (Optional) The description of the field configuration scheme. Maximum 1024 characters.

### mapping Block

The `mapping` block defines the association between an issue type and a field configuration. You may specify zero or more `mapping` blocks.

* `issue_type_id` - (Required) The ID of the issue type. Use `"default"` for the default mapping that applies to all issue types not explicitly mapped.
* `field_configuration_id` - (Required) The ID of the field configuration to associate with this issue type.

## Attributes Reference

* `id` - The numeric ID of the field configuration scheme.

## Import

Import using the field configuration scheme ID:

```shell
terraform import atlassian_jira_field_configuration_scheme.example 10200
```
