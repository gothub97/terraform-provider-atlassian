package jira

import (
	"context"
	"fmt"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &IssueTypesDataSource{}

// IssueTypesDataSource implements the atlassian_jira_issue_types data source.
type IssueTypesDataSource struct {
	client *atlassian.Client
}

// IssueTypesDataSourceModel describes the data source data model.
type IssueTypesDataSourceModel struct {
	ProjectID  types.String          `tfsdk:"project_id"`
	IssueTypes []IssueTypeDataModel  `tfsdk:"issue_types"`
}

// IssueTypeDataModel describes an individual issue type in the data source output.
type IssueTypeDataModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	HierarchyLevel types.Int64  `tfsdk:"hierarchy_level"`
	Subtask        types.Bool   `tfsdk:"subtask"`
	IconURL        types.String `tfsdk:"icon_url"`
}

// NewIssueTypesDataSource returns a new data source factory function.
func NewIssueTypesDataSource() datasource.DataSource {
	return &IssueTypesDataSource{}
}

func (d *IssueTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_issue_types"
}

func (d *IssueTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve all issue types, optionally filtered by project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Description: "Optional project ID to filter issue types for a specific project.",
				Optional:    true,
			},
			"issue_types": schema.ListNestedAttribute{
				Description: "The list of issue types.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the issue type.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the issue type.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the issue type.",
							Computed:    true,
						},
						"hierarchy_level": schema.Int64Attribute{
							Description: "The hierarchy level of the issue type.",
							Computed:    true,
						},
						"subtask": schema.BoolAttribute{
							Description: "Whether the issue type is a subtask type.",
							Computed:    true,
						},
						"icon_url": schema.StringAttribute{
							Description: "The URL of the issue type icon.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *IssueTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IssueTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config IssueTypesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiPath string
	if !config.ProjectID.IsNull() && !config.ProjectID.IsUnknown() {
		apiPath = fmt.Sprintf("/rest/api/3/issuetype/project?projectId=%s", config.ProjectID.ValueString())
	} else {
		apiPath = "/rest/api/3/issuetype"
	}

	var apiResp []issueTypeAPIResponse
	if err := d.client.Get(ctx, apiPath, &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Issue Types",
			"An error occurred while calling the Jira API to list issue types.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state := IssueTypesDataSourceModel{
		ProjectID:  config.ProjectID,
		IssueTypes: make([]IssueTypeDataModel, len(apiResp)),
	}

	for i, it := range apiResp {
		state.IssueTypes[i] = IssueTypeDataModel{
			ID:             types.StringValue(it.ID),
			Name:           types.StringValue(it.Name),
			Description:    types.StringValue(it.Description),
			HierarchyLevel: types.Int64Value(it.HierarchyLevel),
			Subtask:        types.BoolValue(it.Subtask),
			IconURL:        types.StringValue(it.IconURL),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
