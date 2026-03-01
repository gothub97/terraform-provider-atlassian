package jira_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/atlassian/terraform-provider-atlassian/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"atlassian": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func TestMyselfDataSource_AllFields(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/api/3/myself" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"accountId": "5b10ac8d82e05b22cc7d4ef5",
				"accountType": "atlassian",
				"displayName": "John Doe",
				"emailAddress": "john@example.com",
				"active": true,
				"timeZone": "America/New_York",
				"locale": "en_US",
				"self": "https://site.atlassian.net/rest/api/3/user?accountId=5b10ac8d82e05b22cc7d4ef5"
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "atlassian" {
  url       = %q
  email     = "test@example.com"
  api_token = "test-token"
}

data "atlassian_jira_myself" "test" {}
`, ts.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.atlassian_jira_myself.test", "id", "5b10ac8d82e05b22cc7d4ef5"),
					resource.TestCheckResourceAttr("data.atlassian_jira_myself.test", "account_id", "5b10ac8d82e05b22cc7d4ef5"),
					resource.TestCheckResourceAttr("data.atlassian_jira_myself.test", "account_type", "atlassian"),
					resource.TestCheckResourceAttr("data.atlassian_jira_myself.test", "display_name", "John Doe"),
					resource.TestCheckResourceAttr("data.atlassian_jira_myself.test", "email_address", "john@example.com"),
					resource.TestCheckResourceAttr("data.atlassian_jira_myself.test", "active", "true"),
					resource.TestCheckResourceAttr("data.atlassian_jira_myself.test", "time_zone", "America/New_York"),
					resource.TestCheckResourceAttr("data.atlassian_jira_myself.test", "locale", "en_US"),
					resource.TestCheckResourceAttr("data.atlassian_jira_myself.test", "self", "https://site.atlassian.net/rest/api/3/user?accountId=5b10ac8d82e05b22cc7d4ef5"),
				),
			},
		},
	})
}

func TestMyselfDataSource_InactiveUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/api/3/myself" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"accountId": "abc123",
				"accountType": "app",
				"displayName": "Bot User",
				"emailAddress": "",
				"active": false,
				"timeZone": "UTC",
				"locale": "en_GB",
				"self": "https://site.atlassian.net/rest/api/3/user?accountId=abc123"
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "atlassian" {
  url       = %q
  email     = "test@example.com"
  api_token = "test-token"
}

data "atlassian_jira_myself" "test" {}
`, ts.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.atlassian_jira_myself.test", "account_id", "abc123"),
					resource.TestCheckResourceAttr("data.atlassian_jira_myself.test", "account_type", "app"),
					resource.TestCheckResourceAttr("data.atlassian_jira_myself.test", "display_name", "Bot User"),
					resource.TestCheckResourceAttr("data.atlassian_jira_myself.test", "active", "false"),
					resource.TestCheckResourceAttr("data.atlassian_jira_myself.test", "time_zone", "UTC"),
					resource.TestCheckResourceAttr("data.atlassian_jira_myself.test", "locale", "en_GB"),
				),
			},
		},
	})
}

func TestMyselfDataSource_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "atlassian" {
  url       = %q
  email     = "test@example.com"
  api_token = "bad-token"
}

data "atlassian_jira_myself" "test" {}
`, ts.URL),
				ExpectError: regexp.MustCompile(`(?i)unable to authenticate|error`),
			},
		},
	})
}
