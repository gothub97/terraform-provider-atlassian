# Feature Specification: Issues and Workflows

**Feature Branch**: `002-issues-and-workflows`
**Created**: 2026-03-01
**Status**: Draft
**Input**: User description: "Second feature branch for the Atlassian Jira Cloud Terraform provider. Implements 9 resources (workflow, workflow scheme, issue, field, field configuration, field configuration scheme, screen, screen scheme, issue type screen scheme) and 3 data sources (workflows, fields, screens) building on the provider foundation from 001."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Manage Custom Fields (Priority: P1)

As a Jira administrator using Terraform, I want to define custom fields (text, number, select, date, etc.) so that I can version-control field definitions and consistently provision them across environments.

**Why this priority**: Custom fields are the foundational building block. Field configurations, screens, and issues all reference fields. Without fields, no downstream configuration can be tested.

**Independent Test**: Can be fully tested by creating a custom field, reading it back, updating its name/description, and deleting it. Delivers immediate value for field lifecycle management.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration declaring a custom text field with a name and description, **When** the user runs `terraform apply`, **Then** the field is created in Jira and its ID is stored in state.
2. **Given** an existing custom field managed by Terraform, **When** the user changes the field name or description, **Then** `terraform apply` updates the field in Jira.
3. **Given** a custom field that already exists in Jira, **When** the user imports it by field ID (e.g., `customfield_10001`), **Then** the field attributes are populated in Terraform state.
4. **Given** an existing custom field managed by Terraform, **When** the user removes it from configuration and runs `terraform apply`, **Then** the field is deleted (or trashed) in Jira.
5. **Given** no filters applied, **When** a user reads the `jira_fields` data source, **Then** all system and custom fields are returned with their types and metadata.
6. **Given** a type filter, **When** a user reads the `jira_fields` data source with `type = "custom"`, **Then** only custom fields are returned.

---

### User Story 2 - Configure Field Behavior (Priority: P2)

As a Jira administrator, I want to define field configurations and field configuration schemes in Terraform so that I can control which fields are required, hidden, or have description overrides per issue type.

**Why this priority**: Field configurations govern field behavior on screens and are a prerequisite for a fully configured project. They build directly on custom fields from P1.

**Independent Test**: Can be tested by creating a field configuration with field items, then creating a field configuration scheme that maps issue types to that configuration.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration declaring a field configuration with a name and field items, **When** the user runs `terraform apply`, **Then** the field configuration is created with the specified field behaviors (required, hidden, description).
2. **Given** an existing field configuration, **When** the user updates a field item to change its required/hidden status, **Then** `terraform apply` updates the configuration in Jira.
3. **Given** a Terraform configuration declaring a field configuration scheme with issue type mappings, **When** the user runs `terraform apply`, **Then** the scheme is created with the correct issue type to field configuration assignments.
4. **Given** an existing field configuration scheme, **When** the user adds or removes an issue type mapping, **Then** `terraform apply` updates the scheme accordingly.
5. **Given** existing field configurations and schemes in Jira, **When** the user imports them by ID, **Then** all attributes including mappings are populated in Terraform state.

---

### User Story 3 - Design Screens (Priority: P3)

As a Jira administrator, I want to define screens with tabs and ordered field lists, screen schemes that map screens to operations, and issue type screen schemes that map screen schemes to issue types — all in Terraform.

**Why this priority**: Screens control which fields users see during create, edit, and view operations. Screen schemes and issue type screen schemes complete the screen configuration chain needed for a fully provisioned project.

**Independent Test**: Can be tested by creating a screen with tabs and fields, creating a screen scheme mapping that screen to operations, and creating an issue type screen scheme.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration declaring a screen with tabs and ordered fields, **When** the user runs `terraform apply`, **Then** the screen is created with the correct tabs and field ordering.
2. **Given** an existing screen, **When** the user reorders fields within a tab or adds/removes fields, **Then** `terraform apply` updates the screen layout.
3. **Given** a screen scheme configuration mapping screens to operations (default, create, edit, view), **When** the user runs `terraform apply`, **Then** the screen scheme is created with correct operation mappings.
4. **Given** an issue type screen scheme configuration, **When** the user maps specific issue types to screen schemes, **Then** the mappings are created correctly.
5. **Given** no filters applied, **When** a user reads the `jira_screens` data source, **Then** all available screens are returned with their names and descriptions.
6. **Given** existing screens, screen schemes, or issue type screen schemes in Jira, **When** the user imports them by ID, **Then** all attributes are populated in state.

---

### User Story 4 - Define Workflows (Priority: P4)

As a Jira administrator, I want to create workflows with named statuses and transitions, and assign them to projects via workflow schemes — all managed through Terraform.

**Why this priority**: Workflows define the process that issues follow. Workflow schemes assign workflows to issue types within a project. Together they are essential for project configuration but depend on statuses (from 001) being available.

**Independent Test**: Can be tested by creating a workflow with statuses and transitions, then creating a workflow scheme that assigns it as the default workflow.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration declaring a workflow with statuses and transitions, **When** the user runs `terraform apply`, **Then** the workflow is created with the correct statuses and transition definitions.
2. **Given** an existing workflow, **When** the user adds or removes transitions, **Then** `terraform apply` updates the workflow.
3. **Given** a workflow scheme configuration with a default workflow and issue type mappings, **When** the user runs `terraform apply`, **Then** the scheme is created correctly.
4. **Given** an existing workflow scheme, **When** the user changes issue type to workflow mappings, **Then** `terraform apply` updates the scheme.
5. **Given** no filters applied, **When** a user reads the `jira_workflows` data source, **Then** all available workflows are returned.
6. **Given** a project filter, **When** a user reads the `jira_workflows` data source filtered by project, **Then** only workflows associated with that project are returned.
7. **Given** existing workflows or workflow schemes in Jira, **When** the user imports them by ID, **Then** all attributes are populated in state.

---

### User Story 5 - Manage Issues (Priority: P5)

As a Terraform user, I want to create and manage individual Jira issues with full field support (summary, description, priority, labels, components, custom fields) so that I can provision seed issues or track infrastructure work items as code.

**Why this priority**: Issues are the end-user facing artifact in Jira. While valuable, they depend on projects, issue types, and optionally workflows all being configured first. Managing issues via Terraform is a specialized use case (e.g., seed data, infrastructure tracking).

**Independent Test**: Can be tested by creating an issue with required fields (project, issue type, summary), reading it back, updating fields, and deleting it.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration declaring an issue with project key, issue type, and summary, **When** the user runs `terraform apply`, **Then** the issue is created in Jira and its key (e.g., PROJ-123) is stored in state.
2. **Given** an existing issue, **When** the user changes the summary, description, priority, labels, or custom field values, **Then** `terraform apply` updates the issue.
3. **Given** an existing issue key, **When** the user imports it (e.g., `terraform import jira_issue.example PROJ-123`), **Then** all supported fields are populated in state.
4. **Given** an issue managed by Terraform, **When** the user removes it from configuration and runs `terraform apply`, **Then** the issue is deleted from Jira.
5. **Given** an issue with a description, **When** the description is specified in Atlassian Document Format (ADF), **Then** the rich-text content is correctly stored and retrieved.

---

### User Story 6 - End-to-End Project Configuration (Priority: P6)

As a Jira administrator, I want to provision a complete project configuration chain in a single Terraform apply: project → custom fields → field configuration → screen → screen scheme → issue type screen scheme → workflow → workflow scheme → issue.

**Why this priority**: This is the integration story that validates all resources work together. It is the ultimate acceptance test but depends on all individual resources being functional first.

**Independent Test**: Can be tested by writing a single HCL configuration that declares all resources in the chain with proper cross-references and running `terraform apply` end-to-end.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration declaring the full chain (project, custom fields, field configuration, field configuration scheme, screen, screen scheme, issue type screen scheme, workflow, workflow scheme, and an issue), **When** the user runs `terraform apply`, **Then** all resources are created in the correct dependency order.
2. **Given** the full chain is provisioned, **When** the user runs `terraform plan`, **Then** no changes are detected (idempotent).
3. **Given** the full chain is provisioned, **When** the user runs `terraform destroy`, **Then** all resources are destroyed in reverse dependency order without errors.

---

### Edge Cases

- What happens when a user attempts to delete a workflow that is assigned to a workflow scheme? The system must return a clear error indicating the workflow is in use.
- What happens when a user attempts to delete a screen used by a screen scheme? The system must return a clear error.
- What happens when a custom field is deleted but still referenced in a field configuration? The system must handle this gracefully.
- What happens when an issue is updated with a status that has no valid transition path from the current status? The system must return a meaningful error listing the available transitions.
- What happens when a workflow scheme is updated while actively in use by projects? The system must handle draft-based updates for active schemes.
- What happens when a resource is deleted outside of Terraform? On the next `terraform plan`, the provider must detect the drift and propose recreation.
- What happens when two Terraform configurations reference the same field configuration or screen? The import mechanism must allow adopting existing resources.
- What happens when creating a workflow with duplicate status names or circular transitions? The system must surface the upstream validation error clearly.

## Requirements *(mandatory)*

### Functional Requirements

**Custom Fields (Resource: jira_field)**

- **FR-001**: System MUST allow users to create custom fields with a name, field type, and optional description and searcher key.
- **FR-002**: System MUST support all standard Jira custom field types: text (single-line and multi-line), number, select, multi-select, checkbox, radio buttons, date, datetime, URL, user picker, multi-user picker, group picker, multi-group picker, version, multi-version, project, labels, cascading select, and read-only.
- **FR-003**: System MUST allow updating a custom field's name, description, and searcher key after creation.
- **FR-004**: System MUST allow deleting (trashing) custom fields.
- **FR-005**: System MUST support importing existing custom fields by field ID.

**Field Configuration (Resource: jira_field_configuration)**

- **FR-006**: System MUST allow creating field configurations with a name, description, and list of field items.
- **FR-007**: Each field item MUST support specifying: field ID, whether the field is required, whether the field is hidden, and an optional description override.
- **FR-008**: System MUST allow updating field configuration metadata and field items.
- **FR-009**: System MUST allow deleting field configurations that are not in use.
- **FR-010**: System MUST support importing field configurations by ID.

**Field Configuration Scheme (Resource: jira_field_configuration_scheme)**

- **FR-011**: System MUST allow creating field configuration schemes with a name, description, and issue type to field configuration mappings.
- **FR-012**: System MUST allow specifying a default field configuration for unmapped issue types.
- **FR-013**: System MUST allow updating scheme metadata and mappings.
- **FR-014**: System MUST allow deleting schemes not assigned to projects.
- **FR-015**: System MUST support importing field configuration schemes by ID.

**Screen (Resource: jira_screen)**

- **FR-016**: System MUST allow creating screens with a name, description, and one or more tabs.
- **FR-017**: Each tab MUST support an ordered list of field references.
- **FR-018**: System MUST allow adding, removing, and reordering fields within tabs.
- **FR-019**: System MUST allow adding and removing tabs from screens.
- **FR-020**: System MUST allow deleting screens not in use by screen schemes.
- **FR-021**: System MUST support importing screens by ID.

**Screen Scheme (Resource: jira_screen_scheme)**

- **FR-022**: System MUST allow creating screen schemes that map screens to operations: default (required), create (optional), edit (optional), and view (optional).
- **FR-023**: System MUST allow updating operation-to-screen mappings.
- **FR-024**: System MUST allow deleting screen schemes not in use by issue type screen schemes.
- **FR-025**: System MUST support importing screen schemes by ID.

**Issue Type Screen Scheme (Resource: jira_issue_type_screen_scheme)**

- **FR-026**: System MUST allow creating issue type screen schemes with a name, description, and issue type to screen scheme mappings.
- **FR-027**: System MUST allow specifying a default screen scheme for unmapped issue types.
- **FR-028**: System MUST allow updating mappings.
- **FR-029**: System MUST allow deleting schemes not assigned to projects.
- **FR-030**: System MUST support importing by ID.

**Workflow (Resource: jira_workflow)**

- **FR-031**: System MUST allow creating workflows with a name, description, statuses, and transitions.
- **FR-032**: Each workflow status MUST reference an existing Jira status and include a name.
- **FR-033**: Each transition MUST specify a name, from-status, to-status, and optionally conditions, validators, and post-functions. All Jira rule types MUST be modeled as structured attributes with per-type configuration (not raw JSON pass-through).
- **FR-034**: System MUST allow updating workflow transitions and statuses.
- **FR-035**: System MUST allow deleting inactive workflows not assigned to workflow schemes.
- **FR-036**: System MUST support importing workflows by ID.

**Workflow Scheme (Resource: jira_workflow_scheme)**

- **FR-037**: System MUST allow creating workflow schemes with a name, description, default workflow, and issue type to workflow mappings.
- **FR-038**: System MUST allow updating scheme metadata and mappings, including handling draft-based updates for active schemes.
- **FR-039**: System MUST allow deleting workflow schemes not assigned to projects.
- **FR-040**: System MUST support importing workflow schemes by ID.

**Issue (Resource: jira_issue)**

- **FR-041**: System MUST allow creating issues with at minimum: project key, issue type, and summary.
- **FR-042**: System MUST support optional fields: description (ADF format), priority, status, assignee, reporter, labels, components, fix versions.
- **FR-043**: System MUST support setting custom field values on issues via a `custom_fields` map attribute where keys are field IDs (e.g., `customfield_10001`) and values are JSON-encoded strings matching the field's expected type.
- **FR-044**: System MUST allow updating all mutable issue fields.
- **FR-044a**: When the `status` attribute is changed, the system MUST find and execute valid workflow transitions to move the issue to the target status.
- **FR-045**: System MUST allow deleting issues.
- **FR-046**: System MUST support importing issues by issue key (e.g., PROJ-123).

**Data Sources**

- **FR-047**: System MUST provide a `jira_fields` data source that lists all system and custom fields, with optional filtering by type (system, custom).
- **FR-048**: System MUST provide a `jira_workflows` data source that lists workflows, with optional filtering by project.
- **FR-049**: System MUST provide a `jira_screens` data source that lists all available screens.

**Cross-Cutting**

- **FR-050**: All resources MUST detect out-of-band deletion and remove the resource from state on the next read.
- **FR-051**: All resources MUST support `terraform import` as specified per resource.
- **FR-052**: All resources and data sources MUST have acceptance tests that run against a real Jira Cloud instance.
- **FR-053**: Terraform Registry-compatible documentation MUST be provided for every resource and data source.
- **FR-054**: An end-to-end example MUST be provided demonstrating the full configuration chain.

### Key Entities

- **Custom Field**: A user-defined field with a name, type (text, number, select, etc.), optional description, and searcher key. Referenced by field configurations and screens.
- **Field Configuration**: Controls field behavior (required, hidden, description override) for a set of fields. Contains a list of field items.
- **Field Configuration Scheme**: Maps issue types to field configurations. Assigned to projects to control field behavior per issue type.
- **Screen**: A UI layout containing one or more tabs, each with an ordered list of fields. Controls which fields appear during create, edit, and view operations.
- **Screen Scheme**: Maps screens to operations (default, create, edit, view). Determines which screen is shown for each operation.
- **Issue Type Screen Scheme**: Maps issue types to screen schemes. Assigned to projects to control screen behavior per issue type.
- **Workflow**: A process definition with named statuses and transitions between them. Transitions may have conditions, validators, and post-functions.
- **Workflow Scheme**: Maps issue types to workflows. Assigned to projects to control which workflow governs each issue type. Active schemes require draft-based updates.
- **Issue**: An individual work item in a Jira project. Has required fields (project, issue type, summary) and optional fields (description, priority, assignee, labels, etc.). Identified by a key (e.g., PROJ-123).

## Clarifications

### Session 2026-03-01

- Q: Should the `status` attribute on `jira_issue` be writable (triggering workflow transitions), read-only computed, or set-on-create only? → A: Status is writable — changing it in HCL triggers the provider to find and execute valid workflow transitions to reach the target status.
- Q: How should custom field values on `jira_issue` be represented in the schema? → A: A single `custom_fields` map(string) attribute where keys are field IDs (e.g., `customfield_10001`) and values are JSON-encoded strings matching the field's expected type.
- Q: What scope of workflow transition rules (conditions, validators, post-functions) should be supported? → A: Full rule support — model all Jira condition, validator, and post-function types as structured attributes with per-type configuration.

## Assumptions

- The provider reuses the existing HTTP client from 001 with retry, pagination, and task-polling support.
- Custom field deletion uses the Jira "trash" mechanism (fields can be restored from trash via the Jira UI). The provider treats trashed fields as deleted.
- Workflow creation uses the bulk workflow creation endpoint which supports creating statuses and workflows atomically.
- The issue resource stores description in ADF (Atlassian Document Format) as a JSON string in Terraform state.
- Field configurations and field configuration schemes use the classic project endpoints, which are the standard for company-managed projects.
- Screen tab ordering follows the order declared in the Terraform configuration (first tab declared = first tab displayed).
- Field ordering within a screen tab follows the order declared in the Terraform configuration.
- The workflow scheme resource handles draft-based updates transparently — if a scheme is active, the provider creates/updates a draft and publishes it.
- Issue status changes via the issue resource use the transitions endpoint to move issues through valid workflow transitions rather than directly setting status.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All 9 resources (field, field configuration, field configuration scheme, screen, screen scheme, issue type screen scheme, workflow, workflow scheme, issue) support create, read, update, delete, and import operations verified by passing acceptance tests.
- **SC-002**: All 3 data sources (fields, workflows, screens) return accurate, typed data verified by passing acceptance tests.
- **SC-003**: A single `terraform apply` of the end-to-end example provisions the complete configuration chain (project → fields → field config → screen → screen scheme → workflow → workflow scheme → issue) without errors.
- **SC-004**: A subsequent `terraform plan` on the provisioned end-to-end example detects zero changes (idempotent).
- **SC-005**: `terraform destroy` on the end-to-end example removes all resources without errors.
- **SC-006**: Every resource can be imported and a subsequent `terraform plan` shows no changes.
- **SC-007**: All acceptance tests pass when run against a real Jira Cloud instance.
- **SC-008**: Terraform Registry-compatible documentation exists for all 9 resources and 3 data sources.
