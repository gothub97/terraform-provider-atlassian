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

var _ datasource.DataSource = &ProjectRolesDataSource{}

// ProjectRolesDataSource implements the atlassian_jira_project_roles data source.
type ProjectRolesDataSource struct {
	client *atlassian.Client
}

// ProjectRolesDataSourceModel describes the data source data model.
type ProjectRolesDataSourceModel struct {
	Roles []ProjectRoleEntryModel `tfsdk:"roles"`
}

// ProjectRoleEntryModel describes a single project role in the roles list.
type ProjectRoleEntryModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

// projectRoleListEntry represents a single project role from GET /rest/api/3/role.
type projectRoleListEntry struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// NewProjectRolesDataSource returns a new data source factory function.
func NewProjectRolesDataSource() datasource.DataSource {
	return &ProjectRolesDataSource{}
}

func (d *ProjectRolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_project_roles"
}

func (d *ProjectRolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve all Jira project roles.",
		Attributes: map[string]schema.Attribute{
			"roles": schema.ListNestedAttribute{
				Description: "The list of project roles.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the project role.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the project role.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the project role.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ProjectRolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectRolesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var apiResp []projectRoleListEntry
	if err := d.client.Get(ctx, "/rest/api/3/role", &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Project Roles",
			"An error occurred while calling the Jira API to list project roles.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state := ProjectRolesDataSourceModel{
		Roles: make([]ProjectRoleEntryModel, len(apiResp)),
	}

	for i, r := range apiResp {
		state.Roles[i] = ProjectRoleEntryModel{
			ID:          types.StringValue(strconv.Itoa(r.ID)),
			Name:        types.StringValue(r.Name),
			Description: types.StringValue(r.Description),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
