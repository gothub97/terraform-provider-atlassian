package jira_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJiraFields_all(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + `
data "atlassian_jira_fields" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.atlassian_jira_fields.test", "fields.#"),
				),
			},
		},
	})
}

func TestAccJiraFields_customFilter(t *testing.T) {
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

data "atlassian_jira_fields" "test" {
  type = "custom"

  depends_on = [atlassian_jira_field.test]
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.atlassian_jira_fields.test", "fields.#"),
					resource.TestCheckResourceAttr("data.atlassian_jira_fields.test", "type", "custom"),
				),
			},
		},
	})
}
