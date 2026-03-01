---
page_title: "Atlassian Provider"
subcategory: ""
description: |-
  The Atlassian provider enables Terraform to manage Jira Cloud resources.
---

# Atlassian Provider

The Atlassian provider enables you to manage [Jira Cloud](https://www.atlassian.com/software/jira) resources as infrastructure-as-code. Define projects, fields, workflows, permissions, notifications, and security schemes declaratively with full CRUD support and import capabilities.

## Example Usage

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

# Look up the authenticated user
data "atlassian_jira_myself" "me" {}

# Create a project
resource "atlassian_jira_project" "example" {
  key                  = "EX"
  name                 = "Example Project"
  project_type_key     = "software"
  project_template_key = "com.pyxis.greenhopper.jira:gh-simplified-scrum-classic"
  lead_account_id      = data.atlassian_jira_myself.me.account_id
}
```

## Authentication

The provider authenticates using [Atlassian API tokens](https://id.atlassian.com/manage-profile/security/api-tokens). Credentials can be set in the provider block or via environment variables.

```hcl
provider "atlassian" {
  url       = "https://your-domain.atlassian.net"   # or ATLASSIAN_URL
  email     = "your-email@example.com"              # or ATLASSIAN_EMAIL
  api_token = "your-api-token"                      # or ATLASSIAN_API_TOKEN
}
```

## Schema

### Optional

- `url` (String) Your Jira Cloud instance URL (e.g., `https://your-domain.atlassian.net`). Can also be set with the `ATLASSIAN_URL` environment variable.
- `email` (String) Email address for your Atlassian account. Can also be set with the `ATLASSIAN_EMAIL` environment variable.
- `api_token` (String, Sensitive) API token from your Atlassian account. Can also be set with the `ATLASSIAN_API_TOKEN` environment variable.
