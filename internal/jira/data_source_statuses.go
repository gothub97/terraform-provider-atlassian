package jira

import (
	"context"
	"fmt"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &StatusesDataSource{}

// StatusesDataSource implements the atlassian_jira_statuses data source.
type StatusesDataSource struct {
	client *atlassian.Client
}

// StatusesDataSourceModel describes the data source data model.
type StatusesDataSourceModel struct {
	ProjectID types.String         `tfsdk:"project_id"`
	Statuses  []StatusEntryModel   `tfsdk:"statuses"`
}

// StatusEntryModel describes a single status in the statuses list.
type StatusEntryModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	StatusCategory types.String `tfsdk:"status_category"`
	ScopeType      types.String `tfsdk:"scope_type"`
	ScopeProjectID types.String `tfsdk:"scope_project_id"`
}

// NewStatusesDataSource returns a new data source factory function.
func NewStatusesDataSource() datasource.DataSource {
	return &StatusesDataSource{}
}

func (d *StatusesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_statuses"
}

func (d *StatusesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve all Jira statuses, optionally filtered by project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Description: "Optional project ID to filter statuses by project scope.",
				Optional:    true,
			},
			"statuses": schema.ListNestedAttribute{
				Description: "The list of statuses.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the status.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the status.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the status.",
							Computed:    true,
						},
						"status_category": schema.StringAttribute{
							Description: "The category of the status (TODO, IN_PROGRESS, or DONE).",
							Computed:    true,
						},
						"scope_type": schema.StringAttribute{
							Description: "The scope type of the status (GLOBAL or PROJECT).",
							Computed:    true,
						},
						"scope_project_id": schema.StringAttribute{
							Description: "The project ID of the status scope, if applicable.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *StatusesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StatusesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config StatusesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	basePath := "/rest/api/3/statuses/search"
	if !config.ProjectID.IsNull() && config.ProjectID.ValueString() != "" {
		basePath = fmt.Sprintf("%s?projectId=%s", basePath, config.ProjectID.ValueString())
	}

	allStatuses, err := atlassian.Paginate[statusAPIResponse](ctx, d.client, basePath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Statuses",
			"An error occurred while calling the Jira API to read statuses.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state := StatusesDataSourceModel{
		ProjectID: config.ProjectID,
		Statuses:  make([]StatusEntryModel, len(allStatuses)),
	}

	for i, s := range allStatuses {
		entry := StatusEntryModel{
			ID:             types.StringValue(s.ID),
			Name:           types.StringValue(s.Name),
			Description:    types.StringValue(s.Description),
			StatusCategory: types.StringValue(s.StatusCategory),
			ScopeType:      types.StringValue(s.Scope.Type),
		}
		if s.Scope.Project != nil {
			entry.ScopeProjectID = types.StringValue(s.Scope.Project.ID)
		} else {
			entry.ScopeProjectID = types.StringNull()
		}
		state.Statuses[i] = entry
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
