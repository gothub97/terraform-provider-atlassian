provider "atlassian" {}

# --- Identity Layer ---
# Groups are org-wide in Atlassian and are managed through the Jira API, but the
# same group can be granted permissions and roles on a Confluence space.

resource "atlassian_jira_group" "team" {
  name = "confluence-docs-team"
}

# --- Confluence Space ---

resource "atlassian_confluence_space" "docs" {
  key         = "DOCS"
  name        = "Documentation"
  description = "Team documentation space managed by Terraform"
}

# --- Granular Space Permissions (v1 API) ---
# Bind the team group to the space with a set of operation/target grants.

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

# --- Space Roles (v2 API) ---
# Discover the roles available for the space so we can reference one by name.

data "atlassian_confluence_space_roles" "docs" {
  space_id = atlassian_confluence_space.docs.id
}

locals {
  # Pick the role named "Admin" if present, otherwise the first available role.
  admin_role_id = try(
    [for r in data.atlassian_confluence_space_roles.docs.roles : r.id if r.name == "Admin"][0],
    data.atlassian_confluence_space_roles.docs.roles[0].id,
  )
}

# --- Space Role Assignment (v2 API) ---

resource "atlassian_confluence_space_role_assignment" "docs_team" {
  space_id       = atlassian_confluence_space.docs.id
  principal_type = "GROUP"
  principal_id   = atlassian_jira_group.team.id
  role_id        = local.admin_role_id
}

# --- Look up an existing space by key ---

data "atlassian_confluence_space" "docs" {
  key = atlassian_confluence_space.docs.key
}

output "space_id" {
  value = data.atlassian_confluence_space.docs.id
}

output "space_homepage_id" {
  value = data.atlassian_confluence_space.docs.homepage_id
}
