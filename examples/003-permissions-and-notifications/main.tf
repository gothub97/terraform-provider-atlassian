terraform {
  required_providers {
    atlassian = {
      source = "atlassian/atlassian"
    }
  }
}

provider "atlassian" {}

# --- Identity Layer ---

data "atlassian_jira_myself" "current" {}

resource "atlassian_jira_group" "developers" {
  name = "example-developers"
}

resource "atlassian_jira_group_membership" "dev_lead" {
  group_id   = atlassian_jira_group.developers.id
  account_id = data.atlassian_jira_myself.current.account_id
}

# --- Project ---

resource "atlassian_jira_project" "example" {
  key              = "EXGOV"
  name             = "Example Governance"
  project_type_key = "software"
  lead_account_id  = data.atlassian_jira_myself.current.account_id
}

# --- Permission Scheme ---

resource "atlassian_jira_permission_scheme" "example" {
  name        = "Example Permission Scheme"
  description = "Managed by Terraform"

  permission {
    permission  = "ADMINISTER_PROJECTS"
    holder_type = "projectLead"
  }

  permission {
    permission  = "BROWSE_PROJECTS"
    holder_type = "group"
    holder_value = atlassian_jira_group.developers.id
  }

  permission {
    permission  = "CREATE_ISSUES"
    holder_type = "anyone"
  }
}

resource "atlassian_jira_project_permission_scheme" "example" {
  project_key = atlassian_jira_project.example.key
  scheme_id   = atlassian_jira_permission_scheme.example.id
}

# --- Notification Scheme ---

resource "atlassian_jira_notification_scheme" "example" {
  name        = "Example Notification Scheme"
  description = "Managed by Terraform"

  notification {
    event_id          = "1" # Issue Created
    notification_type = "CurrentAssignee"
  }

  notification {
    event_id          = "1" # Issue Created
    notification_type = "Reporter"
  }

  notification {
    event_id          = "2" # Issue Updated
    notification_type = "CurrentAssignee"
  }
}

resource "atlassian_jira_project_notification_scheme" "example" {
  project_key = atlassian_jira_project.example.key
  scheme_id   = atlassian_jira_notification_scheme.example.id
}

# --- Security Scheme ---

resource "atlassian_jira_security_scheme" "example" {
  name        = "Example Security Scheme"
  description = "Managed by Terraform"

  level {
    name        = "Confidential"
    description = "Only project leads and reporters"
    is_default  = true

    member {
      type = "reporter"
    }
  }

  level {
    name        = "Internal"
    description = "Group members only"
    is_default  = false

    member {
      type      = "group"
      parameter = atlassian_jira_group.developers.id
    }

    member {
      type = "reporter"
    }
  }
}

# --- Data Sources ---

data "atlassian_jira_groups" "all" {}
data "atlassian_jira_users" "search" {
  query = "."
}
data "atlassian_jira_project_roles" "all" {}
data "atlassian_jira_permission_schemes" "all" {}
data "atlassian_jira_notification_schemes" "all" {}
data "atlassian_jira_security_schemes" "all" {}

# --- Outputs ---

output "group_id" {
  value = atlassian_jira_group.developers.id
}

output "permission_scheme_id" {
  value = atlassian_jira_permission_scheme.example.id
}

output "notification_scheme_id" {
  value = atlassian_jira_notification_scheme.example.id
}

output "security_scheme_id" {
  value = atlassian_jira_security_scheme.example.id
}
