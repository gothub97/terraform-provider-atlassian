package confluence_test

import (
	"github.com/atlassian/terraform-provider-atlassian/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProviderConfig returns a provider block that relies on env vars
// (ATLASSIAN_URL, ATLASSIAN_EMAIL, ATLASSIAN_API_TOKEN).
const testAccProviderConfig = `
provider "atlassian" {}
`

// testAccProtoV6ProviderFactories are used to instantiate the provider during
// acceptance testing. The factory function is called for each Terraform CLI
// command to create a provider server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"atlassian": providerserver.NewProtocol6WithError(provider.New("test")()),
}
