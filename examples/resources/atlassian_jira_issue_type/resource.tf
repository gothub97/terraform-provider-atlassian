resource "atlassian_jira_issue_type" "bug" {
  name            = "Custom Bug"
  description     = "A custom bug type"
  hierarchy_level = 0
}
