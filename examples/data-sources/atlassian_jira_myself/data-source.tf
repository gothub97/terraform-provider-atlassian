data "atlassian_jira_myself" "current" {}

output "my_account_id" {
  value = data.atlassian_jira_myself.current.account_id
}

output "my_display_name" {
  value = data.atlassian_jira_myself.current.display_name
}
