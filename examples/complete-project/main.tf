# =============================================================================
# Complete Jira Project — all 22 resource types
#
# This example provisions a fully-configured Jira project from scratch:
#   - Custom issue type, priority, and statuses
#   - Custom fields with field configuration and scheme
#   - Screens (create/edit/view) with screen scheme and issue type screen scheme
#   - 4-status workflow (Open → In Progress → In Review → Done) with scheme
#   - Team group with membership and project role with actor
#   - Permission, notification, and security schemes assigned to the project
# =============================================================================

# -----------------------------------------------------------------------------
# Data Sources
# -----------------------------------------------------------------------------

data "atlassian_jira_myself" "current" {}

# -----------------------------------------------------------------------------
# Foundation: Issue Type, Priority, Custom Fields, Group, Project Role
# -----------------------------------------------------------------------------

resource "atlassian_jira_issue_type" "task" {
  name            = "${var.project_key} Task"
  description     = "Custom task type for ${var.project_name}"
  hierarchy_level = 0
}

resource "atlassian_jira_priority" "high" {
  name         = "${var.project_key} High"
  description  = "High priority for ${var.project_name}"
  status_color = "#FF5630"
  icon_url     = "https://hubertgauthier5.atlassian.net/images/icons/priorities/highest_new.svg"
}

resource "atlassian_jira_field" "story_points" {
  name         = "${var.project_key} Story Points"
  type         = "com.atlassian.jira.plugin.system.customfieldtypes:float"
  description  = "Estimated story points"
  searcher_key = "com.atlassian.jira.plugin.system.customfieldtypes:exactnumber"
}

resource "atlassian_jira_field" "sprint_goal" {
  name         = "${var.project_key} Sprint Goal"
  type         = "com.atlassian.jira.plugin.system.customfieldtypes:textfield"
  description  = "Goal for the current sprint"
  searcher_key = "com.atlassian.jira.plugin.system.customfieldtypes:textsearcher"
}

resource "atlassian_jira_field" "release_notes" {
  name         = "${var.project_key} Release Notes"
  type         = "com.atlassian.jira.plugin.system.customfieldtypes:textarea"
  description  = "Release notes entry for this issue"
  searcher_key = "com.atlassian.jira.plugin.system.customfieldtypes:textsearcher"
}

resource "atlassian_jira_group" "team" {
  name = var.team_name
}

resource "atlassian_jira_project_role" "developer" {
  name        = "${var.project_key} Developer"
  description = "Developer role for ${var.project_name}"
}

# -----------------------------------------------------------------------------
# Membership: Add current user to team group
# -----------------------------------------------------------------------------

resource "atlassian_jira_group_membership" "team_lead" {
  group_id   = atlassian_jira_group.team.id
  account_id = data.atlassian_jira_myself.current.account_id
}

# -----------------------------------------------------------------------------
# Field Configuration: configure how fields behave
# -----------------------------------------------------------------------------

resource "atlassian_jira_field_configuration" "main" {
  name        = "${var.project_key} Field Configuration"
  description = "Field configuration for ${var.project_name}"

  field_item {
    field_id    = "summary"
    is_required = true
  }

  field_item {
    field_id    = atlassian_jira_field.story_points.id
    description = "Story point estimate"
  }

  field_item {
    field_id    = atlassian_jira_field.sprint_goal.id
    description = "Sprint goal"
  }

  field_item {
    field_id    = atlassian_jira_field.release_notes.id
    description = "Release notes entry"
  }
}

resource "atlassian_jira_field_configuration_scheme" "main" {
  name        = "${var.project_key} Field Config Scheme"
  description = "Maps issue types to field configurations for ${var.project_name}"

  mapping {
    issue_type_id          = "default"
    field_configuration_id = atlassian_jira_field_configuration.main.id
  }
}

# -----------------------------------------------------------------------------
# Screens: separate create, edit, and view screens
# -----------------------------------------------------------------------------

resource "atlassian_jira_screen" "create" {
  name        = "${var.project_key} Create Screen"
  description = "Create screen for ${var.project_name}"

  tab {
    name = "Details"
    fields = [
      "summary",
      "description",
      atlassian_jira_field.story_points.id,
      atlassian_jira_field.sprint_goal.id,
    ]
  }
}

resource "atlassian_jira_screen" "edit" {
  name        = "${var.project_key} Edit Screen"
  description = "Edit screen for ${var.project_name}"

  tab {
    name = "Details"
    fields = [
      "summary",
      "description",
      atlassian_jira_field.story_points.id,
      atlassian_jira_field.sprint_goal.id,
      atlassian_jira_field.release_notes.id,
    ]
  }
}

resource "atlassian_jira_screen" "view" {
  name        = "${var.project_key} View Screen"
  description = "View screen for ${var.project_name}"

  tab {
    name = "Details"
    fields = [
      "summary",
      "description",
      atlassian_jira_field.story_points.id,
      atlassian_jira_field.release_notes.id,
    ]
  }
}

resource "atlassian_jira_screen_scheme" "main" {
  name              = "${var.project_key} Screen Scheme"
  description       = "Screen scheme for ${var.project_name}"
  default_screen_id = atlassian_jira_screen.view.id
  create_screen_id  = atlassian_jira_screen.create.id
  edit_screen_id    = atlassian_jira_screen.edit.id
  view_screen_id    = atlassian_jira_screen.view.id
}

resource "atlassian_jira_issue_type_screen_scheme" "main" {
  name        = "${var.project_key} Issue Type Screen Scheme"
  description = "Issue type screen scheme for ${var.project_name}"

  mapping {
    issue_type_id    = "default"
    screen_scheme_id = atlassian_jira_screen_scheme.main.id
  }
}

# -----------------------------------------------------------------------------
# Statuses + Workflow: 4-status flow with transitions
# -----------------------------------------------------------------------------

# Standalone status resources (demonstrate the atlassian_jira_status resource).
# These have distinct names from the workflow's inline statuses because the
# workflow API creates its own statuses and would conflict on duplicate names.

resource "atlassian_jira_status" "open" {
  name            = "${var.project_key} Backlog"
  description     = "Issue is in the backlog"
  status_category = "TODO"
  scope_type      = "GLOBAL"
}

resource "atlassian_jira_status" "in_progress" {
  name            = "${var.project_key} Active"
  description     = "Issue is actively being worked on"
  status_category = "IN_PROGRESS"
  scope_type      = "GLOBAL"
  depends_on      = [atlassian_jira_status.open]
}

resource "atlassian_jira_status" "in_review" {
  name            = "${var.project_key} Reviewing"
  description     = "Issue is under review"
  status_category = "IN_PROGRESS"
  scope_type      = "GLOBAL"
  depends_on      = [atlassian_jira_status.in_progress]
}

resource "atlassian_jira_status" "done" {
  name            = "${var.project_key} Closed"
  description     = "Issue is complete"
  status_category = "DONE"
  scope_type      = "GLOBAL"
  depends_on      = [atlassian_jira_status.in_review]
}

# Workflow with its own inline statuses (distinct from the standalone ones above).

resource "atlassian_jira_workflow" "main" {
  name        = "${var.project_key} Workflow"
  description = "4-status workflow: Open → In Progress → In Review → Done"

  depends_on = [atlassian_jira_status.done]

  status {
    name             = "${var.project_key} Open"
    status_reference = "open"
    status_category  = "TODO"
  }

  status {
    name             = "${var.project_key} In Progress"
    status_reference = "in_progress"
    status_category  = "IN_PROGRESS"
  }

  status {
    name             = "${var.project_key} In Review"
    status_reference = "in_review"
    status_category  = "IN_PROGRESS"
  }

  status {
    name             = "${var.project_key} Done"
    status_reference = "done"
    status_category  = "DONE"
  }

  transition {
    name                = "Create"
    to_status_reference = "open"
    type                = "initial"
  }

  transition {
    name                  = "Start Progress"
    from_status_reference = "open"
    to_status_reference   = "in_progress"
    type                  = "directed"
  }

  transition {
    name                  = "Submit for Review"
    from_status_reference = "in_progress"
    to_status_reference   = "in_review"
    type                  = "directed"
  }

  transition {
    name                  = "Approve"
    from_status_reference = "in_review"
    to_status_reference   = "done"
    type                  = "directed"
  }

  transition {
    name                  = "Reject"
    from_status_reference = "in_review"
    to_status_reference   = "in_progress"
    type                  = "directed"
  }
}

resource "atlassian_jira_workflow_scheme" "main" {
  name             = "${var.project_key} Workflow Scheme"
  description      = "Workflow scheme for ${var.project_name}"
  default_workflow = atlassian_jira_workflow.main.name

  mapping {
    issue_type_id = atlassian_jira_issue_type.task.id
    workflow_name = atlassian_jira_workflow.main.name
  }
}

# -----------------------------------------------------------------------------
# Project: classic (company-managed) software project
# -----------------------------------------------------------------------------

resource "atlassian_jira_project" "main" {
  key                  = var.project_key
  name                 = var.project_name
  project_type_key     = "software"
  project_template_key = "com.pyxis.greenhopper.jira:gh-simplified-scrum-classic"
  lead_account_id      = data.atlassian_jira_myself.current.account_id
  description          = "A complete Jira project fully managed by Terraform"
  assignee_type        = "PROJECT_LEAD"

  # Ensure project is destroyed before governance schemes so scheme
  # associations are removed with the project, allowing scheme deletion.
  depends_on = [
    atlassian_jira_permission_scheme.main,
    atlassian_jira_notification_scheme.main,
    atlassian_jira_security_scheme.main,
  ]
}

# -----------------------------------------------------------------------------
# Governance: Permission, Notification, and Security Schemes
# -----------------------------------------------------------------------------

resource "atlassian_jira_permission_scheme" "main" {
  name        = "${var.project_key} Permission Scheme"
  description = "Permission scheme for ${var.project_name}"

  permission {
    permission  = "ADMINISTER_PROJECTS"
    holder_type = "projectLead"
  }

  permission {
    permission   = "BROWSE_PROJECTS"
    holder_type  = "projectRole"
    holder_value = atlassian_jira_project_role.developer.id
  }

  permission {
    permission   = "CREATE_ISSUES"
    holder_type  = "projectRole"
    holder_value = atlassian_jira_project_role.developer.id
  }

  permission {
    permission   = "EDIT_ISSUES"
    holder_type  = "projectRole"
    holder_value = atlassian_jira_project_role.developer.id
  }

  permission {
    permission  = "ASSIGN_ISSUES"
    holder_type = "projectLead"
  }
}

resource "atlassian_jira_project_permission_scheme" "main" {
  project_key = atlassian_jira_project.main.key
  scheme_id   = atlassian_jira_permission_scheme.main.id
}

resource "atlassian_jira_notification_scheme" "main" {
  name        = "${var.project_key} Notification Scheme"
  description = "Notification scheme for ${var.project_name}"

  notification {
    event_id          = "1"
    notification_type = "CurrentAssignee"
  }

  notification {
    event_id          = "1"
    notification_type = "ProjectRole"
    parameter         = atlassian_jira_project_role.developer.id
  }

  notification {
    event_id          = "2"
    notification_type = "CurrentAssignee"
  }

  notification {
    event_id          = "3"
    notification_type = "CurrentAssignee"
  }

  notification {
    event_id          = "5"
    notification_type = "Reporter"
  }
}

resource "atlassian_jira_project_notification_scheme" "main" {
  project_key = atlassian_jira_project.main.key
  scheme_id   = atlassian_jira_notification_scheme.main.id
}

resource "atlassian_jira_security_scheme" "main" {
  name        = "${var.project_key} Security Scheme"
  description = "Security scheme for ${var.project_name}"

  level {
    name        = "Internal"
    description = "Visible to team members"
    is_default  = true

    member {
      type      = "projectrole"
      parameter = atlassian_jira_project_role.developer.id
    }
  }

  level {
    name        = "Confidential"
    description = "Visible to reporter and project lead only"

    member {
      type = "reporter"
    }

    member {
      type      = "user"
      parameter = data.atlassian_jira_myself.current.account_id
    }
  }
}

resource "atlassian_jira_project_security_scheme" "main" {
  project_key = atlassian_jira_project.main.key
  scheme_id   = atlassian_jira_security_scheme.main.id
}

# -----------------------------------------------------------------------------
# Project Role Actor: add the team group to the developer role
# -----------------------------------------------------------------------------

resource "atlassian_jira_project_role_actor" "team" {
  project_key = atlassian_jira_project.main.key
  role_id     = atlassian_jira_project_role.developer.id
  actor_type  = "group"
  actor_value = atlassian_jira_group.team.id
}
