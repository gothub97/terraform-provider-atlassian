package jira

import (
	"context"
	"fmt"
	"net/url"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &GroupsDataSource{}

// GroupsDataSource implements the atlassian_jira_groups data source.
type GroupsDataSource struct {
	client *atlassian.Client
}

// GroupsDataSourceModel describes the data source data model.
type GroupsDataSourceModel struct {
	Query  types.String       `tfsdk:"query"`
	Groups []GroupEntryModel  `tfsdk:"groups"`
}

// GroupEntryModel describes a single group in the groups list.
type GroupEntryModel struct {
	GroupID types.String `tfsdk:"group_id"`
	Name    types.String `tfsdk:"name"`
}

// groupsPickerResponse represents the response from GET /rest/api/3/groups/picker.
type groupsPickerResponse struct {
	Groups []groupPickerEntry `json:"groups"`
}

// groupPickerEntry represents a single group from the groups/picker response.
type groupPickerEntry struct {
	Name    string `json:"name"`
	GroupID string `json:"groupId"`
}

// NewGroupsDataSource returns a new data source factory function.
func NewGroupsDataSource() datasource.DataSource {
	return &GroupsDataSource{}
}

func (d *GroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_groups"
}

func (d *GroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve Jira groups, optionally filtered by a name prefix.",
		Attributes: map[string]schema.Attribute{
			"query": schema.StringAttribute{
				Description: "Optional name prefix filter for groups.",
				Optional:    true,
			},
			"groups": schema.ListNestedAttribute{
				Description: "The list of groups.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"group_id": schema.StringAttribute{
							Description: "The ID of the group.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the group.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *GroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config GroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := "/rest/api/3/groups/picker?maxResults=1000"
	if !config.Query.IsNull() && config.Query.ValueString() != "" {
		path = fmt.Sprintf("/rest/api/3/groups/picker?query=%s&maxResults=1000", url.QueryEscape(config.Query.ValueString()))
	}

	var apiResp groupsPickerResponse
	if err := d.client.Get(ctx, path, &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Groups",
			"An error occurred while calling the Jira API to list groups.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state := GroupsDataSourceModel{
		Query:  config.Query,
		Groups: make([]GroupEntryModel, len(apiResp.Groups)),
	}

	for i, g := range apiResp.Groups {
		state.Groups[i] = GroupEntryModel{
			GroupID: types.StringValue(g.GroupID),
			Name:    types.StringValue(g.Name),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
