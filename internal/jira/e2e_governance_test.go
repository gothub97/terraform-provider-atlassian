package jira_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccE2EGovernanceConfig provisions a complete governance chain:
// group → membership → role actor → permission scheme → notification scheme → security scheme → project associations
func testAccE2EGovernanceConfig(suffix string) string {
	rKey := fmt.Sprintf("TFG%s", suffix[:4])
	return testAccProviderConfig + fmt.Sprintf(`
data "atlassian_jira_myself" "e2e" {}

# --- Group + Membership ---

resource "atlassian_jira_group" "e2e" {
  name = "tf-e2e-group-%[1]s"
}

resource "atlassian_jira_group_membership" "e2e" {
  group_id   = atlassian_jira_group.e2e.id
  account_id = data.atlassian_jira_myself.e2e.account_id
}

# --- Project (needed for role actors and scheme associations) ---

resource "atlassian_jira_project" "e2e_gov" {
  key                  = %[2]q
  name                 = "TF E2E Gov %[1]s"
  project_type_key     = "software"
  project_template_key = "com.pyxis.greenhopper.jira:gh-simplified-scrum-classic"
  lead_account_id      = data.atlassian_jira_myself.e2e.account_id
}

# --- Permission Scheme ---

resource "atlassian_jira_permission_scheme" "e2e" {
  name        = "tf-e2e-perm-%[1]s"
  description = "E2E governance test"

  permission {
    permission  = "ADMINISTER_PROJECTS"
    holder_type = "projectLead"
  }

  permission {
    permission  = "BROWSE_PROJECTS"
    holder_type = "anyone"
  }
}

resource "atlassian_jira_project_permission_scheme" "e2e" {
  project_key = atlassian_jira_project.e2e_gov.key
  scheme_id   = atlassian_jira_permission_scheme.e2e.id
}

# --- Notification Scheme ---

resource "atlassian_jira_notification_scheme" "e2e" {
  name        = "tf-e2e-notif-%[1]s"
  description = "E2E governance test"

  notification {
    event_id          = "1"
    notification_type = "CurrentAssignee"
  }

  notification {
    event_id          = "1"
    notification_type = "Reporter"
  }
}

resource "atlassian_jira_project_notification_scheme" "e2e" {
  project_key = atlassian_jira_project.e2e_gov.key
  scheme_id   = atlassian_jira_notification_scheme.e2e.id
}

# --- Security Scheme ---

resource "atlassian_jira_security_scheme" "e2e" {
  name        = "tf-e2e-sec-%[1]s"
  description = "E2E governance test"

  level {
    name       = "Confidential"
    is_default = true

    member {
      type = "reporter"
    }
  }
}

# --- Data Sources ---

data "atlassian_jira_groups" "e2e" {
  depends_on = [atlassian_jira_group.e2e]
}

data "atlassian_jira_users" "e2e" {
  query = "."
}

data "atlassian_jira_project_roles" "e2e" {}

data "atlassian_jira_permission_schemes" "e2e" {
  depends_on = [atlassian_jira_permission_scheme.e2e]
}

data "atlassian_jira_notification_schemes" "e2e" {
  depends_on = [atlassian_jira_notification_scheme.e2e]
}

data "atlassian_jira_security_schemes" "e2e" {
  depends_on = [atlassian_jira_security_scheme.e2e]
}
`, suffix, rKey)
}

func TestAccE2E_governance(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	config := testAccE2EGovernanceConfig(suffix)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Apply full governance chain
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Group
					resource.TestCheckResourceAttrSet("atlassian_jira_group.e2e", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_group.e2e", "name", fmt.Sprintf("tf-e2e-group-%s", suffix)),

					// Group Membership
					resource.TestCheckResourceAttrSet("atlassian_jira_group_membership.e2e", "group_id"),
					resource.TestCheckResourceAttrSet("atlassian_jira_group_membership.e2e", "account_id"),

					// Project
					resource.TestCheckResourceAttrSet("atlassian_jira_project.e2e_gov", "id"),

					// Permission Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_permission_scheme.e2e", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_permission_scheme.e2e", "permission.#", "2"),

					// Project Permission Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_project_permission_scheme.e2e", "scheme_id"),

					// Notification Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_notification_scheme.e2e", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_notification_scheme.e2e", "notification.#", "2"),

					// Project Notification Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_project_notification_scheme.e2e", "scheme_id"),

					// Security Scheme
					resource.TestCheckResourceAttrSet("atlassian_jira_security_scheme.e2e", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_security_scheme.e2e", "level.#", "1"),

					// Data sources
					resource.TestCheckResourceAttrSet("data.atlassian_jira_groups.e2e", "groups.#"),
					resource.TestCheckResourceAttrSet("data.atlassian_jira_users.e2e", "users.#"),
					resource.TestCheckResourceAttrSet("data.atlassian_jira_project_roles.e2e", "roles.#"),
					resource.TestCheckResourceAttrSet("data.atlassian_jira_permission_schemes.e2e", "schemes.#"),
					resource.TestCheckResourceAttrSet("data.atlassian_jira_notification_schemes.e2e", "schemes.#"),
					resource.TestCheckResourceAttrSet("data.atlassian_jira_security_schemes.e2e", "schemes.#"),
				),
			},
			// Step 2: Verify idempotency
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
