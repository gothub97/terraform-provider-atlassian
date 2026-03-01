package jira_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJiraScreenScheme_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_screen" "test" {
  name = "%s-screen"

  tab {
    name   = "Default"
    fields = ["summary"]
  }
}

resource "atlassian_jira_screen_scheme" "test" {
  name              = %q
  default_screen_id = atlassian_jira_screen.test.id
}
`, rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_screen_scheme.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_screen_scheme.test", "name", rName),
					resource.TestCheckResourceAttrSet("atlassian_jira_screen_scheme.test", "default_screen_id"),
				),
			},
		},
	})
}

func TestAccJiraScreenScheme_update(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	rNameUpdated := fmt.Sprintf("tf-acc-%s-upd", acctest.RandStringFromCharSet(6, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_screen" "test" {
  name = "%s-screen"

  tab {
    name   = "Default"
    fields = ["summary"]
  }
}

resource "atlassian_jira_screen_scheme" "test" {
  name              = %q
  default_screen_id = atlassian_jira_screen.test.id
}
`, rName, rName),
				Check: resource.TestCheckResourceAttr("atlassian_jira_screen_scheme.test", "name", rName),
			},
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_screen" "test" {
  name = "%s-screen"

  tab {
    name   = "Default"
    fields = ["summary"]
  }
}

resource "atlassian_jira_screen_scheme" "test" {
  name              = %q
  description       = "Updated screen scheme"
  default_screen_id = atlassian_jira_screen.test.id
}
`, rName, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_screen_scheme.test", "name", rNameUpdated),
					resource.TestCheckResourceAttr("atlassian_jira_screen_scheme.test", "description", "Updated screen scheme"),
				),
			},
		},
	})
}

func TestAccJiraScreenScheme_import(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_screen" "test" {
  name = "%s-screen"

  tab {
    name   = "Default"
    fields = ["summary"]
  }
}

resource "atlassian_jira_screen_scheme" "test" {
  name              = %q
  default_screen_id = atlassian_jira_screen.test.id
}
`, rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_screen_scheme.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_screen_scheme.test", "name", rName),
				),
			},
			{
				ResourceName:      "atlassian_jira_screen_scheme.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
