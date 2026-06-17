package confluence

import (
	"context"
	"fmt"
	"net/url"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SpaceDataSource{}

// SpaceDataSource implements the atlassian_confluence_space data source.
type SpaceDataSource struct {
	client *atlassian.Client
}

// SpaceDataSourceModel describes the data source data model.
type SpaceDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Status      types.String `tfsdk:"status"`
	HomepageID  types.String `tfsdk:"homepage_id"`
	Description types.String `tfsdk:"description"`
}

// NewSpaceDataSource returns a new data source factory function.
func NewSpaceDataSource() datasource.DataSource {
	return &SpaceDataSource{}
}

func (d *SpaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_confluence_space"
}

func (d *SpaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a Confluence space by its key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The numeric ID of the space.",
				Computed:    true,
			},
			"key": schema.StringAttribute{
				Description: "The key of the Confluence space to look up.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the space.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the space (e.g., global, personal).",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the space (e.g., current, archived).",
				Computed:    true,
			},
			"homepage_id": schema.StringAttribute{
				Description: "The ID of the space's homepage.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The plain-text description of the space.",
				Computed:    true,
			},
		},
	}
}

func (d *SpaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, errMsg, ok := configureClient(req.ProviderData)
	if !ok {
		if errMsg != "" {
			resp.Diagnostics.AddError("Unexpected Data Source Configure Type", errMsg)
		}
		return
	}
	d.client = client
}

func (d *SpaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config SpaceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := config.Key.ValueString()
	path := fmt.Sprintf("%s/spaces?keys=%s&description-format=plain", apiV2Base, url.QueryEscape(key))

	var listResp struct {
		Results []spaceV2 `json:"results"`
	}
	if err := d.client.Get(ctx, path, &listResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Space",
			"An error occurred while calling the Confluence API to look up the space.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	if len(listResp.Results) == 0 {
		resp.Diagnostics.AddError(
			"Space Not Found",
			fmt.Sprintf("No Confluence space found with key %q.", key),
		)
		return
	}

	sp := listResp.Results[0]
	state := SpaceDataSourceModel{
		ID:          types.StringValue(sp.ID),
		Key:         types.StringValue(sp.Key),
		Name:        types.StringValue(sp.Name),
		Type:        types.StringValue(sp.Type),
		Status:      types.StringValue(sp.Status),
		HomepageID:  types.StringValue(sp.HomepageID),
		Description: types.StringValue(sp.Description.Plain.Value),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
