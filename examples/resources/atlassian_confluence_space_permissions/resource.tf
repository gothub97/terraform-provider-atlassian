resource "atlassian_confluence_space_permissions" "docs_team" {
  space_key = atlassian_confluence_space.docs.key
  group_id  = atlassian_jira_group.team.id

  permission {
    operation = "read"
    target    = "space"
  }

  permission {
    operation = "create"
    target    = "page"
  }

  permission {
    operation = "delete"
    target    = "page"
  }
}
