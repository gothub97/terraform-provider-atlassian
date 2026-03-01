# Implementation Plan: Permissions and Notifications

## Tech Stack
- Go 1.24.1 + hashicorp/terraform-plugin-framework v1.18.0, hashicorp/terraform-plugin-testing v1.14.0
- Existing HTTP client with retry, pagination, task polling

## Architecture
Same patterns as 001/002: `internal/jira/` for resources/data sources, `internal/atlassian/` for client, `internal/provider/` for registration.

## Implementation Phases

### Phase 1: Groups + Memberships (US1)
- `resource_group.go` — CRUD for jira_group
- `resource_group_membership.go` — CRUD for jira_group_membership
- `data_source_groups.go` — jira_groups data source
- `data_source_users.go` — jira_users data source

### Phase 2: Project Roles + Actors (US2)
- `resource_project_role.go` — CRUD for jira_project_role
- `resource_project_role_actor.go` — CRUD for jira_project_role_actor
- `data_source_project_roles.go` — jira_project_roles data source

### Phase 3: Permission Schemes (US3)
- `resource_permission_scheme.go` — CRUD with grants
- `resource_project_permission_scheme.go` — project association
- `data_source_permission_schemes.go` — list all schemes

### Phase 4: Notification Schemes (US4)
- `resource_notification_scheme.go` — CRUD with event notifications
- `resource_project_notification_scheme.go` — project association
- `data_source_notification_schemes.go` — list all schemes

### Phase 5: Security Schemes (US5)
- `resource_security_scheme.go` — CRUD with levels and members
- `resource_project_security_scheme.go` — project association (async)
- `data_source_security_schemes.go` — list all schemes

### Phase 6: E2E + Polish
- Register all resources/data sources in provider.go
- E2E test, docs, examples
