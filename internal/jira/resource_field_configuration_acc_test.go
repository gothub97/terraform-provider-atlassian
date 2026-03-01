package jira_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJiraFieldConfiguration_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_field_configuration" "test" {
  name        = %q
  description = "Acceptance test field configuration"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_field_configuration.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration.test", "description", "Acceptance test field configuration"),
				),
			},
		},
	})
}

func TestAccJiraFieldConfiguration_withFieldItems(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_field_configuration" "test" {
  name        = %q
  description = "Field config with items"

  field_item {
    field_id    = "summary"
    is_required = true
  }
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_field_configuration.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration.test", "field_item.#", "1"),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration.test", "field_item.0.field_id", "summary"),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration.test", "field_item.0.is_required", "true"),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration.test", "field_item.0.is_hidden", "false"),
				),
			},
		},
	})
}

func TestAccJiraFieldConfiguration_update(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	rNameUpdated := fmt.Sprintf("tf-acc-%s-upd", acctest.RandStringFromCharSet(6, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_field_configuration" "test" {
  name        = %q
  description = "Original description"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration.test", "description", "Original description"),
				),
			},
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_field_configuration" "test" {
  name        = %q
  description = "Updated description"
}
`, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration.test", "name", rNameUpdated),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccJiraFieldConfiguration_import(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_field_configuration" "test" {
  name        = %q
  description = "Import test field configuration"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_field_configuration.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration.test", "name", rName),
				),
			},
			{
				ResourceName:            "atlassian_jira_field_configuration.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"field_item"},
			},
		},
	})
}
