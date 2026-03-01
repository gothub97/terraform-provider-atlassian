package jira_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJiraWorkflowScheme_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_workflow_scheme" "test" {
  name        = %q
  description = "Acceptance test workflow scheme"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_workflow_scheme.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_workflow_scheme.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_workflow_scheme.test", "description", "Acceptance test workflow scheme"),
					resource.TestCheckResourceAttr("atlassian_jira_workflow_scheme.test", "default_workflow", "jira"),
				),
			},
		},
	})
}

func TestAccJiraWorkflowScheme_update(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	rNameUpdated := fmt.Sprintf("tf-acc-%s-upd", acctest.RandStringFromCharSet(6, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_workflow_scheme" "test" {
  name        = %q
  description = "Original description"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_workflow_scheme.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_workflow_scheme.test", "description", "Original description"),
				),
			},
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_workflow_scheme" "test" {
  name        = %q
  description = "Updated description"
}
`, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_workflow_scheme.test", "name", rNameUpdated),
					resource.TestCheckResourceAttr("atlassian_jira_workflow_scheme.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccJiraWorkflowScheme_import(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_workflow_scheme" "test" {
  name        = %q
  description = "Import test workflow scheme"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_workflow_scheme.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_workflow_scheme.test", "name", rName),
				),
			},
			{
				ResourceName:            "atlassian_jira_workflow_scheme.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"mapping"},
			},
		},
	})
}
