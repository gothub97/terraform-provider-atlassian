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

var _ datasource.DataSource = &PermissionSchemesDataSource{}

// PermissionSchemesDataSource implements the atlassian_jira_permission_schemes data source.
type PermissionSchemesDataSource struct {
	client *atlassian.Client
}

// PermissionSchemesDataSourceModel describes the data source data model.
type PermissionSchemesDataSourceModel struct {
	Schemes []PermissionSchemeEntryModel `tfsdk:"schemes"`
}

// PermissionSchemeEntryModel describes a single permission scheme in the schemes list.
type PermissionSchemeEntryModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

// permissionSchemesListResponse represents the response from GET /rest/api/3/permissionscheme.
type permissionSchemesListResponse struct {
	PermissionSchemes []permissionSchemeListEntry `json:"permissionSchemes"`
}

// permissionSchemeListEntry represents a single permission scheme from the API response.
type permissionSchemeListEntry struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// NewPermissionSchemesDataSource returns a new data source factory function.
func NewPermissionSchemesDataSource() datasource.DataSource {
	return &PermissionSchemesDataSource{}
}

func (d *PermissionSchemesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_permission_schemes"
}

func (d *PermissionSchemesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve all Jira permission schemes.",
		Attributes: map[string]schema.Attribute{
			"schemes": schema.ListNestedAttribute{
				Description: "The list of permission schemes.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the permission scheme.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the permission scheme.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the permission scheme.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *PermissionSchemesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PermissionSchemesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var apiResp permissionSchemesListResponse
	if err := d.client.Get(ctx, "/rest/api/3/permissionscheme", &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Permission Schemes",
			"An error occurred while calling the Jira API to list permission schemes.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state := PermissionSchemesDataSourceModel{
		Schemes: make([]PermissionSchemeEntryModel, len(apiResp.PermissionSchemes)),
	}

	for i, s := range apiResp.PermissionSchemes {
		state.Schemes[i] = PermissionSchemeEntryModel{
			ID:          types.StringValue(strconv.Itoa(s.ID)),
			Name:        types.StringValue(s.Name),
			Description: types.StringValue(s.Description),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
