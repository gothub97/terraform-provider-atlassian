package jira_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccE2EConfig returns a Terraform config that provisions the full chain:
// field → field_configuration → field_configuration_scheme → screen → screen_scheme → ITSS → workflow → workflow_scheme
func testAccE2EConfig(suffix string) string {
	return testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_field" "e2e" {
  name = "tf-e2e-field-%s"
  type = "com.atlassian.jira.plugin.system.customfieldtypes:textfield"
}

resource "atlassian_jira_field_configuration" "e2e" {
  name        = "tf-e2e-fc-%s"
  description = "E2E test field configuration"

  field_item {
    field_id    = "summary"
    is_required = true
  }

  field_item {
    field_id    = "description"
    is_required = false
  }
}

resource "atlassian_jira_field_configuration_scheme" "e2e" {
  name        = "tf-e2e-fcs-%s"
  description = "E2E test field configuration scheme"

  mapping {
    issue_type_id          = "default"
    field_configuration_id = atlassian_jira_field_configuration.e2e.id
  }
}

resource "atlassian_jira_screen" "e2e" {
  name = "tf-e2e-screen-%s"

  tab {
    name   = "E2E Tab"
    fields = ["summary", "description"]
  }
}

resource "atlassian_jira_screen_scheme" "e2e" {
  name              = "tf-e2e-ss-%s"
  default_screen_id = atlassian_jira_screen.e2e.id
}

resource "atlassian_jira_issue_type_screen_scheme" "e2e" {
  name = "tf-e2e-itss-%s"

  mapping {
    issue_type_id    = "default"
    screen_scheme_id = atlassian_jira_screen_scheme.e2e.id
  }
}

resource "atlassian_jira_workflow" "e2e" {
  name        = "tf-e2e-wf-%s"
  description = "E2E test workflow"

  status {
    name             = "E2eOpen%s"
    status_reference = "e2e_open"
    status_category  = "TODO"
  }

  status {
    name             = "E2eInProgress%s"
    status_reference = "e2e_in_progress"
    status_category  = "IN_PROGRESS"
  }

  status {
    name             = "E2eDone%s"
    status_reference = "e2e_done"
    status_category  = "DONE"
  }

  transition {
    name                = "Create"
    to_status_reference = "e2e_open"
    type                = "initial"
  }

  transition {
    name                  = "Start Work"
    from_status_reference = "e2e_open"
    to_status_reference   = "e2e_in_progress"
    type                  = "directed"
  }

  transition {
    name                  = "Complete"
    from_status_reference = "e2e_in_progress"
    to_status_reference   = "e2e_done"
    type                  = "directed"
  }

  transition {
    name                  = "Reopen"
    from_status_reference = "e2e_done"
    to_status_reference   = "e2e_open"
    type                  = "directed"
  }
}

resource "atlassian_jira_workflow_scheme" "e2e" {
  name             = "tf-e2e-ws-%s"
  description      = "E2E test workflow scheme"
  default_workflow = atlassian_jira_workflow.e2e.name
}
`, suffix, suffix, suffix, suffix, suffix, suffix, suffix, suffix, suffix, suffix, suffix)
}

func TestAccE2E_fullChain(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	config := testAccE2EConfig(suffix)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Apply full chain and verify all resources
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Field
					resource.TestCheckResourceAttrSet("atlassian_jira_field.e2e", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_field.e2e", "name", fmt.Sprintf("tf-e2e-field-%s", suffix)),

					// Field Configuration
					resource.TestCheckResourceAttrSet("atlassian_jira_field_configuration.e2e", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration.e2e", "name", fmt.Sprintf("tf-e2e-fc-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration.e2e", "field_item.#", "2"),

					// Field Configuration Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_field_configuration_scheme.e2e", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration_scheme.e2e", "name", fmt.Sprintf("tf-e2e-fcs-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration_scheme.e2e", "mapping.#", "1"),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration_scheme.e2e", "mapping.0.issue_type_id", "default"),
					resource.TestCheckResourceAttrSet("atlassian_jira_field_configuration_scheme.e2e", "mapping.0.field_configuration_id"),

					// Screen
					resource.TestCheckResourceAttrSet("atlassian_jira_screen.e2e", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_screen.e2e", "name", fmt.Sprintf("tf-e2e-screen-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_screen.e2e", "tab.#", "1"),
					resource.TestCheckResourceAttr("atlassian_jira_screen.e2e", "tab.0.fields.#", "2"),

					// Screen Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_screen_scheme.e2e", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_screen_scheme.e2e", "name", fmt.Sprintf("tf-e2e-ss-%s", suffix)),
					resource.TestCheckResourceAttrSet("atlassian_jira_screen_scheme.e2e", "default_screen_id"),

					// Issue Type Screen Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_issue_type_screen_scheme.e2e", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type_screen_scheme.e2e", "name", fmt.Sprintf("tf-e2e-itss-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type_screen_scheme.e2e", "mapping.#", "1"),

					// Workflow
					resource.TestCheckResourceAttrSet("atlassian_jira_workflow.e2e", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_workflow.e2e", "name", fmt.Sprintf("tf-e2e-wf-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_workflow.e2e", "status.#", "3"),
					resource.TestCheckResourceAttr("atlassian_jira_workflow.e2e", "transition.#", "4"),

					// Workflow Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_workflow_scheme.e2e", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_workflow_scheme.e2e", "name", fmt.Sprintf("tf-e2e-ws-%s", suffix)),
					resource.TestCheckResourceAttr("atlassian_jira_workflow_scheme.e2e", "default_workflow", fmt.Sprintf("tf-e2e-wf-%s", suffix)),
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
