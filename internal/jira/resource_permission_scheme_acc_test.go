package jira_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJiraPermissionScheme_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_permission_scheme" "test" {
  name        = %q
  description = "Acceptance test scheme"

  permission {
    permission  = "ADMINISTER_PROJECTS"
    holder_type = "projectLead"
  }
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_permission_scheme.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_permission_scheme.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_permission_scheme.test", "description", "Acceptance test scheme"),
					resource.TestCheckResourceAttr("atlassian_jira_permission_scheme.test", "permission.#", "1"),
				),
			},
			// ImportState
			{
				ResourceName:      "atlassian_jira_permission_scheme.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccJiraPermissionScheme_update(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_permission_scheme" "test" {
  name = %q

  permission {
    permission  = "ADMINISTER_PROJECTS"
    holder_type = "projectLead"
  }
}
`, rName),
				Check: resource.TestCheckResourceAttr("atlassian_jira_permission_scheme.test", "permission.#", "1"),
			},
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_permission_scheme" "test" {
  name        = "%s-updated"
  description = "Updated description"

  permission {
    permission  = "ADMINISTER_PROJECTS"
    holder_type = "projectLead"
  }

  permission {
    permission  = "BROWSE_PROJECTS"
    holder_type = "anyone"
  }
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_permission_scheme.test", "name", rName+"-updated"),
					resource.TestCheckResourceAttr("atlassian_jira_permission_scheme.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("atlassian_jira_permission_scheme.test", "permission.#", "2"),
				),
			},
		},
	})
}

func TestAccJiraPermissionSchemes_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + `
data "atlassian_jira_permission_schemes" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.atlassian_jira_permission_schemes.test", "schemes.#"),
				),
			},
		},
	})
}

func TestAccJiraProjectPermissionScheme_basic(t *testing.T) {
	rKey := fmt.Sprintf("TFACC%s", acctest.RandStringFromCharSet(4, "ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
	rName := fmt.Sprintf("TF Acc %s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	sName := fmt.Sprintf("tf-acc-ps-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
data "atlassian_jira_myself" "test" {}

resource "atlassian_jira_project" "test" {
  key                  = %q
  name                 = %q
  project_type_key     = "software"
  project_template_key = "com.pyxis.greenhopper.jira:gh-simplified-scrum-classic"
  lead_account_id      = data.atlassian_jira_myself.test.account_id
}

resource "atlassian_jira_permission_scheme" "test" {
  name = %q

  permission {
    permission  = "ADMINISTER_PROJECTS"
    holder_type = "projectLead"
  }
}

resource "atlassian_jira_project_permission_scheme" "test" {
  project_key = atlassian_jira_project.test.key
  scheme_id   = atlassian_jira_permission_scheme.test.id
}
`, rKey, rName, sName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_project_permission_scheme.test", "project_key", rKey),
					resource.TestCheckResourceAttrSet("atlassian_jira_project_permission_scheme.test", "scheme_id"),
				),
			},
		},
	})
}
