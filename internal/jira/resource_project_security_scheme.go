package jira

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &ProjectSecuritySchemeResource{}
	_ resource.ResourceWithImportState = &ProjectSecuritySchemeResource{}
)

// ProjectSecuritySchemeResource assigns an issue security scheme to a project.
type ProjectSecuritySchemeResource struct {
	client *atlassian.Client
}

// ProjectSecuritySchemeResourceModel describes the resource data model.
type ProjectSecuritySchemeResourceModel struct {
	ProjectKey types.String `tfsdk:"project_key"`
	SchemeID   types.String `tfsdk:"scheme_id"`
}

// projectSecuritySchemeAssignRequest is the body for PUT /rest/api/3/issuesecurityschemes/project.
type projectSecuritySchemeAssignRequest struct {
	SchemeID  string `json:"schemeId"`
	ProjectID string `json:"projectId"`
}

// projectIssueSecuritySchemeResponse is the response from GET /rest/api/3/project/{key}/issuesecuritylevelscheme.
type projectIssueSecuritySchemeResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NewProjectSecuritySchemeResource returns a new resource factory function.
func NewProjectSecuritySchemeResource() resource.Resource {
	return &ProjectSecuritySchemeResource{}
}

func (r *ProjectSecuritySchemeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_project_security_scheme"
}

func (r *ProjectSecuritySchemeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Assigns an issue security scheme to a Jira project. This operation is asynchronous.",
		Attributes: map[string]schema.Attribute{
			"project_key": schema.StringAttribute{
				Description: "The key of the project.",
				Required:    true,
			},
			"scheme_id": schema.StringAttribute{
				Description: "The ID of the issue security scheme to assign.",
				Required:    true,
			},
		},
	}
}

func (r *ProjectSecuritySchemeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectSecuritySchemeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectSecuritySchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diags := r.assignScheme(ctx, plan.ProjectKey.ValueString(), plan.SchemeID.ValueString()); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectSecuritySchemeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectSecuritySchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp projectIssueSecuritySchemeResponse
	err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/project/%s/issuesecuritylevelscheme", state.ProjectKey.ValueString()), &apiResp)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Project Security Scheme",
			"An error occurred while calling the Jira API.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// If no scheme is assigned, the ID will be empty or "0"
	if apiResp.ID == "" || apiResp.ID == "0" {
		resp.State.RemoveResource(ctx)
		return
	}

	state.SchemeID = types.StringValue(apiResp.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProjectSecuritySchemeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectSecuritySchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diags := r.assignScheme(ctx, plan.ProjectKey.ValueString(), plan.SchemeID.ValueString()); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectSecuritySchemeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectSecuritySchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remove the security scheme by assigning empty/none
	if diags := r.assignScheme(ctx, state.ProjectKey.ValueString(), ""); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

func (r *ProjectSecuritySchemeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by project key
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_key"), req.ID)...)

	var apiResp projectIssueSecuritySchemeResponse
	err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/project/%s/issuesecuritylevelscheme", req.ID), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Import Project Security Scheme",
			"An error occurred while calling the Jira API.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scheme_id"), apiResp.ID)...)
}

func (r *ProjectSecuritySchemeResource) assignScheme(ctx context.Context, projectKey, schemeID string) diag.Diagnostics {
	var diags diag.Diagnostics

	// First, resolve the project key to project ID
	var project struct {
		ID string `json:"id"`
	}
	if err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/project/%s", projectKey), &project); err != nil {
		diags.AddError(
			"Unable to Read Project",
			"An error occurred while resolving the project key to an ID.\n\n"+
				"Error: "+err.Error(),
		)
		return diags
	}

	assignReq := projectSecuritySchemeAssignRequest{
		SchemeID:  schemeID,
		ProjectID: project.ID,
	}

	// This is an async operation — the API returns a task ID
	var taskResp struct {
		TaskID string `json:"taskId"`
	}
	if err := r.client.Put(ctx, "/rest/api/3/issuesecurityschemes/project", assignReq, &taskResp); err != nil {
		diags.AddError(
			"Unable to Assign Security Scheme",
			"An error occurred while calling the Jira API to assign the security scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return diags
	}

	// Wait for the async task to complete
	if taskResp.TaskID != "" {
		if err := r.client.WaitForTask(ctx, fmt.Sprintf("/rest/api/3/task/%s", taskResp.TaskID), 0); err != nil {
			diags.AddError(
				"Security Scheme Assignment Failed",
				"The async assignment of the security scheme failed.\n\n"+
					"Error: "+err.Error(),
			)
		}
	}

	return diags
}
