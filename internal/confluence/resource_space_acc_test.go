package confluence_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConfluenceSpace_basic(t *testing.T) {
	rKey := fmt.Sprintf("TFACC%s", strings.ToUpper(acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)))
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_confluence_space" "test" {
  key         = %q
  name        = %q
  description = "Acceptance test space"
}
`, rKey, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("atlassian_confluence_space.test", "id"),
					resource.TestCheckResourceAttr("atlassian_confluence_space.test", "key", rKey),
					resource.TestCheckResourceAttr("atlassian_confluence_space.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_confluence_space.test", "description", "Acceptance test space"),
					resource.TestCheckResourceAttrSet("atlassian_confluence_space.test", "status"),
					resource.TestCheckResourceAttrSet("atlassian_confluence_space.test", "homepage_id"),
					resource.TestCheckResourceAttrSet("atlassian_confluence_space.test", "url"),
				),
			},
			// ImportState (by space key).
			{
				ResourceName:      "atlassian_confluence_space.test",
				ImportState:       true,
				ImportStateId:     rKey,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccConfluenceSpace_update(t *testing.T) {
	rKey := fmt.Sprintf("TFACC%s", strings.ToUpper(acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)))
	rName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_confluence_space" "test" {
  key         = %q
  name        = %q
  description = "Initial description"
}
`, rKey, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_confluence_space.test", "name", rName),
					resource.TestCheckResourceAttr("atlassian_confluence_space.test", "description", "Initial description"),
				),
			},
			{
				Config: testAccProviderConfig + fmt.Sprintf(`
resource "atlassian_confluence_space" "test" {
  key         = %q
  name        = "%s-updated"
  description = "Updated description"
}
`, rKey, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("atlassian_confluence_space.test", "name", rName+"-updated"),
					resource.TestCheckResourceAttr("atlassian_confluence_space.test", "description", "Updated description"),
				),
			},
		},
	})
}
