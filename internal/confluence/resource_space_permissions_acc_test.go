package confluence_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConfluenceSpacePermissions_basic(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)
	spaceKey := fmt.Sprintf("TFACC%s", acctest.RandStringFromCharSet(5, "ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
	spaceName := fmt.Sprintf("tf-acc-space-%s", suffix)
	groupName := fmt.Sprintf("tf-acc-grp-%s", suffix)

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

resource "atlassian_confluence_space_permissions" "test" {
  space_key = atlassian_confluence_space.test.key
  group_id  = atlassian_jira_group.test.id

  permission {
    operation = "read"
    target    = "space"
  }
}
`, groupName, spaceKey, spaceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_confluence_space_permissions.test", "id"),
					resource.TestCheckResourceAttr("atlassian_confluence_space_permissions.test", "space_key", spaceKey),
					resource.TestCheckResourceAttr("atlassian_confluence_space_permissions.test", "permission.#", "1"),
				),
			},
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_group" "test" {
  name = %q
}

resource "atlassian_confluence_space" "test" {
  key  = %q
  name = %q
}

resource "atlassian_confluence_space_permissions" "test" {
  space_key = atlassian_confluence_space.test.key
  group_id  = atlassian_jira_group.test.id

  permission {
    operation = "read"
    target    = "space"
  }

  permission {
    operation = "create"
    target    = "page"
  }
}
`, groupName, spaceKey, spaceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_confluence_space_permissions.test", "permission.#", "2"),
				),
			},
			{
				ResourceName:      "atlassian_confluence_space_permissions.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
