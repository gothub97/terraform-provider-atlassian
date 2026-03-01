package jira_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJiraScreen_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_screen" "test" {
  name = %q

  tab {
    name   = "Tab 1"
    fields = ["summary"]
  }
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_screen.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_screen.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_screen.test", "tab.#", "1"),
					resource.TestCheckResourceAttr("atlassian_jira_screen.test", "tab.0.name", "Tab 1"),
					resource.TestCheckResourceAttrSet("atlassian_jira_screen.test", "tab.0.id"),
					resource.TestCheckResourceAttr("atlassian_jira_screen.test", "tab.0.fields.#", "1"),
					resource.TestCheckResourceAttr("atlassian_jira_screen.test", "tab.0.fields.0", "summary"),
				),
			},
		},
	})
}

func TestAccJiraScreen_update(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	rNameUpdated := fmt.Sprintf("tf-acc-%s-upd", acctest.RandStringFromCharSet(6, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_screen" "test" {
  name = %q

  tab {
    name   = "Tab 1"
    fields = ["summary"]
  }
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_screen.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_screen.test", "tab.0.name", "Tab 1"),
				),
			},
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_screen" "test" {
  name        = %q
  description = "Updated screen"

  tab {
    name   = "Renamed Tab"
    fields = ["summary", "description"]
  }
}
`, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_screen.test", "name", rNameUpdated),
					resource.TestCheckResourceAttr("atlassian_jira_screen.test", "description", "Updated screen"),
					resource.TestCheckResourceAttr("atlassian_jira_screen.test", "tab.0.name", "Renamed Tab"),
					resource.TestCheckResourceAttr("atlassian_jira_screen.test", "tab.0.fields.#", "2"),
					resource.TestCheckResourceAttr("atlassian_jira_screen.test", "tab.0.fields.0", "summary"),
					resource.TestCheckResourceAttr("atlassian_jira_screen.test", "tab.0.fields.1", "description"),
				),
			},
		},
	})
}

func TestAccJiraScreen_import(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_screen" "test" {
  name = %q

  tab {
    name   = "Tab 1"
    fields = ["summary"]
  }
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_screen.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_screen.test", "name", rName),
				),
			},
			{
				ResourceName:      "atlassian_jira_screen.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
