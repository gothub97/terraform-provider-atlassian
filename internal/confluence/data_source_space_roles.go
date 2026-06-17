package confluence

import (
	"context"
	"fmt"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SpaceRolesDataSource{}

// SpaceRolesDataSource implements the atlassian_confluence_space_roles data source.
type SpaceRolesDataSource struct {
	client *atlassian.Client
}

// SpaceRolesDataSourceModel describes the data source data model.
type SpaceRolesDataSourceModel struct {
	SpaceID types.String          `tfsdk:"space_id"`
	Roles   []SpaceRoleEntryModel `tfsdk:"roles"`
}

// SpaceRoleEntryModel describes a single space role in the roles list.
type SpaceRoleEntryModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
}

// spaceRolesResponse represents the response from GET /wiki/api/v2/space-roles.
type spaceRolesResponse struct {
	Results []spaceRoleAPIEntry `json:"results"`
}

// spaceRoleAPIEntry represents a single space role from the API response.
type spaceRoleAPIEntry struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// NewSpaceRolesDataSource returns a new data source factory function.
func NewSpaceRolesDataSource() datasource.DataSource {
	return &SpaceRolesDataSource{}
}

func (d *SpaceRolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_confluence_space_roles"
}

func (d *SpaceRolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve the space roles available on the Confluence site. Useful for discovering role IDs needed when configuring space role assignments. Space roles are defined site-wide (the same set applies to every space).",
		Attributes: map[string]schema.Attribute{
			"space_id": schema.StringAttribute{
				Description: "Optional ID of a Confluence space. Space roles are site-wide, so this is informational only and does not filter the result.",
				Optional:    true,
			},
			"roles": schema.ListNestedAttribute{
				Description: "The list of roles available for the space.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the space role.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the space role.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of the space role (e.g., USER, GROUP).",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the space role.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *SpaceRolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, errMsg, ok := configureClient(req.ProviderData)
	if !ok {
		if errMsg != "" {
			resp.Diagnostics.AddError("Unexpected Data Source Configure Type", errMsg)
		}
		return
	}
	d.client = client
}

func (d *SpaceRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config SpaceRolesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/space-roles", apiV2Base)

	var apiResp spaceRolesResponse
	if err := d.client.Get(ctx, path, &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Space Roles",
			"An error occurred while calling the Confluence API to list available space roles.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state := SpaceRolesDataSourceModel{
		SpaceID: config.SpaceID,
		Roles:   make([]SpaceRoleEntryModel, len(apiResp.Results)),
	}

	for i, r := range apiResp.Results {
		state.Roles[i] = SpaceRoleEntryModel{
			ID:          types.StringValue(r.ID),
			Name:        types.StringValue(r.Name),
			Type:        types.StringValue(r.Type),
			Description: types.StringValue(r.Description),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
