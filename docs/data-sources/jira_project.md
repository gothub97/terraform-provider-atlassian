---
page_title: "atlassian_jira_project Data Source - atlassian"
subcategory: "Jira"
description: |-
  Retrieves information about a Jira project by key.
---

# atlassian_jira_project (Data Source)

Use this data source to look up an existing Jira project by its key. This is useful for referencing projects that are managed outside of Terraform.

## Example Usage

```hcl
data "atlassian_jira_project" "existing" {
  key = "PROJ"
}

output "project_lead" {
  value = data.atlassian_jira_project.existing.lead_account_id
}
```

## Argument Reference

* `key` - (Required) The project key to look up.

## Attributes Reference

* `id` - The numeric ID of the project.
* `name` - The name of the project.
* `project_type_key` - The type of project (`software`, `business`, `service_desk`).
* `lead_account_id` - The account ID of the project lead.
* `description` - The description of the project.
* `assignee_type` - The default assignee type for the project.
* `self` - The URL of the project.
