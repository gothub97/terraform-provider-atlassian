---
page_title: "atlassian_jira_workflow_scheme Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira workflow scheme.
---

# atlassian_jira_workflow_scheme (Resource)

Manages a Jira workflow scheme. A workflow scheme maps issue types to workflows, determining which workflow governs the lifecycle of each issue type within a project.

When a workflow scheme is active (assigned to one or more projects), updates are applied using Jira's draft/publish flow. The provider handles this automatically.

## Example Usage

```hcl
resource "atlassian_jira_workflow" "bug_workflow" {
  name = "Bug Workflow"
  # ... statuses and transitions ...
}

resource "atlassian_jira_workflow_scheme" "example" {
  name             = "My Workflow Scheme"
  description      = "Custom workflow scheme for the project"
  default_workflow = "jira"

  mapping {
    issue_type_id = "10001"
    workflow_name = atlassian_jira_workflow.bug_workflow.name
  }

  mapping {
    issue_type_id = "10002"
    workflow_name = "jira"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the workflow scheme.
* `description` - (Optional) The description of the workflow scheme.
* `default_workflow` - (Optional) The name of the default workflow. Defaults to `"jira"` if not specified. This workflow is used for any issue type that is not explicitly mapped.

### mapping Block

The `mapping` block defines the association between an issue type and a workflow. You may specify zero or more `mapping` blocks.

* `issue_type_id` - (Required) The issue type ID.
* `workflow_name` - (Required) The workflow name to assign to this issue type.

## Attributes Reference

* `id` - The numeric ID of the workflow scheme.
* `default_workflow` - The name of the default workflow (computed if not explicitly set).

## Import

Import using the workflow scheme ID:

```shell
terraform import atlassian_jira_workflow_scheme.example 10600
```
