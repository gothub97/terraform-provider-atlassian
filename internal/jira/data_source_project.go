package jira

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ProjectDataSource{}

// ProjectDataSource implements the atlassian_jira_project data source.
type ProjectDataSource struct {
	client *atlassian.Client
}

// ProjectDataSourceModel describes the data source data model.
type ProjectDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Key             types.String `tfsdk:"key"`
	Name            types.String `tfsdk:"name"`
	ProjectTypeKey  types.String `tfsdk:"project_type_key"`
	LeadAccountID   types.String `tfsdk:"lead_account_id"`
	Description     types.String `tfsdk:"description"`
	AssigneeType    types.String `tfsdk:"assignee_type"`
	Self            types.String `tfsdk:"self"`
}

// NewProjectDataSource returns a new data source factory function.
func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

func (d *ProjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_project"
}

func (d *ProjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up a Jira project by its key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The numeric ID of the project.",
				Computed:    true,
			},
			"key": schema.StringAttribute{
				Description: "The project key to look up.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the project.",
				Computed:    true,
			},
			"project_type_key": schema.StringAttribute{
				Description: "The type of the project (e.g., software, business, service_desk).",
				Computed:    true,
			},
			"lead_account_id": schema.StringAttribute{
				Description: "The account ID of the project lead.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the project.",
				Computed:    true,
			},
			"assignee_type": schema.StringAttribute{
				Description: "The default assignee type for the project.",
				Computed:    true,
			},
			"self": schema.StringAttribute{
				Description: "The URL of the project.",
				Computed:    true,
			},
		},
	}
}

func (d *ProjectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ProjectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := config.Key.ValueString()

	var apiResp projectAPIResponse
	err := d.client.Get(ctx, fmt.Sprintf("/rest/api/3/project/%s", key), &apiResp)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.Diagnostics.AddError(
				"Project Not Found",
				fmt.Sprintf("No project found with key %q.", key),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Project",
			"An error occurred while calling the Jira API to read the project.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state := ProjectDataSourceModel{
		ID:             types.StringValue(apiResp.ID),
		Key:            types.StringValue(apiResp.Key),
		Name:           types.StringValue(apiResp.Name),
		ProjectTypeKey: types.StringValue(apiResp.ProjectTypeKey),
		LeadAccountID:  types.StringValue(apiResp.Lead.AccountID),
		Description:    types.StringValue(apiResp.Description),
		AssigneeType:   types.StringValue(apiResp.AssigneeType),
		Self:           types.StringValue(apiResp.Self),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
