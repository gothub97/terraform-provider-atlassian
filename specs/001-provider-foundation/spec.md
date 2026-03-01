# Feature Specification: Provider Foundation

**Feature Branch**: `001-provider-foundation`
**Created**: 2026-03-01
**Status**: Draft
**Input**: User description: "Build the foundation of a Terraform provider for Atlassian Jira Cloud"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure and Authenticate the Provider (Priority: P1)

An infrastructure engineer installs the Atlassian Terraform provider and
configures it with their Jira Cloud instance URL, email address, and API
token. The provider authenticates against the Jira Cloud REST API v3 and
confirms the connection is valid before managing any resources. The
engineer can supply credentials either directly in HCL or through
environment variables for CI/CD pipelines.

**Why this priority**: Without a working provider that authenticates
successfully, no other resources or data sources can function. This is
the absolute prerequisite for everything else.

**Independent Test**: Can be fully tested by writing a minimal HCL
configuration with only the provider block and running `terraform plan`.
The provider connects to the Jira instance and the `jira_myself` data
source returns the authenticated user's information, confirming the
connection is healthy.

**Acceptance Scenarios**:

1. **Given** valid HCL credentials (URL, email, API token), **When** the
   user runs `terraform plan`, **Then** the provider authenticates
   successfully and reports no errors.
2. **Given** credentials set via environment variables (ATLASSIAN_URL,
   ATLASSIAN_EMAIL, ATLASSIAN_API_TOKEN) and no HCL credentials,
   **When** the user runs `terraform plan`, **Then** the provider picks
   up the environment variables and authenticates successfully.
3. **Given** HCL credentials that override environment variables,
   **When** both are present, **Then** the HCL values take precedence.
4. **Given** an invalid API token, **When** the user runs `terraform
   plan`, **Then** the provider returns a clear authentication error
   message identifying the credential that failed.
5. **Given** a valid configuration with the `jira_myself` data source,
   **When** the user runs `terraform apply`, **Then** the data source
   returns the authenticated user's account ID, display name, email,
   and active status.

---

### User Story 2 - Manage Jira Projects (Priority: P2)

An infrastructure engineer manages Jira projects as code. They can
create new projects by specifying a key, name, project type, lead, and
description. They can update project attributes (name, description,
lead) and delete projects they no longer need. They can also import
existing projects into Terraform state by project key or ID. A read-only
data source allows looking up any existing project by its key.

**Why this priority**: Projects are the top-level organizational
container in Jira. Issue types, priorities, and statuses all exist in
the context of projects, making this the first resource other entities
depend on.

**Independent Test**: Can be fully tested by creating a project with
`terraform apply`, verifying it exists in Jira, updating its name,
verifying the update, importing an existing project, and finally
destroying it.

**Acceptance Scenarios**:

1. **Given** a valid project configuration with key, name, and project
   type, **When** the user runs `terraform apply`, **Then** the project
   is created in Jira with the specified attributes.
2. **Given** an existing managed project, **When** the user changes the
   project name or description in HCL and runs `terraform apply`,
   **Then** the project is updated in Jira to match.
3. **Given** an existing managed project, **When** the user removes it
   from the configuration and runs `terraform apply`, **Then** the
   project is deleted from Jira.
4. **Given** a project that already exists in Jira, **When** the user
   runs `terraform import` with the project key or ID, **Then** the
   project's full state is imported into Terraform.
5. **Given** a project that was modified directly in Jira (outside
   Terraform), **When** the user runs `terraform plan`, **Then**
   Terraform detects the drift and shows the differences.
6. **Given** the `jira_project` data source with a project key,
   **When** the user runs `terraform plan`, **Then** the data source
   returns the project's ID, name, type, lead, and description.

---

### User Story 3 - Manage Issue Types (Priority: P3)

An infrastructure engineer defines custom issue types as code. They can
create new issue types with a name, description, and hierarchy level
(standard or subtask). They can update issue type attributes and delete
issue types that are no longer needed. A data source lists all issue
types, with optional filtering by project.

**Why this priority**: Issue types are fundamental to Jira workflows.
They define what kinds of work items a project supports and are required
before issues can be created.

**Independent Test**: Can be fully tested by creating an issue type with
`terraform apply`, verifying it appears in Jira, updating its
description, importing an existing issue type by ID, and deleting it.

**Acceptance Scenarios**:

1. **Given** a valid issue type configuration with name and hierarchy
   level, **When** the user runs `terraform apply`, **Then** the issue
   type is created in Jira.
2. **Given** an existing managed issue type, **When** the user changes
   its name or description in HCL, **Then** the issue type is updated
   in Jira.
3. **Given** an existing managed issue type, **When** the user removes
   it from configuration, **Then** the issue type is deleted from Jira.
4. **Given** an issue type that exists in Jira, **When** the user runs
   `terraform import` with the issue type ID, **Then** the state is
   fully imported.
5. **Given** the `jira_issue_types` data source, **When** the user runs
   `terraform plan`, **Then** all issue types are returned with their
   IDs, names, descriptions, and hierarchy levels.
6. **Given** the `jira_issue_types` data source with a project filter,
   **When** the user runs `terraform plan`, **Then** only issue types
   associated with that project are returned.

---

### User Story 4 - Manage Priorities (Priority: P4)

An infrastructure engineer defines issue priorities as code. They can
create priorities with a name, description, status color, and icon.
They can update and delete priorities. A data source lists all
available priorities.

**Why this priority**: Priorities customize how teams triage work. They
are independent of projects and issue types but round out the core
configuration vocabulary.

**Independent Test**: Can be fully tested by creating a priority,
verifying it in Jira, updating its description and color, importing an
existing priority, and deleting it.

**Acceptance Scenarios**:

1. **Given** a valid priority configuration with name and status color,
   **When** the user runs `terraform apply`, **Then** the priority is
   created in Jira.
2. **Given** an existing managed priority, **When** the user changes its
   name, description, or color, **Then** the priority is updated in
   Jira.
3. **Given** an existing managed priority, **When** the user removes it
   from configuration, **Then** the priority is deleted from Jira.
4. **Given** a priority that exists in Jira, **When** the user runs
   `terraform import` with the priority ID, **Then** the state is
   fully imported.
5. **Given** the `jira_priorities` data source, **When** the user runs
   `terraform plan`, **Then** all priorities are returned with IDs,
   names, descriptions, colors, and icon URLs.

---

### User Story 5 - Manage Statuses (Priority: P5)

An infrastructure engineer defines workflow statuses as code. They can
create statuses within a status category (to-do, in-progress, done),
with a name and description. They can update and delete statuses. A data
source lists all statuses, with optional project filtering.

**Why this priority**: Statuses define the workflow states that issues
move through. They complete the core set of Jira configuration
primitives needed for a functional provider.

**Independent Test**: Can be fully tested by creating a status,
verifying it in Jira, updating its name, importing an existing status,
and deleting it.

**Acceptance Scenarios**:

1. **Given** a valid status configuration with name, status category,
   and scope, **When** the user runs `terraform apply`, **Then** the
   status is created in Jira.
2. **Given** an existing managed status, **When** the user changes its
   name or description, **Then** the status is updated in Jira.
3. **Given** an existing managed status, **When** the user removes it
   from configuration, **Then** the status is deleted from Jira.
4. **Given** a status that exists in Jira, **When** the user runs
   `terraform import` with the status ID, **Then** the state is
   fully imported.
5. **Given** the `jira_statuses` data source, **When** the user runs
   `terraform plan`, **Then** all statuses are returned with IDs,
   names, descriptions, and categories.
6. **Given** the `jira_statuses` data source with a project filter,
   **When** the user runs `terraform plan`, **Then** only statuses
   associated with that project are returned.

---

### User Story 6 - End-to-End Example Configuration (Priority: P6)

An infrastructure engineer uses a sample HCL configuration from the
examples directory to provision a complete Jira project setup: a project
with custom issue types, priorities, and statuses. The example
demonstrates real-world provider usage and serves as a quick-start
guide.

**Why this priority**: The example configuration validates that all
resources work together and gives users a concrete starting point. It
is a deliverable rather than core functionality.

**Independent Test**: Can be tested by running `terraform apply` on the
example configuration against a live Jira instance and verifying all
resources are created.

**Acceptance Scenarios**:

1. **Given** the sample configuration in the examples directory,
   **When** a user runs `terraform init && terraform apply`, **Then**
   a Jira project with custom issue types, priorities, and statuses is
   provisioned end to end.
2. **Given** the provisioned example resources, **When** the user runs
   `terraform destroy`, **Then** all resources are cleaned up in the
   correct dependency order.

---

### Out of Scope (This Branch)

- Project scheme fields (notification, permission, issue security, issue
  type, workflow, field configuration, issue type screen schemes).
- Project category and avatar management.
- Project archiving and restoring.
- Jira Server or Data Center — this provider targets Jira Cloud only.
- Workflow and screen management resources.
- Issue creation or manipulation resources.

### Edge Cases

- What happens when a user attempts to create a project with a key that
  already exists? The provider MUST surface the Jira API conflict error
  clearly.
- What happens when a resource is deleted outside Terraform (drift)?
  The provider MUST detect the missing resource during `Read` and
  remove it from state gracefully.
- What happens when the API token is revoked mid-session? The provider
  MUST return a clear authentication error, not a generic 500.
- What happens when the Jira instance rate-limits the provider? The
  HTTP client MUST retry with exponential backoff transparently.
- What happens when deleting a priority that is in use by issues? The
  API may return a conflict; the provider MUST surface this clearly.
- What happens when deleting an issue type that is assigned to issues?
  The API requires an alternative issue type ID; the provider MUST
  document this requirement and surface the error.
- What happens when the status API's bulk operations partially fail?
  The provider MUST report per-status errors through diagnostics.

## Clarifications

### Session 2026-03-01

- Q: When a user changes a project key in HCL, should Terraform destroy and recreate the project (ForceNew) or attempt an in-place rename? → A: ForceNew — changing the project key forces replacement.
- Q: The Jira API requires a projectTemplateKey for project creation. Should users specify it, or should the provider derive it from projectTypeKey? → A: Auto-derive a default template from the project type; expose an optional override attribute for advanced use cases.
- Q: Should the jira_project resource expose advanced API fields (schemes, category, avatar, URL) in this foundation branch? → A: No — core fields only (key, name, type, template, lead, description, assignee type). Schemes and advanced fields are explicitly deferred to a follow-up branch.

## Requirements *(mandatory)*

### Functional Requirements

**Provider Configuration:**

- **FR-001**: The provider MUST accept an Atlassian Cloud instance URL,
  an email address, and an API token as configuration attributes.
- **FR-002**: The provider MUST accept configuration via environment
  variables (ATLASSIAN_URL, ATLASSIAN_EMAIL, ATLASSIAN_API_TOKEN) as
  fallback when HCL attributes are not set.
- **FR-003**: HCL-specified configuration MUST take precedence over
  environment variables when both are present.
- **FR-004**: The API token MUST be marked as sensitive and MUST NOT
  appear in logs or plan output.
- **FR-005**: The provider MUST authenticate using HTTP Basic Auth
  (email + API token) against the Jira Cloud REST API v3.
- **FR-006**: The provider MUST validate connectivity during
  configuration by making a lightweight API call and return a clear
  error if authentication fails.

**HTTP Client:**

- **FR-007**: The provider MUST use a single centralized HTTP client
  for all API interactions.
- **FR-008**: The HTTP client MUST implement retry with exponential
  backoff and jitter for HTTP 429 (rate limit) and 5xx responses.
- **FR-009**: The HTTP client MUST transparently handle paginated
  endpoints, returning complete result sets to callers.

**Resource — jira_project:**

- **FR-010**: Users MUST be able to create a Jira project by specifying
  key, name, project type, lead account ID, and description. The
  project key MUST be immutable after creation — changing it forces
  replacement (destroy + recreate). The provider MUST auto-derive a
  default project template from the project type. An optional project
  template key attribute MUST be available for advanced overrides.
- **FR-011**: Users MUST be able to update a project's name,
  description, lead, and assignee type. The project key is NOT
  updatable in-place.
- **FR-012**: Users MUST be able to delete a project.
- **FR-013**: Users MUST be able to import an existing project by
  project key or ID.
- **FR-014**: The Read operation MUST detect drift between remote state
  and Terraform state.

**Resource — jira_issue_type:**

- **FR-015**: Users MUST be able to create an issue type with a name,
  description, and hierarchy level (standard or subtask).
- **FR-016**: Users MUST be able to update an issue type's name,
  description, and avatar.
- **FR-017**: Users MUST be able to delete an issue type.
- **FR-018**: Users MUST be able to import an existing issue type by ID.
- **FR-019**: The Read operation MUST detect drift.

**Resource — jira_priority:**

- **FR-020**: Users MUST be able to create a priority with a name,
  description, status color, and optionally an icon URL or avatar ID.
- **FR-021**: Users MUST be able to update a priority's name,
  description, color, and icon.
- **FR-022**: Users MUST be able to delete a priority (the provider
  MUST handle the asynchronous delete by polling the task endpoint).
- **FR-023**: Users MUST be able to import an existing priority by ID.
- **FR-024**: The Read operation MUST detect drift.

**Resource — jira_status:**

- **FR-025**: Users MUST be able to create a status with a name,
  description, status category (TODO, IN_PROGRESS, DONE), and scope
  (GLOBAL or PROJECT with project ID).
- **FR-026**: Users MUST be able to update a status's name, description,
  and status category.
- **FR-027**: Users MUST be able to delete a status.
- **FR-028**: Users MUST be able to import an existing status by ID.
- **FR-029**: The Read operation MUST detect drift.

**Data Sources:**

- **FR-030**: The `jira_project` data source MUST look up a project by
  key and return its full attributes.
- **FR-031**: The `jira_issue_types` data source MUST list all issue
  types, with optional filtering by project ID.
- **FR-032**: The `jira_priorities` data source MUST list all
  priorities.
- **FR-033**: The `jira_statuses` data source MUST list all statuses,
  with optional filtering by project ID.
- **FR-034**: The `jira_myself` data source MUST return the
  authenticated user's account ID, display name, email, and active
  status.

**Documentation:**

- **FR-035**: Every resource and data source MUST have Terraform
  Registry-compatible documentation with at least one HCL example.
- **FR-036**: The provider-level documentation MUST explain
  authentication setup and environment variable configuration.

**Testing:**

- **FR-037**: Every resource MUST have acceptance tests covering create,
  read, update, delete, and import — run against a live Jira instance.
- **FR-038**: Every data source MUST have acceptance tests verifying
  correct data retrieval against a live Jira instance.

**Examples:**

- **FR-039**: A sample HCL configuration in examples/ MUST demonstrate
  provisioning a project with custom issue types, priorities, and
  statuses end to end.

### Key Entities

- **Provider Configuration**: The authentication and connection settings
  for the Atlassian Cloud instance. Attributes: instance URL, email,
  API token.
- **Project**: The top-level organizational container in Jira.
  Attributes: ID, key (immutable), name, project type, project
  template key (optional, auto-derived from type), lead account ID,
  description, assignee type.
- **Issue Type**: Defines the kinds of work items (e.g., Bug, Story,
  Task). Attributes: ID, name, description, hierarchy level, avatar
  ID, icon URL, subtask flag.
- **Priority**: Defines the urgency levels for issues (e.g., Highest,
  High, Medium). Attributes: ID, name, description, status color, icon
  URL, avatar ID, default flag.
- **Status**: Defines the workflow states issues move through (e.g.,
  To Do, In Progress, Done). Attributes: ID, name, description, status
  category, scope (global or project-scoped).
- **Myself**: The currently authenticated user. Attributes: account ID,
  display name, email address, active flag, account type, time zone.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The provider initializes and authenticates against a live
  Jira Cloud instance on first `terraform plan` without errors.
- **SC-002**: Each of the four managed resources (project, issue type,
  priority, status) completes a full create-read-update-delete cycle
  through Terraform without manual intervention.
- **SC-003**: Each of the four managed resources can be imported from an
  existing Jira instance into Terraform state without data loss.
- **SC-004**: Each of the five data sources returns accurate, complete
  data matching what the Jira UI shows.
- **SC-005**: Drift detection identifies and reports changes made
  outside Terraform for every managed resource.
- **SC-006**: All acceptance tests pass against a live Jira Cloud
  instance in a single test run.
- **SC-007**: The sample end-to-end HCL configuration provisions a
  project with custom issue types, priorities, and statuses
  successfully on a clean Jira instance.
- **SC-008**: Terraform Registry-compatible documentation is generated
  for the provider, all four resources, and all five data sources.
