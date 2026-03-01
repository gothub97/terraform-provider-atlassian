package jira_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJiraField_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_field" "test" {
  name = %q
  type = "com.atlassian.jira.plugin.system.customfieldtypes:textfield"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_field.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_field.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_field.test", "type", "com.atlassian.jira.plugin.system.customfieldtypes:textfield"),
				),
			},
		},
	})
}

func TestAccJiraField_update(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	rNameUpdated := fmt.Sprintf("tf-acc-%s-upd", acctest.RandStringFromCharSet(6, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_field" "test" {
  name = %q
  type = "com.atlassian.jira.plugin.system.customfieldtypes:textfield"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_field.test", "name", rName),
				),
			},
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_field" "test" {
  name        = %q
  type        = "com.atlassian.jira.plugin.system.customfieldtypes:textfield"
  description = "Updated description"
}
`, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_field.test", "name", rNameUpdated),
					resource.TestCheckResourceAttr("atlassian_jira_field.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccJiraField_import(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_field" "test" {
  name        = %q
  type        = "com.atlassian.jira.plugin.system.customfieldtypes:textfield"
  description = "Import test field"
  searcher_key = "com.atlassian.jira.plugin.system.customfieldtypes:textsearcher"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_field.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_field.test", "name", rName),
				),
			},
			{
				ResourceName:            "atlassian_jira_field.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"searcher_key"},
			},
		},
	})
}
