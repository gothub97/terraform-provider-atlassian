resource "atlassian_jira_priority" "critical" {
  name         = "Critical"
  description  = "Needs immediate attention"
  status_color = "#FF0000"
}
