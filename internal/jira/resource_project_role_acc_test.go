package jira_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJiraProjectRole_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_project_role" "test" {
  name        = %q
  description = "Acceptance test role"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_project_role.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_project_role.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_project_role.test", "description", "Acceptance test role"),
				),
			},
			// ImportState
			{
				ResourceName:      "atlassian_jira_project_role.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"self"},
			},
		},
	})
}

func TestAccJiraProjectRole_update(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	rNameUpdated := fmt.Sprintf("tf-acc-%s-upd", acctest.RandStringFromCharSet(6, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_project_role" "test" {
  name        = %q
  description = "Original"
}
`, rName),
				Check: resource.TestCheckResourceAttr("atlassian_jira_project_role.test", "name", rName),
			},
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_project_role" "test" {
  name        = %q
  description = "Updated"
}
`, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_project_role.test", "name", rNameUpdated),
					resource.TestCheckResourceAttr("atlassian_jira_project_role.test", "description", "Updated"),
				),
			},
		},
	})
}

func TestAccJiraProjectRoleActor_basic(t *testing.T) {
	rKey := fmt.Sprintf("TFACC%s", acctest.RandStringFromCharSet(4, "ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
	rName := fmt.Sprintf("TF Acc %s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
data "atlassian_jira_myself" "test" {}

resource "atlassian_jira_project" "test" {
  key                  = %[1]q
  name                 = %[2]q
  project_type_key     = "software"
  project_template_key = "com.pyxis.greenhopper.jira:gh-simplified-scrum-classic"
  lead_account_id      = data.atlassian_jira_myself.test.account_id
}

resource "atlassian_jira_project_role" "test" {
  name        = "tf-acc-role-%[1]s"
  description = "Test role for actor test"
}

resource "atlassian_jira_project_role_actor" "test" {
  project_key = atlassian_jira_project.test.key
  role_id     = atlassian_jira_project_role.test.id
  actor_type  = "user"
  actor_value = data.atlassian_jira_myself.test.account_id
}
`, rKey, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_project_role_actor.test", "project_key"),
					resource.TestCheckResourceAttrSet("atlassian_jira_project_role_actor.test", "role_id"),
					resource.TestCheckResourceAttr("atlassian_jira_project_role_actor.test", "actor_type", "user"),
				),
			},
		},
	})
}

func TestAccJiraProjectRoles_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + `
data "atlassian_jira_project_roles" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.atlassian_jira_project_roles.test", "roles.#"),
				),
			},
		},
	})
}
