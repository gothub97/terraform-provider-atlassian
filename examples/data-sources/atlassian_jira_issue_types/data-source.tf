data "atlassian_jira_issue_types" "all" {}

output "issue_type_names" {
  value = [for it in data.atlassian_jira_issue_types.all.issue_types : it.name]
}
