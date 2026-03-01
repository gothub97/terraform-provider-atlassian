package jira_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJiraFieldConfigurationScheme_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_field_configuration_scheme" "test" {
  name        = %q
  description = "Acceptance test field configuration scheme"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_field_configuration_scheme.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration_scheme.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration_scheme.test", "description", "Acceptance test field configuration scheme"),
				),
			},
		},
	})
}

func TestAccJiraFieldConfigurationScheme_withMappings(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	rFCName := fmt.Sprintf("tf-acc-fc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_field_configuration" "test" {
  name        = %q
  description = "Field config for scheme mapping test"
}

resource "atlassian_jira_field_configuration_scheme" "test" {
  name        = %q
  description = "Scheme with default mapping"

  mapping {
    issue_type_id          = "default"
    field_configuration_id = atlassian_jira_field_configuration.test.id
  }
}
`, rFCName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_field_configuration_scheme.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration_scheme.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration_scheme.test", "description", "Scheme with default mapping"),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration_scheme.test", "mapping.#", "1"),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration_scheme.test", "mapping.0.issue_type_id", "default"),
					resource.TestCheckResourceAttrSet("atlassian_jira_field_configuration_scheme.test", "mapping.0.field_configuration_id"),
				),
			},
		},
	})
}

func TestAccJiraFieldConfigurationScheme_update(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	rNameUpdated := fmt.Sprintf("tf-acc-%s-upd", acctest.RandStringFromCharSet(6, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_field_configuration_scheme" "test" {
  name        = %q
  description = "Original description"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration_scheme.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration_scheme.test", "description", "Original description"),
				),
			},
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_field_configuration_scheme" "test" {
  name        = %q
  description = "Updated description"
}
`, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration_scheme.test", "name", rNameUpdated),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration_scheme.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccJiraFieldConfigurationScheme_import(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_field_configuration_scheme" "test" {
  name        = %q
  description = "Import test field configuration scheme"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_field_configuration_scheme.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_field_configuration_scheme.test", "name", rName),
				),
			},
			{
				ResourceName:            "atlassian_jira_field_configuration_scheme.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"mapping"},
			},
		},
	})
}
