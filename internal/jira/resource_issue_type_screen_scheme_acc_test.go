package jira_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJiraIssueTypeScreenScheme_basic(t *testing.T) {
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
  name              = "%s-ss"
  default_screen_id = atlassian_jira_screen.test.id
}

resource "atlassian_jira_issue_type_screen_scheme" "test" {
  name = %q

  mapping {
    issue_type_id    = "default"
    screen_scheme_id = atlassian_jira_screen_scheme.test.id
  }
}
`, rName, rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_issue_type_screen_scheme.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type_screen_scheme.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type_screen_scheme.test", "mapping.#", "1"),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type_screen_scheme.test", "mapping.0.issue_type_id", "default"),
					resource.TestCheckResourceAttrSet("atlassian_jira_issue_type_screen_scheme.test", "mapping.0.screen_scheme_id"),
				),
			},
		},
	})
}

func TestAccJiraIssueTypeScreenScheme_update(t *testing.T) {
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
  name              = "%s-ss"
  default_screen_id = atlassian_jira_screen.test.id
}

resource "atlassian_jira_issue_type_screen_scheme" "test" {
  name = %q

  mapping {
    issue_type_id    = "default"
    screen_scheme_id = atlassian_jira_screen_scheme.test.id
  }
}
`, rName, rName, rName),
				Check: resource.TestCheckResourceAttr("atlassian_jira_issue_type_screen_scheme.test", "name", rName),
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
  name              = "%s-ss"
  default_screen_id = atlassian_jira_screen.test.id
}

resource "atlassian_jira_issue_type_screen_scheme" "test" {
  name        = %q
  description = "Updated ITSS"

  mapping {
    issue_type_id    = "default"
    screen_scheme_id = atlassian_jira_screen_scheme.test.id
  }
}
`, rName, rName, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_issue_type_screen_scheme.test", "name", rNameUpdated),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type_screen_scheme.test", "description", "Updated ITSS"),
				),
			},
		},
	})
}

func TestAccJiraIssueTypeScreenScheme_import(t *testing.T) {
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
  name              = "%s-ss"
  default_screen_id = atlassian_jira_screen.test.id
}

resource "atlassian_jira_issue_type_screen_scheme" "test" {
  name = %q

  mapping {
    issue_type_id    = "default"
    screen_scheme_id = atlassian_jira_screen_scheme.test.id
  }
}
`, rName, rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_issue_type_screen_scheme.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type_screen_scheme.test", "name", rName),
				),
			},
			{
				ResourceName:            "atlassian_jira_issue_type_screen_scheme.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"mapping"},
			},
		},
	})
}
