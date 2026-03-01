---
page_title: "atlassian_jira_workflow Resource - atlassian"
subcategory: "Jira"
description: |-
  Manages a Jira workflow.
---

# atlassian_jira_workflow (Resource)

Manages a Jira workflow. Workflows define the set of statuses and transitions that an issue moves through during its lifecycle. Each workflow consists of statuses (representing states) and transitions (representing the allowed movements between states).

## Example Usage

### Basic Workflow

```hcl
resource "atlassian_jira_workflow" "simple" {
  name        = "Simple Workflow"
  description = "A basic three-status workflow"

  status {
    name             = "To Do"
    status_reference = "todo"
    status_category  = "TODO"
  }

  status {
    name             = "In Progress"
    status_reference = "inprogress"
    status_category  = "IN_PROGRESS"
  }

  status {
    name             = "Done"
    status_reference = "done"
    status_category  = "DONE"
  }

  transition {
    name                  = "Create"
    to_status_reference   = "todo"
    type                  = "initial"
  }

  transition {
    name                  = "Start Work"
    from_status_reference = "todo"
    to_status_reference   = "inprogress"
    type                  = "directed"
  }

  transition {
    name                  = "Complete"
    from_status_reference = "inprogress"
    to_status_reference   = "done"
    type                  = "directed"
  }
}
```

### Workflow with Validators and Conditions

```hcl
resource "atlassian_jira_workflow" "advanced" {
  name        = "Advanced Workflow"
  description = "A workflow with validators and conditions"

  status {
    name             = "Open"
    status_reference = "open"
    status_category  = "TODO"
  }

  status {
    name             = "In Review"
    status_reference = "review"
    status_category  = "IN_PROGRESS"
  }

  status {
    name             = "Closed"
    status_reference = "closed"
    status_category  = "DONE"
  }

  transition {
    name                = "Create"
    to_status_reference = "open"
    type                = "initial"
  }

  transition {
    name                  = "Submit for Review"
    from_status_reference = "open"
    to_status_reference   = "review"
    type                  = "directed"

    validator {
      rule_key = "system:validate-field-value"
      parameters = {
        fieldKey  = "description"
        errorMessage = "Description is required before review"
      }
    }

    condition {
      operator = "AND"

      rule {
        rule_key = "system:check-permission"
        parameters = {
          permissionKey = "EDIT_ISSUES"
        }
      }
    }

    post_function {
      rule_key = "system:assign-to-current-user"
    }
  }

  transition {
    name                  = "Close"
    from_status_reference = "review"
    to_status_reference   = "closed"
    type                  = "directed"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the workflow.
* `description` - (Optional) The description of the workflow.

### status Block

The `status` block defines the statuses used in the workflow. At least one status is required.

* `name` - (Required) The name of the status.
* `status_reference` - (Required) A local reference string used to link statuses and transitions within the workflow definition.
* `status_category` - (Required) The category of the status. Must be one of: `TODO`, `IN_PROGRESS`, `DONE`.

### transition Block

The `transition` block defines the transitions between statuses. At least one transition is required.

* `name` - (Required) The name of the transition.
* `to_status_reference` - (Required) The status reference the transition goes to.
* `type` - (Required) The type of transition. Must be one of: `initial`, `directed`, `global`.
* `from_status_reference` - (Optional) The status reference the transition originates from. Omit for `initial` or `global` transitions.

#### validator Block (nested inside transition)

The `validator` block defines validators for the transition. You may specify zero or more `validator` blocks.

* `rule_key` - (Required) The rule key for the validator.
* `parameters` - (Optional) A map of string parameters for the validator rule.

#### condition Block (nested inside transition)

The `condition` block defines the condition for the transition. At most one `condition` block is allowed per transition.

* `operator` - (Required) The logical operator. Must be one of: `AND`, `OR`.

##### rule Block (nested inside condition)

The `rule` block defines individual condition rules. You may specify zero or more `rule` blocks within a `condition`.

* `rule_key` - (Required) The rule key for the condition.
* `parameters` - (Optional) A map of string parameters for the condition rule.

#### post_function Block (nested inside transition)

The `post_function` block defines post-functions for the transition. You may specify zero or more `post_function` blocks.

* `rule_key` - (Required) The rule key for the post-function.
* `parameters` - (Optional) A map of string parameters for the post-function rule.

## Attributes Reference

* `id` - The entity ID (UUID) of the workflow.
* `status.status_id` - The Jira status ID, resolved after creation.

## Import

Import using the workflow entity ID (UUID):

```shell
terraform import atlassian_jira_workflow.example 12345678-1234-1234-1234-123456789abc
```
