# Quickstart: Atlassian Terraform Provider

## Prerequisites

1. A Jira Cloud instance (e.g., `https://yoursite.atlassian.net`)
2. An Atlassian account email address
3. An API token generated at https://id.atlassian.com/manage-profile/security/api-tokens
4. Terraform 1.0+

## Provider Configuration

```hcl
terraform {
  required_providers {
    atlassian = {
      source  = "registry.terraform.io/yournamespace/atlassian"
      version = "~> 0.1"
    }
  }
}

provider "atlassian" {
  url       = "https://yoursite.atlassian.net"
  email     = "you@example.com"
  api_token = var.atlassian_api_token  # Never hardcode tokens
}
```

Or via environment variables:

```bash
export ATLASSIAN_URL="https://yoursite.atlassian.net"
export ATLASSIAN_EMAIL="you@example.com"
export ATLASSIAN_API_TOKEN="your-api-token"
```

## Verify Connection

```hcl
data "atlassian_jira_myself" "current" {}

output "connected_as" {
  value = data.atlassian_jira_myself.current.display_name
}
```

```bash
terraform init && terraform plan
# Should output your display name
```

## Create a Project with Issue Types, Priorities, and Statuses

```hcl
resource "atlassian_jira_project" "demo" {
  key              = "DEMO"
  name             = "Demo Project"
  project_type_key = "software"
  lead_account_id  = data.atlassian_jira_myself.current.account_id
  description      = "A demo project managed by Terraform"
}

resource "atlassian_jira_issue_type" "bug" {
  name            = "TF Bug"
  description     = "A bug tracked by Terraform"
  hierarchy_level = 0
}

resource "atlassian_jira_issue_type" "subtask" {
  name            = "TF Subtask"
  description     = "A subtask tracked by Terraform"
  hierarchy_level = -1
}

resource "atlassian_jira_priority" "critical" {
  name         = "TF Critical"
  description  = "Needs immediate attention"
  status_color = "#FF0000"
}

resource "atlassian_jira_status" "in_review" {
  name            = "TF In Review"
  description     = "Work is being reviewed"
  status_category = "IN_PROGRESS"
  scope_type      = "GLOBAL"
}
```

```bash
terraform apply
```

## Import Existing Resources

```bash
terraform import atlassian_jira_project.existing MYKEY
terraform import atlassian_jira_issue_type.existing 10001
terraform import atlassian_jira_priority.existing 10100
terraform import atlassian_jira_status.existing 10200
```

## Clean Up

```bash
terraform destroy
```
