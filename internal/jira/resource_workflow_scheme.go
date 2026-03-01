package jira

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &WorkflowSchemeResource{}
	_ resource.ResourceWithImportState = &WorkflowSchemeResource{}
)

// WorkflowSchemeResource implements the atlassian_jira_workflow_scheme resource.
type WorkflowSchemeResource struct {
	client *atlassian.Client
}

// WorkflowSchemeResourceModel describes the resource data model.
type WorkflowSchemeResourceModel struct {
	ID              types.String                   `tfsdk:"id"`
	Name            types.String                   `tfsdk:"name"`
	Description     types.String                   `tfsdk:"description"`
	DefaultWorkflow types.String                   `tfsdk:"default_workflow"`
	Mappings        []WorkflowSchemeMappingModel   `tfsdk:"mapping"`
}

// WorkflowSchemeMappingModel describes a single issue type to workflow mapping.
type WorkflowSchemeMappingModel struct {
	IssueTypeID  types.String `tfsdk:"issue_type_id"`
	WorkflowName types.String `tfsdk:"workflow_name"`
}

// --- API types ---

// workflowSchemeAPIRequest is the body for POST/PUT /rest/api/3/workflowscheme.
type workflowSchemeAPIRequest struct {
	Name              string            `json:"name"`
	Description       string            `json:"description,omitempty"`
	DefaultWorkflow   string            `json:"defaultWorkflow,omitempty"`
	IssueTypeMappings map[string]string `json:"issueTypeMappings,omitempty"`
}

// workflowSchemeAPIResponse represents the response from the workflow scheme API.
type workflowSchemeAPIResponse struct {
	ID                int               `json:"id"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	DefaultWorkflow   string            `json:"defaultWorkflow"`
	IssueTypeMappings map[string]string `json:"issueTypeMappings"`
}

// workflowSchemeUsagesResponse represents the response from GET /rest/api/3/workflowscheme/{id}/projectUsages.
type workflowSchemeUsagesResponse struct {
	Values []struct {
		ProjectID string `json:"projectId"`
	} `json:"values"`
}

// workflowSchemePublishRequest is the body for POST /rest/api/3/workflowscheme/{id}/draft/publish.
type workflowSchemePublishRequest struct {
	StatusMappings []interface{} `json:"statusMappings"`
}

// NewWorkflowSchemeResource returns a new resource factory function.
func NewWorkflowSchemeResource() resource.Resource {
	return &WorkflowSchemeResource{}
}

func (r *WorkflowSchemeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_workflow_scheme"
}

func (r *WorkflowSchemeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira workflow scheme.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The numeric ID of the workflow scheme.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the workflow scheme.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the workflow scheme.",
				Optional:    true,
			},
			"default_workflow": schema.StringAttribute{
				Description: "The name of the default workflow. Defaults to \"jira\" if not specified.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"mapping": schema.ListNestedBlock{
				Description: "Issue type to workflow mappings.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"issue_type_id": schema.StringAttribute{
							Description: "The issue type ID.",
							Required:    true,
						},
						"workflow_name": schema.StringAttribute{
							Description: "The workflow name to assign to this issue type.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *WorkflowSchemeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*atlassian.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *atlassian.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *WorkflowSchemeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkflowSchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := workflowSchemeAPIRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	// Set default workflow
	if !plan.DefaultWorkflow.IsNull() && !plan.DefaultWorkflow.IsUnknown() {
		apiReq.DefaultWorkflow = plan.DefaultWorkflow.ValueString()
	} else {
		apiReq.DefaultWorkflow = "jira"
	}

	// Build issue type mappings
	if len(plan.Mappings) > 0 {
		apiReq.IssueTypeMappings = make(map[string]string)
		for _, m := range plan.Mappings {
			apiReq.IssueTypeMappings[m.IssueTypeID.ValueString()] = m.WorkflowName.ValueString()
		}
	}

	var apiResp workflowSchemeAPIResponse
	if err := r.client.Post(ctx, "/rest/api/3/workflowscheme", apiReq, &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Workflow Scheme",
			"An error occurred while calling the Jira API to create the workflow scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	mapWorkflowSchemeAPIToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WorkflowSchemeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkflowSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp workflowSchemeAPIResponse
	err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/workflowscheme/%s", state.ID.ValueString()), &apiResp)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Workflow Scheme",
			"An error occurred while calling the Jira API to read the workflow scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	mapWorkflowSchemeAPIToState(&state, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WorkflowSchemeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkflowSchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state WorkflowSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := workflowSchemeAPIRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	if !plan.DefaultWorkflow.IsNull() && !plan.DefaultWorkflow.IsUnknown() {
		apiReq.DefaultWorkflow = plan.DefaultWorkflow.ValueString()
	} else {
		apiReq.DefaultWorkflow = "jira"
	}

	// Build issue type mappings
	apiReq.IssueTypeMappings = make(map[string]string)
	for _, m := range plan.Mappings {
		apiReq.IssueTypeMappings[m.IssueTypeID.ValueString()] = m.WorkflowName.ValueString()
	}

	// Check if the scheme is active (assigned to projects)
	active, diags := r.isSchemeActive(ctx, state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !active {
		// Inactive scheme: direct update
		var apiResp workflowSchemeAPIResponse
		if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/workflowscheme/%s", state.ID.ValueString()), apiReq, &apiResp); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Workflow Scheme",
				"An error occurred while calling the Jira API to update the workflow scheme.\n\n"+
					"Error: "+err.Error(),
			)
			return
		}
		mapWorkflowSchemeAPIToState(&plan, &apiResp)
	} else {
		// Active scheme: draft flow
		diags := r.updateViaDraft(ctx, state.ID.ValueString(), apiReq)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Read back the scheme after publish
		var apiResp workflowSchemeAPIResponse
		if err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/workflowscheme/%s", state.ID.ValueString()), &apiResp); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Workflow Scheme After Update",
				"The workflow scheme was updated via draft but could not be read back.\n\n"+
					"Error: "+err.Error(),
			)
			return
		}
		mapWorkflowSchemeAPIToState(&plan, &apiResp)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WorkflowSchemeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkflowSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/rest/api/3/workflowscheme/%s", state.ID.ValueString()), nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			// Already deleted, nothing to do.
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Workflow Scheme",
			"An error occurred while calling the Jira API to delete the workflow scheme.\n\n"+
				"Error: "+err.Error(),
		)
	}
}

func (r *WorkflowSchemeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// isSchemeActive checks whether the workflow scheme is assigned to any projects.
func (r *WorkflowSchemeResource) isSchemeActive(ctx context.Context, schemeID string) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	var usages workflowSchemeUsagesResponse
	if err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/workflowscheme/%s/projectUsages", schemeID), &usages); err != nil {
		diags.AddError(
			"Unable to Check Workflow Scheme Usage",
			"An error occurred while checking whether the workflow scheme is assigned to projects.\n\n"+
				"Error: "+err.Error(),
		)
		return false, diags
	}

	return len(usages.Values) > 0, diags
}

// updateViaDraft updates an active workflow scheme using the draft/publish flow.
func (r *WorkflowSchemeResource) updateViaDraft(ctx context.Context, schemeID string, apiReq workflowSchemeAPIRequest) diag.Diagnostics {
	var diags diag.Diagnostics

	// Try to get existing draft, or create one
	var draft workflowSchemeAPIResponse
	err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/workflowscheme/%s/draft", schemeID), &draft)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			// No draft exists; updating the draft endpoint will auto-create one
		} else {
			diags.AddError(
				"Unable to Check Workflow Scheme Draft",
				"An error occurred while checking for an existing draft.\n\n"+
					"Error: "+err.Error(),
			)
			return diags
		}
	}

	// Update the draft
	if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/workflowscheme/%s/draft", schemeID), apiReq, nil); err != nil {
		diags.AddError(
			"Unable to Update Workflow Scheme Draft",
			"An error occurred while updating the workflow scheme draft.\n\n"+
				"Error: "+err.Error(),
		)
		return diags
	}

	// Publish the draft
	publishReq := workflowSchemePublishRequest{
		StatusMappings: []interface{}{},
	}

	location, err := r.client.PostWithRedirect(ctx, fmt.Sprintf("/rest/api/3/workflowscheme/%s/draft/publish", schemeID), publishReq)
	if err != nil {
		diags.AddError(
			"Unable to Publish Workflow Scheme Draft",
			"An error occurred while publishing the workflow scheme draft.\n\n"+
				"Error: "+err.Error(),
		)
		return diags
	}

	// If we got a redirect location, poll the task
	if location != "" {
		taskPath, parseErr := extractWorkflowSchemeTaskPath(location)
		if parseErr != nil {
			diags.AddError(
				"Unable to Parse Task URL",
				"An error occurred while parsing the task URL from the publish response.\n\n"+
					"Error: "+parseErr.Error(),
			)
			return diags
		}

		if err := r.client.WaitForTask(ctx, taskPath, 5*time.Minute); err != nil {
			diags.AddError(
				"Publish Task Failed",
				"The async publish task failed.\n\n"+
					"Error: "+err.Error(),
			)
			return diags
		}
	}

	return diags
}

// mapWorkflowSchemeAPIToState maps a workflow scheme API response to the Terraform state model.
func mapWorkflowSchemeAPIToState(state *WorkflowSchemeResourceModel, apiResp *workflowSchemeAPIResponse) {
	state.ID = types.StringValue(fmt.Sprintf("%d", apiResp.ID))
	state.Name = types.StringValue(apiResp.Name)

	if apiResp.Description != "" {
		state.Description = types.StringValue(apiResp.Description)
	} else {
		state.Description = types.StringNull()
	}

	if apiResp.DefaultWorkflow != "" {
		state.DefaultWorkflow = types.StringValue(apiResp.DefaultWorkflow)
	} else {
		state.DefaultWorkflow = types.StringValue("jira")
	}

	// Map issue type mappings
	if len(apiResp.IssueTypeMappings) > 0 {
		state.Mappings = make([]WorkflowSchemeMappingModel, 0, len(apiResp.IssueTypeMappings))
		for issueTypeID, workflowName := range apiResp.IssueTypeMappings {
			state.Mappings = append(state.Mappings, WorkflowSchemeMappingModel{
				IssueTypeID:  types.StringValue(issueTypeID),
				WorkflowName: types.StringValue(workflowName),
			})
		}
	} else {
		state.Mappings = []WorkflowSchemeMappingModel{}
	}
}

// extractWorkflowSchemeTaskPath extracts the path component from a full task URL.
func extractWorkflowSchemeTaskPath(locationURL string) (string, error) {
	u, err := url.Parse(locationURL)
	if err != nil {
		return "", fmt.Errorf("parsing task URL %q: %w", locationURL, err)
	}
	return u.Path, nil
}
