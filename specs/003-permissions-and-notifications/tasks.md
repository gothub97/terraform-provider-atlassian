# Task Breakdown: Permissions and Notifications

## Phase 1: Groups and Memberships

- [x] T001: Implement `jira_group` resource (resource_group.go) — DONE, needs acc test
- [x] T002: Implement `jira_group_membership` resource (resource_group_membership.go) — DONE, needs acc test
- [x] T003: Implement `jira_groups` data source (data_source_groups.go) — DONE, needs acc test
- [x] T004: Implement `jira_users` data source (data_source_users.go) — DONE, needs acc test

## Phase 2: Project Roles and Actors

- [x] T005: Implement `jira_project_role` resource (resource_project_role.go) — DONE, needs acc test
- [x] T006: Implement `jira_project_role_actor` resource (resource_project_role_actor.go) — DONE, needs acc test
- [x] T007: Implement `jira_project_roles` data source (data_source_project_roles.go) — DONE, needs acc test

## Phase 3: Permission Schemes

- [x] T008: Implement `jira_permission_scheme` resource (resource_permission_scheme.go) — DONE, needs acc test
- [ ] T009: Implement `jira_project_permission_scheme` resource (resource_project_permission_scheme.go + acc test)
- [x] T010: Implement `jira_permission_schemes` data source (data_source_permission_schemes.go) — DONE, needs acc test

## Phase 4: Notification Schemes

- [ ] T011: Implement `jira_notification_scheme` resource (resource_notification_scheme.go + resource_notification_scheme_acc_test.go)
- [ ] T012: Implement `jira_project_notification_scheme` resource (resource_project_notification_scheme.go + resource_project_notification_scheme_acc_test.go)
- [ ] T013: Implement `jira_notification_schemes` data source (data_source_notification_schemes.go + data_source_notification_schemes_acc_test.go)

## Phase 5: Security Schemes

- [ ] T014: Implement `jira_security_scheme` resource (resource_security_scheme.go + resource_security_scheme_acc_test.go)
- [ ] T015: Implement `jira_project_security_scheme` resource (resource_project_security_scheme.go + resource_project_security_scheme_acc_test.go)
- [ ] T016: Implement `jira_security_schemes` data source (data_source_security_schemes.go + data_source_security_schemes_acc_test.go)

## Phase 6: Integration and Polish

- [ ] T017: Register all new resources and data sources in provider.go
- [ ] T018: Write E2E acceptance test (e2e_governance_test.go)
- [ ] T019: Write Terraform Registry docs for all resources and data sources
- [ ] T020: Create examples/003-permissions-and-notifications/main.tf
