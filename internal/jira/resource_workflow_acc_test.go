package jira_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJiraWorkflow_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	suffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_workflow" "test" {
  name        = %q
  description = "Acceptance test workflow"

  status {
    name             = "TfOpen%s"
    status_reference = "open"
    status_category  = "TODO"
  }

  status {
    name             = "TfInProgress%s"
    status_reference = "in_progress"
    status_category  = "IN_PROGRESS"
  }

  status {
    name             = "TfDone%s"
    status_reference = "done"
    status_category  = "DONE"
  }

  transition {
    name                = "Create"
    to_status_reference = "open"
    type                = "initial"
  }

  transition {
    name                  = "Start Work"
    from_status_reference = "open"
    to_status_reference   = "in_progress"
    type                  = "directed"
  }

  transition {
    name                  = "Complete"
    from_status_reference = "in_progress"
    to_status_reference   = "done"
    type                  = "directed"
  }
}
`, rName, suffix, suffix, suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_workflow.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_workflow.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_workflow.test", "description", "Acceptance test workflow"),
					resource.TestCheckResourceAttr("atlassian_jira_workflow.test", "status.#", "3"),
					resource.TestCheckResourceAttr("atlassian_jira_workflow.test", "transition.#", "3"),
				),
			},
		},
	})
}

func TestAccJiraWorkflow_update(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	suffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_workflow" "test" {
  name        = %q
  description = "Original description"

  status {
    name             = "TfNew%s"
    status_reference = "new"
    status_category  = "TODO"
  }

  status {
    name             = "TfClosed%s"
    status_reference = "closed"
    status_category  = "DONE"
  }

  transition {
    name                = "Create"
    to_status_reference = "new"
    type                = "initial"
  }

  transition {
    name                  = "Finish"
    from_status_reference = "new"
    to_status_reference   = "closed"
    type                  = "directed"
  }
}
`, rName, suffix, suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_workflow.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_workflow.test", "description", "Original description"),
				),
			},
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_workflow" "test" {
  name        = %q
  description = "Updated description"

  status {
    name             = "TfNew%s"
    status_reference = "new"
    status_category  = "TODO"
  }

  status {
    name             = "TfClosed%s"
    status_reference = "closed"
    status_category  = "DONE"
  }

  transition {
    name                = "Create"
    to_status_reference = "new"
    type                = "initial"
  }

  transition {
    name                  = "Finish"
    from_status_reference = "new"
    to_status_reference   = "closed"
    type                  = "directed"
  }
}
`, rName, suffix, suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_workflow.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_workflow.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccJiraWorkflow_import(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	suffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_workflow" "test" {
  name        = %q
  description = "Import test workflow"

  status {
    name             = "TfPending%s"
    status_reference = "pending"
    status_category  = "TODO"
  }

  status {
    name             = "TfResolved%s"
    status_reference = "resolved"
    status_category  = "DONE"
  }

  transition {
    name                = "Create"
    to_status_reference = "pending"
    type                = "initial"
  }

  transition {
    name                  = "Resolve"
    from_status_reference = "pending"
    to_status_reference   = "resolved"
    type                  = "directed"
  }
}
`, rName, suffix, suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_workflow.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_workflow.test", "name", rName),
				),
			},
			{
				ResourceName:            "atlassian_jira_workflow.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"status", "transition"},
			},
		},
	})
}
