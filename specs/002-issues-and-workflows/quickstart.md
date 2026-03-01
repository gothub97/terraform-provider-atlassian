# Quickstart: Issues and Workflows

**Branch**: `002-issues-and-workflows` | **Date**: 2026-03-01

## Prerequisites

- Provider foundation from 001 (project, issue type, priority, status resources)
- Jira Cloud instance with admin permissions
- `ATLASSIAN_URL`, `ATLASSIAN_EMAIL`, `ATLASSIAN_API_TOKEN` environment variables set

## Minimal Example: Custom Field

```hcl
provider "atlassian" {}

# Create a custom field
resource "atlassian_jira_field" "story_points" {
  name         = "Story Points"
  type         = "com.atlassian.jira.plugin.system.customfieldtypes:float"
  description  = "Effort estimate"
  searcher_key = "com.atlassian.jira.plugin.system.customfieldtypes:exactnumber"
}
```

## Full Chain Example

See `examples/002-issues-and-workflows/main.tf` for the complete configuration chain:

```
project → custom fields → field configuration → field configuration scheme
       → screen → screen scheme → issue type screen scheme
       → workflow → workflow scheme
```

## Running Tests

```bash
export PATH="/home/gauthier/.local/go/bin:$HOME/go/bin:$PATH"
source ~/.bashrc

# Run all tests
TF_ACC=1 go test -v -count=1 -timeout 30m ./...

# Run specific resource tests
TF_ACC=1 go test -v -count=1 -run TestAccJiraField ./internal/jira/
TF_ACC=1 go test -v -count=1 -run TestAccJiraWorkflow ./internal/jira/
```

## Implementation Order

1. `jira_field` + `jira_fields` data source (no dependencies)
2. `jira_field_configuration` (depends on fields)
3. `jira_field_configuration_scheme` (depends on field configs)
4. `jira_screen` + `jira_screens` data source (depends on fields)
5. `jira_screen_scheme` (depends on screens)
6. `jira_issue_type_screen_scheme` (depends on screen schemes)
7. `jira_workflow` + `jira_workflows` data source (depends on statuses from 001)
8. `jira_workflow_scheme` (depends on workflows)
9. End-to-end example + documentation
