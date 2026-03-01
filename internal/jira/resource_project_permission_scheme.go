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
	_ resource.Resource                = &ProjectPermissionSchemeResource{}
	_ resource.ResourceWithImportState = &ProjectPermissionSchemeResource{}
)

// ProjectPermissionSchemeResource assigns a permission scheme to a project.
type ProjectPermissionSchemeResource struct {
	client *atlassian.Client
}

// ProjectPermissionSchemeResourceModel describes the resource data model.
type ProjectPermissionSchemeResourceModel struct {
	ProjectKey types.String `tfsdk:"project_key"`
	SchemeID   types.String `tfsdk:"scheme_id"`
}

// projectPermissionSchemeAssignRequest is the body for PUT /rest/api/3/project/{key}/permissionscheme.
type projectPermissionSchemeAssignRequest struct {
	ID int `json:"id"`
}

// projectPermissionSchemeResponse is the response from GET /rest/api/3/project/{key}/permissionscheme.
type projectPermissionSchemeResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// NewProjectPermissionSchemeResource returns a new resource factory function.
func NewProjectPermissionSchemeResource() resource.Resource {
	return &ProjectPermissionSchemeResource{}
}

func (r *ProjectPermissionSchemeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_project_permission_scheme"
}

func (r *ProjectPermissionSchemeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Assigns a permission scheme to a Jira project.",
		Attributes: map[string]schema.Attribute{
			"project_key": schema.StringAttribute{
				Description: "The key of the project.",
				Required:    true,
			},
			"scheme_id": schema.StringAttribute{
				Description: "The ID of the permission scheme to assign.",
				Required:    true,
			},
		},
	}
}

func (r *ProjectPermissionSchemeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectPermissionSchemeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectPermissionSchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diags := r.assignScheme(ctx, plan.ProjectKey.ValueString(), plan.SchemeID.ValueString()); diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectPermissionSchemeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectPermissionSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp projectPermissionSchemeResponse
	err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/project/%s/permissionscheme", state.ProjectKey.ValueString()), &apiResp)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Project Permission Scheme",
			"An error occurred while calling the Jira API.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state.SchemeID = types.StringValue(fmt.Sprintf("%d", apiResp.ID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProjectPermissionSchemeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectPermissionSchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diags := r.assignScheme(ctx, plan.ProjectKey.ValueString(), plan.SchemeID.ValueString()); diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectPermissionSchemeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectPermissionSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Assign the default permission scheme (ID 0) on delete
	if diags := r.assignScheme(ctx, state.ProjectKey.ValueString(), "0"); diags != nil {
		resp.Diagnostics.Append(diags...)
	}
}

func (r *ProjectPermissionSchemeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by project key
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_key"), req.ID)...)

	// Read current scheme
	var apiResp projectPermissionSchemeResponse
	err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/project/%s/permissionscheme", req.ID), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Import Project Permission Scheme",
			"An error occurred while calling the Jira API.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scheme_id"), fmt.Sprintf("%d", apiResp.ID))...)
}

func (r *ProjectPermissionSchemeResource) assignScheme(ctx context.Context, projectKey, schemeID string) diag.Diagnostics {
	var schemeIDInt int
	if _, err := fmt.Sscanf(schemeID, "%d", &schemeIDInt); err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"Invalid Scheme ID",
			fmt.Sprintf("The scheme_id %q is not a valid integer: %s", schemeID, err.Error()),
		)
		return diags
	}

	apiReq := projectPermissionSchemeAssignRequest{ID: schemeIDInt}
	if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/project/%s/permissionscheme", projectKey), apiReq, nil); err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"Unable to Assign Permission Scheme",
			"An error occurred while calling the Jira API to assign the permission scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return diags
	}
	return nil
}
