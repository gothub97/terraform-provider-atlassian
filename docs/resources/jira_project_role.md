---
page_title: "atlassian_jira_project_role Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira project role.
---

# atlassian_jira_project_role (Resource)

Manages a Jira project role. Project roles are used to associate users and groups with specific projects, and can be referenced in permission schemes, notification schemes, and issue security schemes.

## Example Usage

```hcl
resource "atlassian_jira_project_role" "tech_lead" {
  name        = "Tech Lead"
  description = "Technical lead for the project"
}

resource "atlassian_jira_project_role" "qa" {
  name = "QA Engineer"
}
```

## Argument Reference

* `name` - (Required) The name of the project role.
* `description` - (Optional) The description of the project role.

## Attributes Reference

* `id` - The numeric ID of the project role.
* `self` - The URL of the project role.

## Import

Import using the project role ID:

```shell
terraform import atlassian_jira_project_role.example 10360
```
