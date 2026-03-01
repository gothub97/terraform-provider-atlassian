package jira_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJiraGroup_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_group" "test" {
  name = %q
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_group.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_group.test", "name", rName),
				),
			},
			// ImportState
			{
				ResourceName:      "atlassian_jira_group.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"self"},
			},
		},
	})
}

func TestAccJiraGroupMembership_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
data "atlassian_jira_myself" "test" {}

resource "atlassian_jira_group" "test" {
  name = %q
}

resource "atlassian_jira_group_membership" "test" {
  group_id   = atlassian_jira_group.test.id
  account_id = data.atlassian_jira_myself.test.account_id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_group_membership.test", "group_id"),
					resource.TestCheckResourceAttrSet("atlassian_jira_group_membership.test", "account_id"),
				),
			},
		},
	})
}

func TestAccJiraGroups_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + `
data "atlassian_jira_groups" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.atlassian_jira_groups.test", "groups.#"),
				),
			},
		},
	})
}

func TestAccJiraUsers_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + `
data "atlassian_jira_users" "test" {
  query = "."
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.atlassian_jira_users.test", "users.#"),
				),
			},
		},
	})
}
