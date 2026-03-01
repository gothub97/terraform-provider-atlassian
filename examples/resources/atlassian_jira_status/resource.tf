resource "atlassian_jira_status" "in_review" {
  name            = "In Review"
  description     = "Work is being reviewed"
  status_category = "IN_PROGRESS"
  scope_type      = "GLOBAL"
}
