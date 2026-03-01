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

# Get the current user's account ID for use as project lead
data "atlassian_jira_myself" "current" {}

# Create a software project
resource "atlassian_jira_project" "demo" {
  key              = "TFDEMO"
  name             = "Terraform Demo Project"
  project_type_key = "software"
  lead_account_id  = data.atlassian_jira_myself.current.account_id
  description      = "A demo project fully managed by Terraform"
}

# Create issue types
resource "atlassian_jira_issue_type" "bug" {
  name            = "TF Demo Bug"
  description     = "A bug tracked by Terraform"
  hierarchy_level = 0
}

resource "atlassian_jira_issue_type" "subtask" {
  name            = "TF Demo Subtask"
  description     = "A subtask tracked by Terraform"
  hierarchy_level = -1
}

# Create a custom priority
resource "atlassian_jira_priority" "critical" {
  name         = "TF Demo Critical"
  description  = "Needs immediate attention"
  status_color = "#FF0000"
}

# Create a global status
resource "atlassian_jira_status" "in_review" {
  name            = "TF Demo In Review"
  description     = "Work is being reviewed"
  status_category = "IN_PROGRESS"
  scope_type      = "GLOBAL"
}

# Outputs
output "connected_as" {
  value = data.atlassian_jira_myself.current.display_name
}

output "project_key" {
  value = atlassian_jira_project.demo.key
}

output "bug_issue_type_id" {
  value = atlassian_jira_issue_type.bug.id
}

output "priority_id" {
  value = atlassian_jira_priority.critical.id
}

output "status_id" {
  value = atlassian_jira_status.in_review.id
}
