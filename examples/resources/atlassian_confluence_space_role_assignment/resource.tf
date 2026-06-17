resource "atlassian_confluence_space_role_assignment" "docs_team" {
  space_id       = atlassian_confluence_space.docs.id
  principal_type = "GROUP"
  principal_id   = atlassian_jira_group.team.id
  role_id        = "1234-role-id"
}
