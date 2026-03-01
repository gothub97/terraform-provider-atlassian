# Terraform Provider for Atlassian Jira

The Atlassian Jira Terraform provider enables you to manage Jira Cloud resources as infrastructure-as-code. Define projects, fields, workflows, permissions, notifications, and security schemes declaratively with full CRUD support.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22 (to build the provider plugin)
- An [Atlassian Cloud](https://www.atlassian.com/software/jira) account with an [API token](https://id.atlassian.com/manage-profile/security/api-tokens)

## Usage

```hcl
terraform {
  required_providers {
    atlassian = {
      source  = "gothub97/atlassian"
      version = "~> 0.1"
    }
  }
}

provider "atlassian" {
  url       = "https://your-domain.atlassian.net"
  email     = "your-email@example.com"
  api_token = "your-api-token"
}
```

## Authentication

The provider requires three configuration values, which can be set via provider block attributes or environment variables:

| Provider Attribute | Environment Variable   | Description                          |
|--------------------|------------------------|--------------------------------------|
| `url`              | `ATLASSIAN_URL`        | Your Jira Cloud instance URL         |
| `email`            | `ATLASSIAN_EMAIL`      | Email for your Atlassian account     |
| `api_token`        | `ATLASSIAN_API_TOKEN`  | API token from Atlassian account     |

## Resources

### Projects and Fields

- `atlassian_jira_project` - Manage Jira projects
- `atlassian_jira_field` - Manage custom fields
- `atlassian_jira_field_configuration` - Manage field configurations
- `atlassian_jira_field_configuration_scheme` - Manage field configuration schemes

### Screens

- `atlassian_jira_screen` - Manage screens with tabs and field layouts
- `atlassian_jira_screen_scheme` - Manage screen schemes
- `atlassian_jira_issue_type_screen_scheme` - Manage issue type screen schemes

### Workflows

- `atlassian_jira_workflow` - Manage workflows with statuses and transitions
- `atlassian_jira_workflow_scheme` - Manage workflow schemes

### Groups and Roles

- `atlassian_jira_group` - Manage groups
- `atlassian_jira_group_membership` - Manage group memberships
- `atlassian_jira_project_role` - Manage project roles
- `atlassian_jira_project_role_actor` - Manage project role actors (user/group)

### Permission Schemes

- `atlassian_jira_permission_scheme` - Manage permission schemes with grants
- `atlassian_jira_project_permission_scheme` - Assign permission schemes to projects

### Notification Schemes

- `atlassian_jira_notification_scheme` - Manage notification schemes
- `atlassian_jira_project_notification_scheme` - Assign notification schemes to projects

### Security Schemes

- `atlassian_jira_security_scheme` - Manage issue security schemes with levels and members
- `atlassian_jira_project_security_scheme` - Assign security schemes to projects

## Data Sources

- `atlassian_jira_myself` - Current authenticated user
- `atlassian_jira_project` - Look up a project
- `atlassian_jira_issue_types` - List issue types
- `atlassian_jira_statuses` - List statuses
- `atlassian_jira_fields` - List fields with optional filters
- `atlassian_jira_screens` - List screens
- `atlassian_jira_workflows` - List workflows
- `atlassian_jira_groups` - List groups
- `atlassian_jira_users` - Search users
- `atlassian_jira_project_roles` - List project roles
- `atlassian_jira_permission_schemes` - List permission schemes
- `atlassian_jira_notification_schemes` - List notification schemes
- `atlassian_jira_security_schemes` - List security schemes

## Example

```hcl
# Create a group and add the current user
data "atlassian_jira_myself" "me" {}

resource "atlassian_jira_group" "dev_team" {
  name = "dev-team"
}

resource "atlassian_jira_group_membership" "dev_lead" {
  group_id   = atlassian_jira_group.dev_team.id
  account_id = data.atlassian_jira_myself.me.account_id
}

# Create a permission scheme
resource "atlassian_jira_permission_scheme" "standard" {
  name        = "Standard Permissions"
  description = "Default permission scheme for projects"

  permission {
    permission  = "ADMINISTER_PROJECTS"
    holder_type = "projectLead"
  }

  permission {
    permission  = "BROWSE_PROJECTS"
    holder_type = "anyone"
  }
}

# Create a project and assign the scheme
resource "atlassian_jira_project" "example" {
  key                  = "EX"
  name                 = "Example Project"
  project_type_key     = "software"
  project_template_key = "com.pyxis.greenhopper.jira:gh-simplified-scrum-classic"
  lead_account_id      = data.atlassian_jira_myself.me.account_id
}

resource "atlassian_jira_project_permission_scheme" "example" {
  project_key = atlassian_jira_project.example.key
  scheme_id   = atlassian_jira_permission_scheme.standard.id
}
```

## Developing the Provider

### Building

```shell
go build ./...
```

### Running Acceptance Tests

Acceptance tests run against a real Jira Cloud instance. Set the required environment variables, then run:

```shell
export ATLASSIAN_URL="https://your-domain.atlassian.net"
export ATLASSIAN_EMAIL="your-email@example.com"
export ATLASSIAN_API_TOKEN="your-api-token"

TF_ACC=1 go test ./internal/jira/ -v -timeout 30m
```

## License

See [LICENSE](LICENSE) for details.
