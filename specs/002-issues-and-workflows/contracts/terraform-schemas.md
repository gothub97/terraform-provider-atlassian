# Terraform Schema Contracts: Issues and Workflows

**Branch**: `002-issues-and-workflows` | **Date**: 2026-03-01

These contracts define the user-facing HCL interface for each resource and data source.

---

## Resources

### atlassian_jira_field

```hcl
resource "atlassian_jira_field" "example" {
  name         = "Story Points"                                                    # Required, string
  type         = "com.atlassian.jira.plugin.system.customfieldtypes:float"         # Required, ForceNew, string
  description  = "Estimated effort in story points"                                # Optional, string
  searcher_key = "com.atlassian.jira.plugin.system.customfieldtypes:exactnumber"   # Optional, string
}

# Computed: id (e.g., "customfield_10001")
# Import: terraform import atlassian_jira_field.example customfield_10001
```

### atlassian_jira_field_configuration

```hcl
resource "atlassian_jira_field_configuration" "example" {
  name        = "Bug Field Configuration"    # Required, string, max 255
  description = "Fields for bug issues"      # Optional, string, max 255

  field_item {                               # Optional, repeatable block
    field_id    = "summary"                  # Required, string
    is_required = true                       # Optional, bool, default false
    is_hidden   = false                      # Optional, bool, default false
    description = "Brief bug summary"        # Optional, string
    renderer    = ""                         # Optional, string
  }

  field_item {
    field_id    = "customfield_10001"
    is_required = false
    is_hidden   = false
  }
}

# Computed: id
# Import: terraform import atlassian_jira_field_configuration.example 10100
```

### atlassian_jira_field_configuration_scheme

```hcl
resource "atlassian_jira_field_configuration_scheme" "example" {
  name        = "Standard Field Config Scheme"    # Required, string, max 255
  description = "Maps issue types to configs"     # Optional, string, max 1024

  mapping {                                       # Optional, repeatable block
    issue_type_id          = "default"            # Required, string ("default" for fallback)
    field_configuration_id = "10100"              # Required, string
  }

  mapping {
    issue_type_id          = "10001"              # Bug issue type
    field_configuration_id = "10101"
  }
}

# Computed: id
# Import: terraform import atlassian_jira_field_configuration_scheme.example 10200
```

### atlassian_jira_screen

```hcl
resource "atlassian_jira_screen" "example" {
  name        = "Bug Screen"                      # Required, string, max 255
  description = "Screen for bug creation"         # Optional, string, max 255

  tab {                                           # Optional, repeatable block (ordered)
    name   = "Details"                            # Required, string, max 255
    fields = ["summary", "description", "priority", "customfield_10001"]  # Optional, list(string), ordered
  }

  tab {
    name   = "Attachments"
    fields = ["attachment"]
  }
}

# Computed: id, tab.*.id
# Import: terraform import atlassian_jira_screen.example 10300
```

### atlassian_jira_screen_scheme

```hcl
resource "atlassian_jira_screen_scheme" "example" {
  name              = "Bug Screen Scheme"         # Required, string
  description       = "Screens for bug workflow"  # Optional, string
  default_screen_id = "10300"                     # Required, string
  create_screen_id  = "10301"                     # Optional, string
  edit_screen_id    = "10302"                     # Optional, string
  view_screen_id    = "10303"                     # Optional, string
}

# Computed: id
# Import: terraform import atlassian_jira_screen_scheme.example 10400
```

### atlassian_jira_issue_type_screen_scheme

```hcl
resource "atlassian_jira_issue_type_screen_scheme" "example" {
  name        = "Standard ITSS"                   # Required, string
  description = "Maps issue types to screens"     # Optional, string

  mapping {                                       # Optional, repeatable block
    issue_type_id    = "default"                  # Required, string ("default" for fallback)
    screen_scheme_id = "10400"                    # Required, string
  }

  mapping {
    issue_type_id    = "10001"                    # Bug
    screen_scheme_id = "10401"
  }
}

# Computed: id
# Import: terraform import atlassian_jira_issue_type_screen_scheme.example 10500
```

### atlassian_jira_workflow

```hcl
resource "atlassian_jira_workflow" "example" {
  name        = "Bug Workflow"                    # Required, string
  description = "Workflow for bug tracking"       # Optional, string

  status {                                        # Required, repeatable block (min 1)
    name             = "Open"                     # Required, string
    status_reference = "open-ref"                 # Required, string (local UUID reference)
    status_category  = "TODO"                     # Required, string: TODO|IN_PROGRESS|DONE
  }

  status {
    name             = "In Progress"
    status_reference = "in-progress-ref"
    status_category  = "IN_PROGRESS"
  }

  status {
    name             = "Done"
    status_reference = "done-ref"
    status_category  = "DONE"
  }

  transition {                                    # Required, repeatable block (min 1)
    name                  = "Create"              # Required, string
    to_status_reference   = "open-ref"            # Required, string
    type                  = "initial"             # Required, string: initial|directed|global
  }

  transition {
    name                  = "Start Work"
    from_status_reference = "open-ref"            # Optional, string (omit for initial/global)
    to_status_reference   = "in-progress-ref"
    type                  = "directed"

    validator {                                   # Optional, repeatable block
      rule_key   = "system:check-permission-validator"
      parameters = {                              # Optional, map(string)
        "permissionKey" = "EDIT_ISSUES"
      }
    }

    condition {                                   # Optional, single block (can be compound)
      operator = "AND"                            # Optional: AND|OR (for compound)

      rule {                                      # Optional, repeatable block (simple conditions)
        rule_key   = "system:restrict-issue-transition"
        parameters = {
          "roleIds" = "10002"
        }
      }
    }

    post_function {                               # Optional, repeatable block
      rule_key   = "system:change-assignee"
      parameters = {
        "type" = "to-current-user"
      }
    }
  }

  transition {
    name                  = "Complete"
    from_status_reference = "in-progress-ref"
    to_status_reference   = "done-ref"
    type                  = "directed"
  }
}

# Computed: id, status.*.status_id
# Import: terraform import atlassian_jira_workflow.example <workflow-uuid>
```

### atlassian_jira_workflow_scheme

```hcl
resource "atlassian_jira_workflow_scheme" "example" {
  name             = "Bug Workflow Scheme"        # Required, string, max 255
  description      = "Assigns bug workflow"       # Optional, string
  default_workflow = "jira"                       # Optional, string, default "jira"

  mapping {                                       # Optional, repeatable block
    issue_type_id = "10001"                       # Required, string
    workflow_name = "Bug Workflow"                 # Required, string
  }
}

# Computed: id
# Import: terraform import atlassian_jira_workflow_scheme.example 10600
# Note: Active schemes (assigned to projects) use draft-based updates transparently
```

### atlassian_jira_issue

```hcl
resource "atlassian_jira_issue" "example" {
  project_key   = "PROJ"                          # Required, ForceNew, string
  issue_type_id = "10001"                         # Required, string
  summary       = "Fix login bug"                 # Required, string
  description   = jsonencode({                    # Optional, string (ADF JSON)
    version = 1
    type    = "doc"
    content = [{
      type    = "paragraph"
      content = [{ type = "text", text = "Detailed description here" }]
    }]
  })
  priority_id      = "3"                          # Optional, string
  status           = "In Progress"                # Optional, string (triggers transitions)
  assignee_id      = "5e68ac137d64450d01a77fa0"   # Optional, string (account ID)
  reporter_id      = "5e68ac137d64450d01a77fa0"   # Optional, string (account ID)
  labels           = ["bug", "critical"]          # Optional, list(string)
  component_ids    = ["10000"]                    # Optional, list(string)
  fix_version_ids  = ["10001"]                    # Optional, list(string)
  custom_fields    = {                            # Optional, map(string)
    "customfield_10001" = "5"                     # Story points (number as string)
    "customfield_10002" = jsonencode({            # Select field
      id = "10100"
    })
  }
}

# Computed: id, key
# Import: terraform import atlassian_jira_issue.example PROJ-123
```

---

## Data Sources

### atlassian_jira_fields

```hcl
data "atlassian_jira_fields" "all" {
  # No required attributes
  type = "custom"                                 # Optional, string: "system"|"custom" (filter)
}

# Output: fields (list of objects with id, name, custom, schema, clause_names)
```

### atlassian_jira_workflows

```hcl
data "atlassian_jira_workflows" "all" {
  # No required attributes
  project_key = "PROJ"                            # Optional, string (filter by project)
}

# Output: workflows (list of objects with id, name, description, statuses, transitions)
```

### atlassian_jira_screens

```hcl
data "atlassian_jira_screens" "all" {
  # No required attributes
}

# Output: screens (list of objects with id, name, description)
```
