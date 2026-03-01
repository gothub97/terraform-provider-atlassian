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

var _ datasource.DataSource = &ScreensDataSource{}

// ScreensDataSource implements the atlassian_jira_screens data source.
type ScreensDataSource struct {
	client *atlassian.Client
}

// ScreensDataSourceModel describes the data source data model.
type ScreensDataSourceModel struct {
	Screens []ScreenEntryModel `tfsdk:"screens"`
}

// ScreenEntryModel describes a single screen in the screens list.
type ScreenEntryModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

// screenItem represents a single screen from the paginated API.
type screenItem struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// NewScreensDataSource returns a new data source factory function.
func NewScreensDataSource() datasource.DataSource {
	return &ScreensDataSource{}
}

func (d *ScreensDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_screens"
}

func (d *ScreensDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve all Jira screens.",
		Attributes: map[string]schema.Attribute{
			"screens": schema.ListNestedAttribute{
				Description: "The list of screens.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the screen.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the screen.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the screen.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ScreensDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ScreensDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	allScreens, err := atlassian.Paginate[screenItem](ctx, d.client, "/rest/api/3/screens")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Screens",
			"An error occurred while calling the Jira API to read screens.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state := ScreensDataSourceModel{
		Screens: make([]ScreenEntryModel, len(allScreens)),
	}

	for i, s := range allScreens {
		state.Screens[i] = ScreenEntryModel{
			ID:          types.StringValue(strconv.Itoa(s.ID)),
			Name:        types.StringValue(s.Name),
			Description: types.StringValue(s.Description),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
