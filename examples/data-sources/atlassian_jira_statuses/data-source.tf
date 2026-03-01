data "atlassian_jira_statuses" "all" {}

output "status_names" {
  value = [for s in data.atlassian_jira_statuses.all.statuses : s.name]
}
