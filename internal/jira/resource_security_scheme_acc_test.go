package jira_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJiraSecurityScheme_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_security_scheme" "test" {
  name        = %q
  description = "Acceptance test security scheme"

  level {
    name        = "Confidential"
    description = "Only team leads"
    is_default  = true

    member {
      type = "reporter"
    }
  }
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_jira_security_scheme.test", "id"),
					resource.TestCheckResourceAttr("atlassian_jira_security_scheme.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_jira_security_scheme.test", "description", "Acceptance test security scheme"),
					resource.TestCheckResourceAttr("atlassian_jira_security_scheme.test", "level.#", "1"),
					resource.TestCheckResourceAttr("atlassian_jira_security_scheme.test", "level.0.name", "Confidential"),
					resource.TestCheckResourceAttr("atlassian_jira_security_scheme.test", "level.0.is_default", "true"),
					resource.TestCheckResourceAttr("atlassian_jira_security_scheme.test", "level.0.member.#", "1"),
				),
			},
			// ImportState
			{
				ResourceName:      "atlassian_jira_security_scheme.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccJiraSecurityScheme_update(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_security_scheme" "test" {
  name = %q

  level {
    name       = "Level1"
    is_default = true

    member {
      type = "reporter"
    }
  }
}
`, rName),
				Check: resource.TestCheckResourceAttr("atlassian_jira_security_scheme.test", "level.#", "1"),
			},
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_jira_security_scheme" "test" {
  name        = "%s-updated"
  description = "Updated"

  level {
    name       = "Level1"
    is_default = true

    member {
      type = "reporter"
    }
  }

  level {
    name       = "Level2"
    is_default = false

    member {
      type = "reporter"
    }
  }
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_jira_security_scheme.test", "name", rName+"-updated"),
					resource.TestCheckResourceAttr("atlassian_jira_security_scheme.test", "level.#", "2"),
				),
			},
		},
	})
}

func TestAccJiraSecuritySchemes_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + `
data "atlassian_jira_security_schemes" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.atlassian_jira_security_schemes.test", "schemes.#"),
				),
			},
		},
	})
}
