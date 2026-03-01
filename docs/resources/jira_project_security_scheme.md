---
page_title: "atlassian_jira_project_security_scheme Resource - atlassian"
subcategory: "Jira"
description: |-
  Assigns an issue security scheme to a Jira project.
---

# atlassian_jira_project_security_scheme (Resource)

Assigns an issue security scheme to a Jira project. This operation is asynchronous -- the provider waits for the assignment task to complete before returning. When this resource is destroyed, the security scheme is removed from the project.

## Example Usage

```hcl
resource "atlassian_jira_security_scheme" "confidential" {
  name = "Confidential Security Scheme"
}

resource "atlassian_jira_project_security_scheme" "example" {
  project_key = "PROJ"
  scheme_id   = atlassian_jira_security_scheme.confidential.id
}
```

## Argument Reference

* `project_key` - (Required) The key of the project.
* `scheme_id` - (Required) The ID of the issue security scheme to assign.

## Attributes Reference

No additional attributes are exported.

## Import

Import using the project key:

```shell
terraform import atlassian_jira_project_security_scheme.example PROJ
```
