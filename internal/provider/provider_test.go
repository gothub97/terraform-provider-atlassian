package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"atlassian": providerserver.NewProtocol6WithError(New("test")()),
}

func TestAccProvider_invalidAuth(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "atlassian" {
  url       = "https://invalid.atlassian.net"
  email     = "invalid@example.com"
  api_token = "invalid-token"
}

data "atlassian_jira_myself" "test" {}
`,
				ExpectError: regexp.MustCompile(`(?i)unable to authenticate|error`),
			},
		},
	})
}
