package jira

import (
	"context"
	"fmt"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &WorkflowsDataSource{}

// WorkflowsDataSource implements the atlassian_jira_workflows data source.
type WorkflowsDataSource struct {
	client *atlassian.Client
}

// WorkflowsDataSourceModel describes the data source data model.
type WorkflowsDataSourceModel struct {
	ProjectKey types.String              `tfsdk:"project_key"`
	Workflows  []WorkflowDataEntryModel  `tfsdk:"workflows"`
}

// WorkflowDataEntryModel describes a single workflow in the output list.
type WorkflowDataEntryModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

// workflowSearchEntry represents a single workflow entry from GET /rest/api/3/workflow/search.
type workflowSearchEntry struct {
	ID          workflowSearchID `json:"id"`
	Description string           `json:"description"`
}

type workflowSearchID struct {
	EntityID string `json:"entityId"`
	Name     string `json:"name"`
}

// projectIDResponse represents the response from GET /rest/api/3/project/{key} for resolving project key to ID.
type projectIDResponse struct {
	ID string `json:"id"`
}

// NewWorkflowsDataSource returns a new data source factory function.
func NewWorkflowsDataSource() datasource.DataSource {
	return &WorkflowsDataSource{}
}

func (d *WorkflowsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_workflows"
}

func (d *WorkflowsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve all Jira workflows, optionally filtered by project.",
		Attributes: map[string]schema.Attribute{
			"project_key": schema.StringAttribute{
				Description: "Optional project key to filter workflows associated with the project.",
				Optional:    true,
			},
			"workflows": schema.ListNestedAttribute{
				Description: "The list of workflows.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The entity ID (UUID) of the workflow.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the workflow.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the workflow.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *WorkflowsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *WorkflowsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config WorkflowsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	basePath := "/rest/api/3/workflow/search"

	// If project_key is set, resolve it to a project ID and use as filter
	if !config.ProjectKey.IsNull() && config.ProjectKey.ValueString() != "" {
		var project projectIDResponse
		if err := d.client.Get(ctx, fmt.Sprintf("/rest/api/3/project/%s", config.ProjectKey.ValueString()), &project); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Resolve Project Key",
				fmt.Sprintf("An error occurred while resolving the project key %q to a project ID.\n\nError: %s",
					config.ProjectKey.ValueString(), err.Error()),
			)
			return
		}
		basePath = fmt.Sprintf("%s?projectId=%s", basePath, project.ID)
	}

	allWorkflows, err := atlassian.Paginate[workflowSearchEntry](ctx, d.client, basePath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Workflows",
			"An error occurred while calling the Jira API to list workflows.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state := WorkflowsDataSourceModel{
		ProjectKey: config.ProjectKey,
		Workflows:  make([]WorkflowDataEntryModel, len(allWorkflows)),
	}

	for i, wf := range allWorkflows {
		state.Workflows[i] = WorkflowDataEntryModel{
			ID:          types.StringValue(wf.ID.EntityID),
			Name:        types.StringValue(wf.ID.Name),
			Description: types.StringValue(wf.Description),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
