---
page_title: "atlassian_jira_project Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira project.
---

# atlassian_jira_project (Resource)

Manages a Jira project. Projects are the primary organizational unit in Jira, containing issues, workflows, and configurations.

## Example Usage

### Software Project

```hcl
data "atlassian_jira_myself" "me" {}

resource "atlassian_jira_project" "example" {
  key                  = "EX"
  name                 = "Example Project"
  project_type_key     = "software"
  project_template_key = "com.pyxis.greenhopper.jira:gh-simplified-scrum-classic"
  lead_account_id      = data.atlassian_jira_myself.me.account_id
  description          = "An example software project"
  assignee_type        = "PROJECT_LEAD"
}
```

### Business Project

```hcl
resource "atlassian_jira_project" "business" {
  key              = "BIZ"
  name             = "Business Project"
  project_type_key = "business"
  lead_account_id  = data.atlassian_jira_myself.me.account_id
}
```

## Argument Reference

* `key` - (Required, ForceNew) The project key. Must match the pattern `^[A-Z][A-Z0-9]{1,9}$` (uppercase letters and digits, 2-10 characters, starting with a letter).
* `name` - (Required) The name of the project.
* `project_type_key` - (Required, ForceNew) The type of project. Must be one of: `software`, `business`, `service_desk`.
* `lead_account_id` - (Required) The account ID of the project lead.
* `project_template_key` - (Optional, ForceNew) The project template to use. Example: `com.pyxis.greenhopper.jira:gh-simplified-scrum-classic` for a classic Scrum board.
* `description` - (Optional) The description of the project.
* `assignee_type` - (Optional) The default assignee for new issues. Must be one of: `PROJECT_LEAD`, `UNASSIGNED`. Defaults to `PROJECT_LEAD`.

## Attributes Reference

* `id` - The numeric ID of the project.
* `self` - The URL of the project.

## Import

Import using the project key:

```shell
terraform import atlassian_jira_project.example EX
```
