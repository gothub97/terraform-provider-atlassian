package provider

import (
	"context"
	"os"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/atlassian/terraform-provider-atlassian/internal/jira"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &AtlassianProvider{}

// AtlassianProvider defines the provider implementation.
type AtlassianProvider struct {
	version string
}

// AtlassianProviderModel describes the provider data model.
type AtlassianProviderModel struct {
	URL      types.String `tfsdk:"url"`
	Email    types.String `tfsdk:"email"`
	APIToken types.String `tfsdk:"api_token"`
}

// New returns a new provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AtlassianProvider{
			version: version,
		}
	}
}

func (p *AtlassianProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "atlassian"
	resp.Version = p.version
}

func (p *AtlassianProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Atlassian provider allows you to manage Jira Cloud resources.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "The base URL of your Atlassian instance (e.g., https://mysite.atlassian.net). Can also be set via ATLASSIAN_URL environment variable.",
				Optional:    true,
			},
			"email": schema.StringAttribute{
				Description: "The email address for authentication. Can also be set via ATLASSIAN_EMAIL environment variable.",
				Optional:    true,
			},
			"api_token": schema.StringAttribute{
				Description: "The API token for authentication. Can also be set via ATLASSIAN_API_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *AtlassianProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config AtlassianProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := os.Getenv("ATLASSIAN_URL")
	email := os.Getenv("ATLASSIAN_EMAIL")
	apiToken := os.Getenv("ATLASSIAN_API_TOKEN")

	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}
	if !config.Email.IsNull() {
		email = config.Email.ValueString()
	}
	if !config.APIToken.IsNull() {
		apiToken = config.APIToken.ValueString()
	}

	if url == "" {
		resp.Diagnostics.AddError(
			"Missing URL",
			"The provider cannot create the Atlassian API client as there is a missing or empty value for the Atlassian URL. "+
				"Set the url value in the configuration or use the ATLASSIAN_URL environment variable.",
		)
	}
	if email == "" {
		resp.Diagnostics.AddError(
			"Missing Email",
			"The provider cannot create the Atlassian API client as there is a missing or empty value for the Atlassian email. "+
				"Set the email value in the configuration or use the ATLASSIAN_EMAIL environment variable.",
		)
	}
	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing API Token",
			"The provider cannot create the Atlassian API client as there is a missing or empty value for the Atlassian API token. "+
				"Set the api_token value in the configuration or use the ATLASSIAN_API_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client := atlassian.NewClient(url, email, apiToken)

	// Validate credentials by calling GET /rest/api/3/myself
	var myself map[string]any
	if err := client.Get(ctx, "/rest/api/3/myself", &myself); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Authenticate",
			"The provider was unable to authenticate with the Atlassian API. "+
				"Please verify your URL, email, and API token are correct.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *AtlassianProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		jira.NewProjectResource,
		jira.NewStatusResource,
		jira.NewIssueTypeResource,
		jira.NewPriorityResource,
	}
}

func (p *AtlassianProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		jira.NewMyselfDataSource,
		jira.NewProjectDataSource,
		jira.NewStatusesDataSource,
		jira.NewIssueTypesDataSource,
		jira.NewPrioritiesDataSource,
	}
}
