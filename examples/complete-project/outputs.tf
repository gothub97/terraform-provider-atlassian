output "project_id" {
  value       = atlassian_jira_project.main.id
  description = "The numeric ID of the created Jira project."
}

output "project_key" {
  value       = atlassian_jira_project.main.key
  description = "The project key."
}

output "project_url" {
  value       = atlassian_jira_project.main.self
  description = "The URL of the created Jira project."
}

output "issue_type_id" {
  value       = atlassian_jira_issue_type.task.id
  description = "The ID of the custom issue type."
}

output "workflow_id" {
  value       = atlassian_jira_workflow.main.id
  description = "The entity ID of the custom workflow."
}

output "team_group_id" {
  value       = atlassian_jira_group.team.id
  description = "The ID of the team group."
}

output "custom_fields" {
  value = {
    story_points  = atlassian_jira_field.story_points.id
    sprint_goal   = atlassian_jira_field.sprint_goal.id
    release_notes = atlassian_jira_field.release_notes.id
  }
  description = "Map of custom field names to their IDs."
}

output "permission_scheme_id" {
  value       = atlassian_jira_permission_scheme.main.id
  description = "The ID of the permission scheme."
}

output "notification_scheme_id" {
  value       = atlassian_jira_notification_scheme.main.id
  description = "The ID of the notification scheme."
}

output "security_scheme_id" {
  value       = atlassian_jira_security_scheme.main.id
  description = "The ID of the security scheme."
}
