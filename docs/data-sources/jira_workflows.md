---
page_title: "atlassian_jira_workflows Data Source - atlassian"
subcategory: "Jira"
description: |-
  Retrieves all Jira workflows, optionally filtered by project.
---

# atlassian_jira_workflows (Data Source)

Use this data source to retrieve all Jira workflows, optionally filtered by project. Workflows define the set of statuses and transitions that an issue moves through during its lifecycle. This data source is useful for discovering workflow IDs and names needed when configuring workflow schemes.

## Example Usage

### Retrieve All Workflows

```hcl
data "atlassian_jira_workflows" "all" {
}
```

### Retrieve Workflows for a Specific Project

```hcl
data "atlassian_jira_workflows" "project" {
  project_key = "MYPROJ"
}
```

## Argument Reference

* `project_key` - (Optional) A project key to filter workflows associated with the project. When set, only workflows linked to the specified project are returned.

## Attributes Reference

* `workflows` - The list of workflows. Each workflow has the following attributes:
  * `id` - The entity ID (UUID) of the workflow.
  * `name` - The name of the workflow.
  * `description` - The description of the workflow.
