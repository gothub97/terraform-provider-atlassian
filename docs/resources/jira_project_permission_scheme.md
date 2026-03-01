---
page_title: "atlassian_jira_project_permission_scheme Resource - atlassian"
subcategory: "Jira"
description: |-
  Assigns a permission scheme to a Jira project.
---

# atlassian_jira_project_permission_scheme (Resource)

Assigns a permission scheme to a Jira project. When this resource is destroyed, the project reverts to the default permission scheme (ID 0).

## Example Usage

```hcl
resource "atlassian_jira_permission_scheme" "standard" {
  name = "Standard Permission Scheme"
}

resource "atlassian_jira_project_permission_scheme" "example" {
  project_key = "PROJ"
  scheme_id   = atlassian_jira_permission_scheme.standard.id
}
```

## Argument Reference

* `project_key` - (Required) The key of the project.
* `scheme_id` - (Required) The ID of the permission scheme to assign.

## Attributes Reference

No additional attributes are exported.

## Import

Import using the project key:

```shell
terraform import atlassian_jira_project_permission_scheme.example PROJ
```
