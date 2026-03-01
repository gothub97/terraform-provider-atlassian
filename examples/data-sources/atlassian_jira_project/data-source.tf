data "atlassian_jira_project" "example" {
  key = "EXAM"
}

output "project_name" {
  value = data.atlassian_jira_project.example.name
}
