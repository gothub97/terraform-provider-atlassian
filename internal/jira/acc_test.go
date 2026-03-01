package jira_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccProviderConfig returns a provider block that relies on env vars.
const testAccProviderConfig = `
provider "atlassian" {}
`

// --- Myself Data Source ---

func TestAccJiraMyself_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + `
data "atlassian_jira_myself" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.atlassian_jira_myself.test", "account_id"),
					resource.TestCheckResourceAttrSet("data.atlassian_jira_myself.test", "display_name"),
					resource.TestCheckResourceAttrSet("data.atlassian_jira_myself.test", "email_address"),
					resource.TestCheckResourceAttr("data.atlassian_jira_myself.test", "active", "true"),
					resource.TestCheckResourceAttrSet("data.atlassian_jira_myself.test", "time_zone"),
					resource.TestCheckResourceAttrSet("data.atlassian_jira_myself.test", "locale"),
				),
			},
		},
	})
}

// --- Issue Type Resource ---

func TestAccJiraIssueType_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_issue_type" "test" {
  name            = %q
  description     = "Acceptance test issue type"
  hierarchy_level = 0
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_issue_type.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type.test", "description", "Acceptance test issue type"),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type.test", "hierarchy_level", "0"),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type.test", "subtask", "false"),
					resource.TestCheckResourceAttrSet("atlassian_jira_issue_type.test", "self"),
				),
			},
			// ImportState
			{
				ResourceName:      "atlassian_jira_issue_type.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccJiraIssueType_update(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	rNameUpdated := fmt.Sprintf("tf-acc-%s-upd", acctest.RandStringFromCharSet(6, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_issue_type" "test" {
  name            = %q
  description     = "Original description"
  hierarchy_level = 0
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_issue_type.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type.test", "description", "Original description"),
				),
			},
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_issue_type" "test" {
  name            = %q
  description     = "Updated description"
  hierarchy_level = 0
}
`, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_issue_type.test", "name", rNameUpdated),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccJiraIssueType_subtask(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-sub-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_issue_type" "test" {
  name            = %q
  description     = "A subtask type"
  hierarchy_level = -1
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_issue_type.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type.test", "hierarchy_level", "-1"),
					resource.TestCheckResourceAttr("atlassian_jira_issue_type.test", "subtask", "true"),
				),
			},
		},
	})
}

// --- Priority Resource ---

func TestAccJiraPriority_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_priority" "test" {
  name         = %q
  description  = "Acceptance test priority"
  status_color = "#FF0000"
  icon_url     = "https://hubertgauthier5.atlassian.net/images/icons/priorities/highest_new.svg"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_priority.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_priority.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_priority.test", "description", "Acceptance test priority"),
					resource.TestCheckResourceAttr("atlassian_jira_priority.test", "status_color", "#FF0000"),
					resource.TestCheckResourceAttrSet("atlassian_jira_priority.test", "icon_url"),
					resource.TestCheckResourceAttrSet("atlassian_jira_priority.test", "self"),
				),
			},
			// ImportState
			{
				ResourceName:      "atlassian_jira_priority.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccJiraPriority_update(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	rNameUpdated := fmt.Sprintf("tf-acc-%s-upd", acctest.RandStringFromCharSet(6, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_priority" "test" {
  name         = %q
  description  = "Original priority"
  status_color = "#FF0000"
  icon_url     = "https://hubertgauthier5.atlassian.net/images/icons/priorities/highest_new.svg"
}
`, rName),
				Check: resource.TestCheckResourceAttr("atlassian_jira_priority.test", "name", rName),
			},
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_priority" "test" {
  name         = %q
  description  = "Updated priority"
  status_color = "#00FF00"
  icon_url     = "https://hubertgauthier5.atlassian.net/images/icons/priorities/highest_new.svg"
}
`, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_priority.test", "name", rNameUpdated),
					resource.TestCheckResourceAttr("atlassian_jira_priority.test", "description", "Updated priority"),
					resource.TestCheckResourceAttr("atlassian_jira_priority.test", "status_color", "#00FF00"),
				),
			},
		},
	})
}

// --- Status Resource ---

func TestAccJiraStatus_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_status" "test" {
  name            = %q
  description     = "Acceptance test status"
  status_category = "IN_PROGRESS"
  scope_type      = "GLOBAL"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_status.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_status.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_status.test", "description", "Acceptance test status"),
					resource.TestCheckResourceAttr("atlassian_jira_status.test", "status_category", "IN_PROGRESS"),
					resource.TestCheckResourceAttr("atlassian_jira_status.test", "scope_type", "GLOBAL"),
				),
			},
			// ImportState
			{
				ResourceName:      "atlassian_jira_status.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccJiraStatus_update(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	rNameUpdated := fmt.Sprintf("tf-acc-%s-upd", acctest.RandStringFromCharSet(6, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_status" "test" {
  name            = %q
  description     = "Original status"
  status_category = "TODO"
  scope_type      = "GLOBAL"
}
`, rName),
				Check: resource.TestCheckResourceAttr("atlassian_jira_status.test", "name", rName),
			},
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_status" "test" {
  name            = %q
  description     = "Updated status"
  status_category = "IN_PROGRESS"
  scope_type      = "GLOBAL"
}
`, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_status.test", "name", rNameUpdated),
					resource.TestCheckResourceAttr("atlassian_jira_status.test", "description", "Updated status"),
					resource.TestCheckResourceAttr("atlassian_jira_status.test", "status_category", "IN_PROGRESS"),
				),
			},
		},
	})
}

// --- Project Resource ---

func TestAccJiraProject_basic(t *testing.T) {
	rKey := fmt.Sprintf("TFACC%s", acctest.RandStringFromCharSet(4, "ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
	rName := fmt.Sprintf("TF Acc %s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
data "atlassian_jira_myself" "test" {}

resource "atlassian_jira_project" "test" {
  key              = %q
  name             = %q
  project_type_key = "software"
  lead_account_id  = data.atlassian_jira_myself.test.account_id
  description      = "Acceptance test project"
}
`, rKey, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_project.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_project.test", "key", rKey),
					resource.TestCheckResourceAttr("atlassian_jira_project.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_project.test", "project_type_key", "software"),
					resource.TestCheckResourceAttr("atlassian_jira_project.test", "description", "Acceptance test project"),
					resource.TestCheckResourceAttrSet("atlassian_jira_project.test", "self"),
				),
			},
			// ImportState
			{
				ResourceName:            "atlassian_jira_project.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"project_template_key"},
			},
		},
	})
}

func TestAccJiraProject_update(t *testing.T) {
	rKey := fmt.Sprintf("TFACC%s", acctest.RandStringFromCharSet(4, "ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
	rName := fmt.Sprintf("TF Acc Orig %s", acctest.RandStringFromCharSet(6, acctest.CharSetAlpha))
	rNameUpdated := fmt.Sprintf("TF Acc Upd %s", acctest.RandStringFromCharSet(6, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
data "atlassian_jira_myself" "test" {}

resource "atlassian_jira_project" "test" {
  key              = %q
  name             = %q
  project_type_key = "software"
  lead_account_id  = data.atlassian_jira_myself.test.account_id
  description      = "Original description"
}
`, rKey, rName),
				Check: resource.TestCheckResourceAttr("atlassian_jira_project.test", "name", rName),
			},
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
data "atlassian_jira_myself" "test" {}

resource "atlassian_jira_project" "test" {
  key              = %q
  name             = %q
  project_type_key = "software"
  lead_account_id  = data.atlassian_jira_myself.test.account_id
  description      = "Updated description"
}
`, rKey, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_project.test", "name", rNameUpdated),
					resource.TestCheckResourceAttr("atlassian_jira_project.test", "description", "Updated description"),
				),
			},
		},
	})
}
