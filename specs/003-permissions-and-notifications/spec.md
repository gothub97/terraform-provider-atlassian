# Feature Specification: Permissions and Notifications

**Feature Branch**: `003-permissions-and-notifications`
**Created**: 2026-03-01
**Status**: Draft
**Input**: Third feature branch for the Atlassian Jira Cloud Terraform provider. Implements 10 resources and 6 data sources for access control and notification configuration.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Manage Groups and Memberships (Priority: P1)

As a Jira administrator, I want to create groups and manage their memberships in Terraform so that I can version-control team composition for use in permission and notification schemes.

**Why this priority**: Groups are the foundational identity construct. Permission schemes, notification schemes, security schemes, and project role actors all reference groups. Without groups, no downstream access configuration can be tested.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration declaring a group with a name, **When** `terraform apply` runs, **Then** the group is created in Jira and its group ID is stored in state.
2. **Given** an existing group, **When** the user removes it from configuration, **Then** `terraform apply` deletes the group.
3. **Given** a group membership resource linking a user account ID to a group, **When** `terraform apply` runs, **Then** the user is added to the group.
4. **Given** an existing group membership, **When** it is removed from configuration, **Then** `terraform apply` removes the user from the group.
5. **Given** existing groups in Jira, **When** the `jira_groups` data source is read, **Then** all groups are returned, optionally filtered by name prefix.
6. **Given** a search query, **When** the `jira_users` data source is read, **Then** matching users are returned with their account IDs and display names.

---

### User Story 2 - Manage Project Roles and Actors (Priority: P2)

As a Jira administrator, I want to define project roles and assign users/groups as actors within specific projects so that I can control access at the role level.

**Why this priority**: Project roles are referenced by permission schemes, notification schemes, and security schemes. They must exist before those schemes can grant permissions to roles.

**Acceptance Scenarios**:

1. **Given** a project role configuration with name and description, **When** `terraform apply` runs, **Then** the role is created globally.
2. **Given** an existing project role, **When** the name or description is changed, **Then** `terraform apply` updates it.
3. **Given** a project role actor resource linking a user or group to a role within a project, **When** `terraform apply` runs, **Then** the actor is added to the role for that project.
4. **Given** an existing role actor, **When** it is removed from configuration, **Then** `terraform apply` removes the actor.
5. **Given** existing project roles, **When** the `jira_project_roles` data source is read, **Then** all roles are returned.

---

### User Story 3 - Manage Permission Schemes (Priority: P3)

As a Jira administrator, I want to create permission schemes with grants (permission key + holder) and assign them to projects so that I can control who can do what in each project.

**Why this priority**: Permission schemes are the core access control mechanism. They depend on groups and roles being available.

**Acceptance Scenarios**:

1. **Given** a permission scheme with name, description, and a list of permission grants, **When** `terraform apply` runs, **Then** the scheme is created with all grants.
2. **Given** an existing permission scheme, **When** grants are added or removed, **Then** `terraform apply` updates the scheme.
3. **Given** a project permission scheme resource linking a project to a permission scheme, **When** `terraform apply` runs, **Then** the scheme is assigned to the project.
4. **Given** existing permission schemes, **When** the `jira_permission_schemes` data source is read, **Then** all schemes are returned.
5. **Given** an existing permission scheme, **When** imported by ID, **Then** all attributes including grants are populated in state.

---

### User Story 4 - Manage Notification Schemes (Priority: P4)

As a Jira administrator, I want to create notification schemes that define who gets notified on issue events, and assign them to projects.

**Why this priority**: Notification schemes depend on groups, roles, and users being available as recipient types.

**Acceptance Scenarios**:

1. **Given** a notification scheme with name, description, and event notifications, **When** `terraform apply` runs, **Then** the scheme is created with all notifications.
2. **Given** an existing notification scheme, **When** name/description is changed or notifications are added/removed, **Then** `terraform apply` updates it.
3. **Given** a project notification scheme resource, **When** `terraform apply` runs, **Then** the scheme is assigned to the project.
4. **Given** existing notification schemes, **When** the `jira_notification_schemes` data source is read, **Then** all schemes are returned.

---

### User Story 5 - Manage Issue Security Schemes (Priority: P5)

As a Jira administrator, I want to create issue security schemes with security levels and members, and assign them to projects so that I can restrict who can view specific issues.

**Why this priority**: Security schemes are the most complex access control resource. They depend on groups, roles, and users.

**Acceptance Scenarios**:

1. **Given** a security scheme with name, description, and security levels with members, **When** `terraform apply` runs, **Then** the scheme is created with all levels and members.
2. **Given** an existing security scheme, **When** name/description is changed, **Then** `terraform apply` updates it.
3. **Given** a project security scheme resource, **When** `terraform apply` runs, **Then** the scheme is assigned to the project.
4. **Given** existing security schemes, **When** the `jira_security_schemes` data source is read, **Then** all schemes are returned.

---

### User Story 6 - End-to-End Project Governance (Priority: P6)

As a Jira administrator, I want to provision a complete project governance setup in a single `terraform apply`: groups → memberships → roles → role actors → permission scheme → notification scheme → security scheme → associate all three to a project.

**Acceptance Scenarios**:

1. **Given** a full governance chain configuration, **When** `terraform apply` runs, **Then** all resources are created in correct dependency order.
2. **Given** the full chain is provisioned, **When** `terraform plan` runs, **Then** zero changes are detected (idempotent).
3. **Given** the full chain is provisioned, **When** `terraform destroy` runs, **Then** all resources are destroyed without errors.

---

### Edge Cases

- Deleting a group that is referenced by a permission scheme grant → Jira allows this (grant becomes orphaned). Provider should handle gracefully.
- Deleting a project role that is in use by schemes → API requires a `swap` role ID. Provider should surface clear error.
- Assigning a security scheme to a project when issues have existing security levels → API requires level mappings. Provider should support this.
- Notification scheme update only supports name/description via PUT; notifications must be managed via separate add/remove endpoints.
- Permission scheme PUT with `permissions` array overwrites ALL grants — provider must send full state on update.

## Requirements *(mandatory)*

### Functional Requirements

**Group (Resource: jira_group)**

- **FR-001**: System MUST create groups with a name. Group ID (UUID) returned by API stored in state.
- **FR-002**: System MUST delete groups. No update supported (groups only have a name which cannot be changed).
- **FR-003**: System MUST support importing groups by group ID.

**Group Membership (Resource: jira_group_membership)**

- **FR-004**: System MUST add a user (by account ID) to a group (by group ID).
- **FR-005**: System MUST remove users from groups on delete.
- **FR-006**: System MUST support importing by composite key `groupId/accountId`.

**Project Role (Resource: jira_project_role)**

- **FR-007**: System MUST create project roles with name and optional description.
- **FR-008**: System MUST update project role name and description (via PUT with both fields).
- **FR-009**: System MUST delete project roles.
- **FR-010**: System MUST support importing by role ID.

**Project Role Actor (Resource: jira_project_role_actor)**

- **FR-011**: System MUST add actors (user by account ID, or group by group ID) to a project role for a specific project.
- **FR-012**: System MUST remove actors on delete.
- **FR-013**: System MUST support importing by composite key `projectKey/roleId/actorType/actorValue`.

**Permission Scheme (Resource: jira_permission_scheme)**

- **FR-014**: System MUST create permission schemes with name, description, and permission grants.
- **FR-015**: Each grant specifies a permission key and a holder (type + value).
- **FR-016**: Supported holder types: `anyone`, `applicationRole`, `assignee`, `group`, `projectLead`, `projectRole`, `reporter`, `user`, `userCustomField`, `groupCustomField`.
- **FR-017**: System MUST update schemes including full grant replacement via PUT.
- **FR-018**: System MUST delete permission schemes.
- **FR-019**: System MUST support importing by scheme ID.

**Project Permission Scheme (Resource: jira_project_permission_scheme)**

- **FR-020**: System MUST assign a permission scheme to a project via PUT.
- **FR-021**: System MUST read the currently assigned scheme.
- **FR-022**: On update, system MUST reassign to the new scheme.
- **FR-023**: On delete, system should reassign the default permission scheme (ID 0) or leave as-is with warning.
- **FR-024**: System MUST support importing by project key.

**Notification Scheme (Resource: jira_notification_scheme)**

- **FR-025**: System MUST create notification schemes with name, description, and event notifications.
- **FR-026**: Each event notification specifies an event ID and a list of recipients (type + parameter).
- **FR-027**: Supported notification types: `CurrentAssignee`, `Reporter`, `CurrentUser`, `ProjectLead`, `ComponentLead`, `User`, `Group`, `ProjectRole`, `EmailAddress`, `AllWatchers`, `UserCustomField`, `GroupCustomField`.
- **FR-028**: System MUST update scheme name/description via PUT and manage notifications via add/remove endpoints.
- **FR-029**: System MUST delete notification schemes.
- **FR-030**: System MUST support importing by scheme ID.

**Project Notification Scheme (Resource: jira_project_notification_scheme)**

- **FR-031**: System MUST assign a notification scheme to a project (via project update endpoint, setting `notificationScheme` field).
- **FR-032**: System MUST read the currently assigned scheme.
- **FR-033**: System MUST support importing by project key.

**Issue Security Scheme (Resource: jira_security_scheme)**

- **FR-034**: System MUST create security schemes with name, description, and security levels with members.
- **FR-035**: Each security level has a name, description, isDefault flag, and a list of members (type + parameter).
- **FR-036**: Supported member types: `reporter`, `group`, `user`, `projectrole`, `applicationRole`.
- **FR-037**: System MUST update scheme name/description via PUT. Levels managed via separate endpoints.
- **FR-038**: System MUST delete security schemes.
- **FR-039**: System MUST support importing by scheme ID.

**Project Security Scheme (Resource: jira_project_security_scheme)**

- **FR-040**: System MUST assign an issue security scheme to a project via PUT (async operation).
- **FR-041**: System MUST read the currently assigned scheme.
- **FR-042**: System MUST support importing by project key.

**Data Sources**

- **FR-043**: `jira_permission_schemes` — list all permission schemes.
- **FR-044**: `jira_project_roles` — list all project roles.
- **FR-045**: `jira_notification_schemes` — list all notification schemes (paginated).
- **FR-046**: `jira_security_schemes` — list all issue security schemes.
- **FR-047**: `jira_groups` — list groups, optionally filtered by name prefix.
- **FR-048**: `jira_users` — search users by query string or account ID.

**Cross-Cutting**

- **FR-049**: All resources MUST detect out-of-band deletion and remove from state on next read.
- **FR-050**: All resources MUST support `terraform import`.
- **FR-051**: All resources and data sources MUST have acceptance tests against real Jira Cloud.
- **FR-052**: Terraform Registry-compatible documentation for every resource and data source.
- **FR-053**: End-to-end example in `examples/003-permissions-and-notifications/`.

### Key Entities

- **Group**: A named collection of users. Identified by group ID (UUID). Referenced by permission/notification/security schemes.
- **Group Membership**: Associates a user (account ID) with a group. Composite identity.
- **Project Role**: A global role definition (e.g., Administrators, Developers). Has actors per project.
- **Project Role Actor**: Links a user or group to a role within a specific project.
- **Permission Scheme**: Defines who can perform which actions. Contains grants (permission key + holder).
- **Permission Grant Holder**: Types include group, user, projectRole, assignee, reporter, projectLead, anyone, applicationRole, userCustomField, groupCustomField.
- **Notification Scheme**: Defines who receives notifications for issue events. Contains event-to-recipient mappings.
- **Issue Security Scheme**: Restricts issue visibility. Contains security levels with members.
- **Security Level**: A named visibility tier within a security scheme. Has members (groups, users, roles, reporter).

## API Endpoint Reference

### Permission Schemes
- `GET /rest/api/3/permissionscheme` — List all (with `expand=permissions`)
- `POST /rest/api/3/permissionscheme` — Create with grants
- `GET /rest/api/3/permissionscheme/{id}` — Read by ID
- `PUT /rest/api/3/permissionscheme/{id}` — Update (grants array overwrites ALL existing)
- `DELETE /rest/api/3/permissionscheme/{id}` — Delete
- `PUT /rest/api/3/project/{key}/permissionscheme` — Assign to project (body: `{"id": N}`)
- `GET /rest/api/3/project/{key}/permissionscheme` — Get project's scheme

### Project Roles
- `GET /rest/api/3/role` — List all
- `POST /rest/api/3/role` — Create
- `GET /rest/api/3/role/{id}` — Read by ID
- `PUT /rest/api/3/role/{id}` — Full update (name + description required)
- `DELETE /rest/api/3/role/{id}` — Delete (optional `swap` query param)
- `POST /rest/api/3/project/{key}/role/{id}` — Add actors
- `DELETE /rest/api/3/project/{key}/role/{id}` — Remove actor (query params: `user` or `groupId`)
- `GET /rest/api/3/project/{key}/role/{id}` — Get role actors for project

### Notification Schemes
- `GET /rest/api/3/notificationscheme` — List all (paginated)
- `POST /rest/api/3/notificationscheme` — Create with events
- `GET /rest/api/3/notificationscheme/{id}` — Read by ID (expand=notificationSchemeEvents)
- `PUT /rest/api/3/notificationscheme/{id}` — Update name/description only
- `DELETE /rest/api/3/notificationscheme/{id}` — Delete
- `PUT /rest/api/3/notificationscheme/{id}/notification` — Add notifications
- `DELETE /rest/api/3/notificationscheme/{id}/notification/{notifId}` — Remove notification
- `GET /rest/api/3/project/{key}/notificationscheme` — Get project's scheme
- Project assignment via `PUT /rest/api/3/project/{key}` with `notificationScheme` field

### Issue Security Schemes
- `GET /rest/api/3/issuesecurityschemes` — List all
- `POST /rest/api/3/issuesecurityschemes` — Create with levels and members
- `GET /rest/api/3/issuesecurityschemes/{id}` — Read by ID
- `PUT /rest/api/3/issuesecurityschemes/{id}` — Update name/description only
- `DELETE /rest/api/3/issuesecurityschemes/{id}` — Delete
- `PUT /rest/api/3/issuesecurityschemes/{schemeId}/level` — Add levels
- `PUT /rest/api/3/issuesecurityschemes/{schemeId}/level/{levelId}` — Update level
- `DELETE /rest/api/3/issuesecurityschemes/{schemeId}/level/{levelId}` — Delete level (async)
- `PUT /rest/api/3/issuesecurityschemes/{schemeId}/level/{levelId}/member` — Add members
- `DELETE /rest/api/3/issuesecurityschemes/{schemeId}/level/{levelId}/member/{memberId}` — Remove member
- `PUT /rest/api/3/issuesecurityschemes/project` — Assign to project (async)
- `GET /rest/api/3/project/{key}/issuesecuritylevelscheme` — Get project's scheme

### Groups
- `POST /rest/api/3/group` — Create (body: `{"name": "..."}`)
- `DELETE /rest/api/3/group?groupId=...` — Delete
- `GET /rest/api/3/group/bulk` — List all (paginated)
- `GET /rest/api/3/groups/picker?query=...` — Search by prefix
- `POST /rest/api/3/group/user?groupId=...` — Add user (body: `{"accountId": "..."}`)
- `DELETE /rest/api/3/group/user?groupId=...&accountId=...` — Remove user
- `GET /rest/api/3/group/member?groupId=...` — List members (paginated)

### Users
- `GET /rest/api/3/user/search?query=...` — Search users
- `GET /rest/api/3/user?accountId=...` — Get user by ID
- `GET /rest/api/3/user/bulk?accountId=...` — Bulk get users

## Assumptions

- Provider reuses existing HTTP client from 001/002 with retry, pagination, and task-polling.
- Group names are immutable — Jira does not support renaming groups. Delete + recreate is required.
- Permission scheme PUT with `permissions` array replaces ALL grants. Provider always sends full grant state.
- Notification scheme notifications are managed incrementally (add/remove), not as a bulk replace.
- Security scheme level deletion is async — provider must poll task endpoint.
- Security scheme project assignment is async — provider must poll task endpoint.
- Project notification scheme assignment uses the project update endpoint (no dedicated assign endpoint).
- All resources use classic (company-managed) project APIs.

## Success Criteria *(mandatory)*

- **SC-001**: All 10 resources support CRUD + import verified by passing acceptance tests.
- **SC-002**: All 6 data sources return accurate typed data verified by passing acceptance tests.
- **SC-003**: E2E example provisions complete governance chain without errors.
- **SC-004**: Subsequent `terraform plan` on E2E shows zero changes (idempotent).
- **SC-005**: `terraform destroy` removes all resources without errors.
- **SC-006**: All acceptance tests pass against real Jira Cloud.
- **SC-007**: Terraform Registry-compatible docs exist for all resources and data sources.
