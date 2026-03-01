package jira

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &ProjectRoleResource{}
	_ resource.ResourceWithImportState = &ProjectRoleResource{}
)

// ProjectRoleResource implements the atlassian_jira_project_role resource.
type ProjectRoleResource struct {
	client *atlassian.Client
}

// ProjectRoleResourceModel describes the resource data model.
type ProjectRoleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Self        types.String `tfsdk:"self"`
}

// projectRoleCreateRequest is the JSON body for POST /rest/api/3/role.
type projectRoleCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// projectRoleUpdateRequest is the JSON body for PUT /rest/api/3/role/{id}.
type projectRoleUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// projectRoleAPIResponse represents the JSON response from the project role API.
type projectRoleAPIResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Self        string `json:"self"`
}

// NewProjectRoleResource returns a new resource factory function.
func NewProjectRoleResource() resource.Resource {
	return &ProjectRoleResource{}
}

func (r *ProjectRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_project_role"
}

func (r *ProjectRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira project role.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The numeric ID of the project role.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the project role.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the project role.",
				Optional:    true,
			},
			"self": schema.StringAttribute{
				Description: "The URL of the project role.",
				Computed:    true,
			},
		},
	}
}

func (r *ProjectRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := projectRoleCreateRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() {
		createReq.Description = plan.Description.ValueString()
	}

	var apiResp projectRoleAPIResponse
	if err := r.client.Post(ctx, "/rest/api/3/role", createReq, &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Project Role",
			"An error occurred while calling the Jira API to create the project role.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	mapProjectRoleAPIToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp projectRoleAPIResponse
	err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/role/%s", state.ID.ValueString()), &apiResp)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Project Role",
			"An error occurred while calling the Jira API to read the project role.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	mapProjectRoleAPIToState(&state, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProjectRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ProjectRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := projectRoleUpdateRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() {
		updateReq.Description = plan.Description.ValueString()
	}

	var apiResp projectRoleAPIResponse
	if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/role/%s", state.ID.ValueString()), updateReq, &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Project Role",
			"An error occurred while calling the Jira API to update the project role.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	mapProjectRoleAPIToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/rest/api/3/role/%s", state.ID.ValueString()), nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			// Already deleted, nothing to do.
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Project Role",
			"An error occurred while calling the Jira API to delete the project role.\n\n"+
				"Error: "+err.Error(),
		)
	}
}

func (r *ProjectRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapProjectRoleAPIToState maps a project role API response to the Terraform state model.
func mapProjectRoleAPIToState(state *ProjectRoleResourceModel, apiResp *projectRoleAPIResponse) {
	state.ID = types.StringValue(fmt.Sprintf("%d", apiResp.ID))
	state.Name = types.StringValue(apiResp.Name)
	state.Self = types.StringValue(apiResp.Self)

	if apiResp.Description != "" {
		state.Description = types.StringValue(apiResp.Description)
	} else {
		state.Description = types.StringNull()
	}
}
