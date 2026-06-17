data "atlassian_confluence_space" "docs" {
  key = "DOCS"
}

output "space_id" {
  value = data.atlassian_confluence_space.docs.id
}

output "space_homepage_id" {
  value = data.atlassian_confluence_space.docs.homepage_id
}
