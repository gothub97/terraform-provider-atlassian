resource "atlassian_jira_project" "example" {
  key              = "EXAM"
  name             = "Example Project"
  project_type_key = "software"
  lead_account_id  = data.atlassian_jira_myself.current.account_id
  description      = "An example project managed by Terraform"
}
