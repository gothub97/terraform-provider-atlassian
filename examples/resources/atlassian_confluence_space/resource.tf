resource "atlassian_confluence_space" "docs" {
  key         = "DOCS"
  name        = "Documentation"
  description = "Team documentation space managed by Terraform"
}
