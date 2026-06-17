data "atlassian_confluence_space" "docs" {
  key = "DOCS"
}

data "atlassian_confluence_space_roles" "docs" {
  space_id = data.atlassian_confluence_space.docs.id
}

output "role_names" {
  value = [for r in data.atlassian_confluence_space_roles.docs.roles : r.name]
}
