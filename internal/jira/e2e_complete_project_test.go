package jira_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccE2ECompleteProjectConfig returns a Terraform config that provisions all 22
// resource types in a single dependency graph — mirroring examples/complete-project.
func testAccE2ECompleteProjectConfig(suffix string) string {
	rKey := fmt.Sprintf("TFC%s", suffix[:4])
	return testAccProviderConfig + fmt.Sprintf(`
data "atlassian_jira_myself" "cp" {}

# --- Foundation: Issue Type, Priority, Fields, Group, Project Role ---

resource "atlassian_jira_issue_type" "cp" {
  name            = "tf-e2e-cp-it-%[1]s"
  description     = "E2E complete project issue type"
  hierarchy_level = 0
}

resource "atlassian_jira_priority" "cp" {
  name         = "tf-e2e-cp-pri-%[1]s"
  description  = "E2E complete project priority"
  status_color = "#FF5630"
  icon_url     = "https://hubertgauthier5.atlassian.net/images/icons/priorities/highest_new.svg"
}

resource "atlassian_jira_field" "cp_story_points" {
  name = "tf-e2e-cp-sp-%[1]s"
  type = "com.atlassian.jira.plugin.system.customfieldtypes:float"
}

resource "atlassian_jira_field" "cp_sprint_goal" {
  name = "tf-e2e-cp-sg-%[1]s"
  type = "com.atlassian.jira.plugin.system.customfieldtypes:textfield"
}

resource "atlassian_jira_field" "cp_release_notes" {
  name = "tf-e2e-cp-rn-%[1]s"
  type = "com.atlassian.jira.plugin.system.customfieldtypes:textarea"
}

resource "atlassian_jira_group" "cp" {
  name = "tf-e2e-cp-group-%[1]s"
}

resource "atlassian_jira_project_role" "cp" {
  name        = "tf-e2e-cp-role-%[1]s"
  description = "E2E complete project role"
}

# --- Membership ---

resource "atlassian_jira_group_membership" "cp" {
  group_id   = atlassian_jira_group.cp.id
  account_id = data.atlassian_jira_myself.cp.account_id
}

# --- Field Configuration ---

resource "atlassian_jira_field_configuration" "cp" {
  name        = "tf-e2e-cp-fc-%[1]s"
  description = "E2E complete project field configuration"

  field_item {
    field_id    = "summary"
    is_required = true
  }

  field_item {
    field_id    = atlassian_jira_field.cp_story_points.id
    description = "Story point estimate"
  }
}

resource "atlassian_jira_field_configuration_scheme" "cp" {
  name        = "tf-e2e-cp-fcs-%[1]s"
  description = "E2E complete project field configuration scheme"

  mapping {
    issue_type_id          = "default"
    field_configuration_id = atlassian_jira_field_configuration.cp.id
  }
}

# --- Screens ---

resource "atlassian_jira_screen" "cp_create" {
  name = "tf-e2e-cp-scr-create-%[1]s"

  tab {
    name   = "Details"
    fields = ["summary", "description", atlassian_jira_field.cp_story_points.id]
  }
}

resource "atlassian_jira_screen" "cp_edit" {
  name = "tf-e2e-cp-scr-edit-%[1]s"

  tab {
    name   = "Details"
    fields = ["summary", "description", atlassian_jira_field.cp_story_points.id, atlassian_jira_field.cp_release_notes.id]
  }
}

resource "atlassian_jira_screen" "cp_view" {
  name = "tf-e2e-cp-scr-view-%[1]s"

  tab {
    name   = "Details"
    fields = ["summary", "description"]
  }
}

resource "atlassian_jira_screen_scheme" "cp" {
  name              = "tf-e2e-cp-ss-%[1]s"
  default_screen_id = atlassian_jira_screen.cp_view.id
  create_screen_id  = atlassian_jira_screen.cp_create.id
  edit_screen_id    = atlassian_jira_screen.cp_edit.id
  view_screen_id    = atlassian_jira_screen.cp_view.id
}

resource "atlassian_jira_issue_type_screen_scheme" "cp" {
  name = "tf-e2e-cp-itss-%[1]s"

  mapping {
    issue_type_id    = "default"
    screen_scheme_id = atlassian_jira_screen_scheme.cp.id
  }
}

# --- Standalone Statuses (test the status resource type independently) ---

resource "atlassian_jira_status" "cp_open" {
  name            = "tf-e2e-cp-st-open-%[1]s"
  description     = "Standalone open status"
  status_category = "TODO"
  scope_type      = "GLOBAL"
}

resource "atlassian_jira_status" "cp_in_progress" {
  name            = "tf-e2e-cp-st-inprog-%[1]s"
  description     = "Standalone in-progress status"
  status_category = "IN_PROGRESS"
  scope_type      = "GLOBAL"
  depends_on      = [atlassian_jira_status.cp_open]
}

resource "atlassian_jira_status" "cp_in_review" {
  name            = "tf-e2e-cp-st-inrev-%[1]s"
  description     = "Standalone in-review status"
  status_category = "IN_PROGRESS"
  scope_type      = "GLOBAL"
  depends_on      = [atlassian_jira_status.cp_in_progress]
}

resource "atlassian_jira_status" "cp_done" {
  name            = "tf-e2e-cp-st-done-%[1]s"
  description     = "Standalone done status"
  status_category = "DONE"
  scope_type      = "GLOBAL"
  depends_on      = [atlassian_jira_status.cp_in_review]
}

# --- Workflow (creates its own inline statuses with distinct names) ---

resource "atlassian_jira_workflow" "cp" {
  name        = "tf-e2e-cp-wf-%[1]s"
  description = "E2E complete project workflow"

  depends_on = [atlassian_jira_status.cp_done]

  status {
    name             = "CpOpen%[1]s"
    status_reference = "cp_open"
    status_category  = "TODO"
  }

  status {
    name             = "CpInProgress%[1]s"
    status_reference = "cp_in_progress"
    status_category  = "IN_PROGRESS"
  }

  status {
    name             = "CpInReview%[1]s"
    status_reference = "cp_in_review"
    status_category  = "IN_PROGRESS"
  }

  status {
    name             = "CpDone%[1]s"
    status_reference = "cp_done"
    status_category  = "DONE"
  }

  transition {
    name                = "Create"
    to_status_reference = "cp_open"
    type                = "initial"
  }

  transition {
    name                  = "Start Progress"
    from_status_reference = "cp_open"
    to_status_reference   = "cp_in_progress"
    type                  = "directed"
  }

  transition {
    name                  = "Submit for Review"
    from_status_reference = "cp_in_progress"
    to_status_reference   = "cp_in_review"
    type                  = "directed"
  }

  transition {
    name                  = "Approve"
    from_status_reference = "cp_in_review"
    to_status_reference   = "cp_done"
    type                  = "directed"
  }

  transition {
    name                  = "Reject"
    from_status_reference = "cp_in_review"
    to_status_reference   = "cp_in_progress"
    type                  = "directed"
  }
}

resource "atlassian_jira_workflow_scheme" "cp" {
  name             = "tf-e2e-cp-ws-%[1]s"
  description      = "E2E complete project workflow scheme"
  default_workflow = atlassian_jira_workflow.cp.name

  mapping {
    issue_type_id = atlassian_jira_issue_type.cp.id
    workflow_name = atlassian_jira_workflow.cp.name
  }
}

# --- Project ---

resource "atlassian_jira_project" "cp" {
  key                  = %[2]q
  name                 = "TF E2E CP %[1]s"
  project_type_key     = "software"
  project_template_key = "com.pyxis.greenhopper.jira:gh-simplified-scrum-classic"
  lead_account_id      = data.atlassian_jira_myself.cp.account_id
  description          = "E2E complete project test"
  assignee_type        = "PROJECT_LEAD"

  # Ensure project is destroyed BEFORE governance schemes so associations
  # are removed with the project, allowing scheme deletion to succeed.
  depends_on = [
    atlassian_jira_permission_scheme.cp,
    atlassian_jira_notification_scheme.cp,
    atlassian_jira_security_scheme.cp,
  ]
}

# --- Governance: Permission, Notification, Security Schemes ---

resource "atlassian_jira_permission_scheme" "cp" {
  name        = "tf-e2e-cp-perm-%[1]s"
  description = "E2E complete project permission scheme"

  permission {
    permission  = "ADMINISTER_PROJECTS"
    holder_type = "projectLead"
  }

  permission {
    permission  = "BROWSE_PROJECTS"
    holder_type = "anyone"
  }
}

resource "atlassian_jira_project_permission_scheme" "cp" {
  project_key = atlassian_jira_project.cp.key
  scheme_id   = atlassian_jira_permission_scheme.cp.id
}

resource "atlassian_jira_notification_scheme" "cp" {
  name        = "tf-e2e-cp-notif-%[1]s"
  description = "E2E complete project notification scheme"

  notification {
    event_id          = "1"
    notification_type = "CurrentAssignee"
  }

  notification {
    event_id          = "1"
    notification_type = "Reporter"
  }
}

resource "atlassian_jira_project_notification_scheme" "cp" {
  project_key = atlassian_jira_project.cp.key
  scheme_id   = atlassian_jira_notification_scheme.cp.id
}

resource "atlassian_jira_security_scheme" "cp" {
  name        = "tf-e2e-cp-sec-%[1]s"
  description = "E2E complete project security scheme"

  level {
    name       = "Confidential"
    is_default = true

    member {
      type = "reporter"
    }
  }
}

resource "atlassian_jira_project_security_scheme" "cp" {
  project_key = atlassian_jira_project.cp.key
  scheme_id   = atlassian_jira_security_scheme.cp.id
}

# --- Project Role Actor ---

resource "atlassian_jira_project_role_actor" "cp" {
  project_key = atlassian_jira_project.cp.key
  role_id     = atlassian_jira_project_role.cp.id
  actor_type  = "group"
  actor_value = atlassian_jira_group.cp.id
}
`, suffix, rKey)
}

func TestAccE2E_completeProject(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rKey := fmt.Sprintf("TFC%s", suffix[:4])

	config := testAccE2ECompleteProjectConfig(suffix)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Apply full config and verify all 29 resource instances
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Myself data source
					resource.TestCheckResourceAttrSet("data.atlassian_jira_myself.cp", "account_id"),

					// Issue Type
					resource.TestCheckResourceAttrSet("atlassian_jira_issue_type.cp", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type.cp", "name", fmt.Sprintf("tf-e2e-cp-it-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type.cp", "hierarchy_level", "0"),

					// Priority
					resource.TestCheckResourceAttrSet("atlassian_jira_priority.cp", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_priority.cp", "name", fmt.Sprintf("tf-e2e-cp-pri-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_priority.cp", "status_color", "#FF5630"),

					// Fields (×3)
					resource.TestCheckResourceAttrSet("atlassian_jira_field.cp_story_points", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_field.cp_story_points", "name", fmt.Sprintf("tf-e2e-cp-sp-%s", suffix)),
					resource.TestCheckResourceAttrSet("atlassian_jira_field.cp_sprint_goal", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_field.cp_sprint_goal", "name", fmt.Sprintf("tf-e2e-cp-sg-%s", suffix)),
					resource.TestCheckResourceAttrSet("atlassian_jira_field.cp_release_notes", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_field.cp_release_notes", "name", fmt.Sprintf("tf-e2e-cp-rn-%s", suffix)),

					// Group
					resource.TestCheckResourceAttrSet("atlassian_jira_group.cp", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_group.cp", "name", fmt.Sprintf("tf-e2e-cp-group-%s", suffix)),

					// Project Role
					resource.TestCheckResourceAttrSet("atlassian_jira_project_role.cp", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_project_role.cp", "name", fmt.Sprintf("tf-e2e-cp-role-%s", suffix)),

					// Group Membership
					resource.TestCheckResourceAttrSet("atlassian_jira_group_membership.cp", "group_id"),
					resource.TestCheckResourceAttrSet("atlassian_jira_group_membership.cp", "account_id"),

					// Field Configuration
					resource.TestCheckResourceAttrSet("atlassian_jira_field_configuration.cp", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration.cp", "name", fmt.Sprintf("tf-e2e-cp-fc-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration.cp", "field_item.#", "2"),

					// Field Configuration Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_field_configuration_scheme.cp", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration_scheme.cp", "name", fmt.Sprintf("tf-e2e-cp-fcs-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration_scheme.cp", "mapping.#", "1"),

					// Screens (×3)
					resource.TestCheckResourceAttrSet("atlassian_jira_screen.cp_create", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_screen.cp_create", "name", fmt.Sprintf("tf-e2e-cp-scr-create-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_screen.cp_create", "tab.#", "1"),
					resource.TestCheckResourceAttrSet("atlassian_jira_screen.cp_edit", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_screen.cp_edit", "name", fmt.Sprintf("tf-e2e-cp-scr-edit-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_screen.cp_edit", "tab.#", "1"),
					resource.TestCheckResourceAttrSet("atlassian_jira_screen.cp_view", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_screen.cp_view", "name", fmt.Sprintf("tf-e2e-cp-scr-view-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_screen.cp_view", "tab.#", "1"),

					// Screen Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_screen_scheme.cp", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_screen_scheme.cp", "name", fmt.Sprintf("tf-e2e-cp-ss-%s", suffix)),
					resource.TestCheckResourceAttrSet("atlassian_jira_screen_scheme.cp", "default_screen_id"),

					// Issue Type Screen Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_issue_type_screen_scheme.cp", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type_screen_scheme.cp", "name", fmt.Sprintf("tf-e2e-cp-itss-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type_screen_scheme.cp", "mapping.#", "1"),

					// Standalone Statuses (×4)
					resource.TestCheckResourceAttrSet("atlassian_jira_status.cp_open", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_status.cp_open", "name", fmt.Sprintf("tf-e2e-cp-st-open-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_status.cp_open", "status_category", "TODO"),
					resource.TestCheckResourceAttrSet("atlassian_jira_status.cp_in_progress", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_status.cp_in_progress", "status_category", "IN_PROGRESS"),
					resource.TestCheckResourceAttrSet("atlassian_jira_status.cp_in_review", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_status.cp_in_review", "status_category", "IN_PROGRESS"),
					resource.TestCheckResourceAttrSet("atlassian_jira_status.cp_done", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_status.cp_done", "status_category", "DONE"),

					// Workflow
					resource.TestCheckResourceAttrSet("atlassian_jira_workflow.cp", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_workflow.cp", "name", fmt.Sprintf("tf-e2e-cp-wf-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_workflow.cp", "status.#", "4"),
					resource.TestCheckResourceAttr("atlassian_jira_workflow.cp", "transition.#", "5"),

					// Workflow Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_workflow_scheme.cp", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_workflow_scheme.cp", "name", fmt.Sprintf("tf-e2e-cp-ws-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_workflow_scheme.cp", "default_workflow", fmt.Sprintf("tf-e2e-cp-wf-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_workflow_scheme.cp", "mapping.#", "1"),

					// Project
					resource.TestCheckResourceAttrSet("atlassian_jira_project.cp", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_project.cp", "key", rKey),
					resource.TestCheckResourceAttr("atlassian_jira_project.cp", "name", fmt.Sprintf("TF E2E CP %s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_project.cp", "project_type_key", "software"),

					// Permission Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_permission_scheme.cp", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_permission_scheme.cp", "name", fmt.Sprintf("tf-e2e-cp-perm-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_permission_scheme.cp", "permission.#", "2"),

					// Project Permission Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_project_permission_scheme.cp", "scheme_id"),

					// Notification Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_notification_scheme.cp", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_notification_scheme.cp", "name", fmt.Sprintf("tf-e2e-cp-notif-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_notification_scheme.cp", "notification.#", "2"),

					// Project Notification Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_project_notification_scheme.cp", "scheme_id"),

					// Security Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_security_scheme.cp", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_security_scheme.cp", "name", fmt.Sprintf("tf-e2e-cp-sec-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_security_scheme.cp", "level.#", "1"),

					// Project Security Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_project_security_scheme.cp", "scheme_id"),

					// Project Role Actor
					resource.TestCheckResourceAttrSet("atlassian_jira_project_role_actor.cp", "project_key"),
					resource.TestCheckResourceAttrSet("atlassian_jira_project_role_actor.cp", "role_id"),
					resource.TestCheckResourceAttr("atlassian_jira_project_role_actor.cp", "actor_type", "group"),
				),
			},
			// Step 2: Re-apply same config with PlanOnly to verify idempotency
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
