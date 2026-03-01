package jira

import (
	"context"
	"fmt"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &PrioritiesDataSource{}

// PrioritiesDataSource implements the atlassian_jira_priorities data source.
type PrioritiesDataSource struct {
	client *atlassian.Client
}

// PrioritiesDataSourceModel describes the data source data model.
type PrioritiesDataSourceModel struct {
	Priorities []PriorityEntryModel `tfsdk:"priorities"`
}

// PriorityEntryModel describes a single priority in the priorities list.
type PriorityEntryModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	StatusColor types.String `tfsdk:"status_color"`
	IconURL     types.String `tfsdk:"icon_url"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
}

// PriorityItem represents a single priority from the paginated search API.
type PriorityItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	StatusColor string `json:"statusColor"`
	IconURL     string `json:"iconUrl"`
	IsDefault   bool   `json:"isDefault"`
}

// NewPrioritiesDataSource returns a new data source factory function.
func NewPrioritiesDataSource() datasource.DataSource {
	return &PrioritiesDataSource{}
}

func (d *PrioritiesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_priorities"
}

func (d *PrioritiesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve all Jira priorities.",
		Attributes: map[string]schema.Attribute{
			"priorities": schema.ListNestedAttribute{
				Description: "The list of priorities.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the priority.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the priority.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the priority.",
							Computed:    true,
						},
						"status_color": schema.StringAttribute{
							Description: "The color of the priority status.",
							Computed:    true,
						},
						"icon_url": schema.StringAttribute{
							Description: "The URL of the icon for the priority.",
							Computed:    true,
						},
						"is_default": schema.BoolAttribute{
							Description: "Whether the priority is the default priority.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *PrioritiesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PrioritiesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	allPriorities, err := atlassian.Paginate[PriorityItem](ctx, d.client, "/rest/api/3/priority/search")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Priorities",
			"An error occurred while calling the Jira API to read priorities.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state := PrioritiesDataSourceModel{
		Priorities: make([]PriorityEntryModel, len(allPriorities)),
	}

	for i, p := range allPriorities {
		state.Priorities[i] = PriorityEntryModel{
			ID:          types.StringValue(p.ID),
			Name:        types.StringValue(p.Name),
			Description: types.StringValue(p.Description),
			StatusColor: types.StringValue(p.StatusColor),
			IconURL:     types.StringValue(p.IconURL),
			IsDefault:   types.BoolValue(p.IsDefault),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
