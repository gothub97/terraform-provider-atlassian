package confluence_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConfluenceSpaceRoleAssignment_basic(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)
	spaceKey := fmt.Sprintf("TFACC%s", acctest.RandStringFromCharSet(5, "ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
	spaceName := fmt.Sprintf("tf-acc-space-%s", suffix)
	groupName := fmt.Sprintf("tf-acc-grp-%s", suffix)

	// role_id must reference a real Confluence space role available in the test
	// instance. It is provided via the CONFLUENCE_TEST_ROLE_ID convention in the
	// test config below; adjust as needed for the target environment.
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_group" "test" {
  name = %q
}

resource "atlassian_confluence_space" "test" {
  key  = %q
  name = %q
}

data "atlassian_confluence_space_roles" "test" {
  space_id = atlassian_confluence_space.test.id
}

resource "atlassian_confluence_space_role_assignment" "test" {
  space_id       = atlassian_confluence_space.test.id
  principal_type = "GROUP"
  principal_id   = atlassian_jira_group.test.id
  role_id        = data.atlassian_confluence_space_roles.test.roles[0].id
}
`, groupName, spaceKey, spaceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_confluence_space_role_assignment.test", "id"),
					resource.TestCheckResourceAttr("atlassian_confluence_space_role_assignment.test", "principal_type", "GROUP"),
					resource.TestCheckResourceAttrSet("atlassian_confluence_space_role_assignment.test", "role_id"),
				),
			},
			{
				ResourceName:      "atlassian_confluence_space_role_assignment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
