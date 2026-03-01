package jira

import (
	"context"
	"fmt"
	"strconv"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SecuritySchemesDataSource{}

// SecuritySchemesDataSource implements the atlassian_jira_security_schemes data source.
type SecuritySchemesDataSource struct {
	client *atlassian.Client
}

// SecuritySchemesDataSourceModel describes the data source data model.
type SecuritySchemesDataSourceModel struct {
	Schemes []SecuritySchemeEntryModel `tfsdk:"schemes"`
}

// SecuritySchemeEntryModel describes a single security scheme in the list.
type SecuritySchemeEntryModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

// securitySchemesListResponse represents the response from GET /rest/api/3/issuesecurityschemes.
type securitySchemesListResponse struct {
	IssueSecuritySchemes []securitySchemeListEntry `json:"issueSecuritySchemes"`
}

type securitySchemeListEntry struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// NewSecuritySchemesDataSource returns a new data source factory function.
func NewSecuritySchemesDataSource() datasource.DataSource {
	return &SecuritySchemesDataSource{}
}

func (d *SecuritySchemesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_security_schemes"
}

func (d *SecuritySchemesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve all Jira issue security schemes.",
		Attributes: map[string]schema.Attribute{
			"schemes": schema.ListNestedAttribute{
				Description: "The list of issue security schemes.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the security scheme.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the security scheme.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the security scheme.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *SecuritySchemesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*atlassian.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *atlassian.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *SecuritySchemesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var apiResp securitySchemesListResponse
	if err := d.client.Get(ctx, "/rest/api/3/issuesecurityschemes", &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Security Schemes",
			"An error occurred while calling the Jira API to list security schemes.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state := SecuritySchemesDataSourceModel{
		Schemes: make([]SecuritySchemeEntryModel, len(apiResp.IssueSecuritySchemes)),
	}

	for i, s := range apiResp.IssueSecuritySchemes {
		state.Schemes[i] = SecuritySchemeEntryModel{
			ID:          types.StringValue(strconv.Itoa(s.ID)),
			Name:        types.StringValue(s.Name),
			Description: types.StringValue(s.Description),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
