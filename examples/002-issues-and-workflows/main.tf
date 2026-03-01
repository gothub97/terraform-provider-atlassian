terraform {
  required_providers {
    atlassian = {
      source = "registry.terraform.io/atlassian/atlassian"
    }
  }
}

provider "atlassian" {
  # Configure via environment variables:
  #   ATLASSIAN_URL, ATLASSIAN_EMAIL, ATLASSIAN_API_TOKEN
}

# =============================================================================
# Data Sources
# =============================================================================

data "atlassian_jira_myself" "current" {}

data "atlassian_jira_issue_types" "all" {}

# =============================================================================
# Project
# =============================================================================

resource "atlassian_jira_project" "demo" {
  key              = "TFWF"
  name             = "TF Workflow Demo"
  project_type_key = "software"
  lead_account_id  = data.atlassian_jira_myself.current.account_id
  description      = "Demonstrates the full Jira configuration chain via Terraform"
}

# =============================================================================
# Custom Fields
# =============================================================================

resource "atlassian_jira_field" "severity" {
  name        = "TF Demo Severity"
  type        = "com.atlassian.jira.plugin.system.customfieldtypes:select"
  description = "Severity level for the issue"
  searcher_key = "com.atlassian.jira.plugin.system.customfieldtypes:multiselectsearcher"
}

resource "atlassian_jira_field" "story_points" {
  name        = "TF Demo Story Points"
  type        = "com.atlassian.jira.plugin.system.customfieldtypes:float"
  description = "Effort estimate in story points"
  searcher_key = "com.atlassian.jira.plugin.system.customfieldtypes:exactnumber"
}

# =============================================================================
# Field Configuration
# =============================================================================

resource "atlassian_jira_field_configuration" "demo" {
  name        = "TF Demo Field Configuration"
  description = "Custom field configuration for workflow demo"

  field_item {
    field_id    = "summary"
    is_required = true
  }

  field_item {
    field_id    = "description"
    is_required = false
  }
}

# =============================================================================
# Field Configuration Scheme
# =============================================================================

resource "atlassian_jira_field_configuration_scheme" "demo" {
  name        = "TF Demo Field Config Scheme"
  description = "Maps issue types to field configurations"

  mapping {
    issue_type_id          = "default"
    field_configuration_id = atlassian_jira_field_configuration.demo.id
  }
}

# =============================================================================
# Screens
# =============================================================================

resource "atlassian_jira_screen" "create" {
  name        = "TF Demo Create Screen"
  description = "Screen shown when creating issues"

  tab {
    name   = "Details"
    fields = ["summary", "description", "priority"]
  }
}

resource "atlassian_jira_screen" "edit" {
  name        = "TF Demo Edit Screen"
  description = "Screen shown when editing issues"

  tab {
    name   = "Details"
    fields = ["summary", "description", "priority"]
  }
}

resource "atlassian_jira_screen" "view" {
  name        = "TF Demo View Screen"
  description = "Screen shown when viewing issues"

  tab {
    name   = "Details"
    fields = ["summary", "description", "priority"]
  }
}

# =============================================================================
# Screen Scheme
# =============================================================================

resource "atlassian_jira_screen_scheme" "demo" {
  name              = "TF Demo Screen Scheme"
  description       = "Maps operations to screens"
  default_screen_id = atlassian_jira_screen.view.id
  create_screen_id  = atlassian_jira_screen.create.id
  edit_screen_id    = atlassian_jira_screen.edit.id
  view_screen_id    = atlassian_jira_screen.view.id
}

# =============================================================================
# Issue Type Screen Scheme
# =============================================================================

resource "atlassian_jira_issue_type_screen_scheme" "demo" {
  name        = "TF Demo ITSS"
  description = "Maps issue types to screen schemes"

  mapping {
    issue_type_id    = "default"
    screen_scheme_id = atlassian_jira_screen_scheme.demo.id
  }
}

# =============================================================================
# Workflow
# =============================================================================

resource "atlassian_jira_workflow" "demo" {
  name        = "TF Demo Workflow"
  description = "A simple three-state workflow managed by Terraform"

  status {
    name             = "Open"
    status_reference = "open"
    status_category  = "TODO"
  }

  status {
    name             = "In Progress"
    status_reference = "in_progress"
    status_category  = "IN_PROGRESS"
  }

  status {
    name             = "Done"
    status_reference = "done"
    status_category  = "DONE"
  }

  transition {
    name                = "Create"
    type                = "initial"
    to_status_reference = "open"
  }

  transition {
    name                  = "Start Work"
    type                  = "directed"
    from_status_reference = "open"
    to_status_reference   = "in_progress"
  }

  transition {
    name                  = "Complete"
    type                  = "directed"
    from_status_reference = "in_progress"
    to_status_reference   = "done"
  }

  transition {
    name                  = "Reopen"
    type                  = "directed"
    from_status_reference = "done"
    to_status_reference   = "open"
  }
}

# =============================================================================
# Workflow Scheme
# =============================================================================

resource "atlassian_jira_workflow_scheme" "demo" {
  name             = "TF Demo Workflow Scheme"
  description      = "Maps issue types to workflows"
  default_workflow = atlassian_jira_workflow.demo.name
}

# =============================================================================
# Data Source Queries
# =============================================================================

data "atlassian_jira_fields" "custom" {
  type = "custom"

  depends_on = [
    atlassian_jira_field.severity,
    atlassian_jira_field.story_points,
  ]
}

data "atlassian_jira_screens" "all" {
  depends_on = [
    atlassian_jira_screen.create,
    atlassian_jira_screen.edit,
    atlassian_jira_screen.view,
  ]
}

data "atlassian_jira_workflows" "all" {
  depends_on = [atlassian_jira_workflow.demo]
}

# =============================================================================
# Outputs
# =============================================================================

output "connected_as" {
  value = data.atlassian_jira_myself.current.display_name
}

output "project_key" {
  value = atlassian_jira_project.demo.key
}

output "custom_field_count" {
  value = length(data.atlassian_jira_fields.custom.fields)
}

output "screen_count" {
  value = length(data.atlassian_jira_screens.all.screens)
}

output "workflow_count" {
  value = length(data.atlassian_jira_workflows.all.workflows)
}
