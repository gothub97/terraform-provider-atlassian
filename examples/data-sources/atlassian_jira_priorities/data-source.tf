data "atlassian_jira_priorities" "all" {}

output "priority_names" {
  value = [for p in data.atlassian_jira_priorities.all.priorities : p.name]
}
