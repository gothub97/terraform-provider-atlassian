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
	_ resource.Resource                = &ProjectNotificationSchemeResource{}
	_ resource.ResourceWithImportState = &ProjectNotificationSchemeResource{}
)

// ProjectNotificationSchemeResource assigns a notification scheme to a project.
type ProjectNotificationSchemeResource struct {
	client *atlassian.Client
}

// ProjectNotificationSchemeResourceModel describes the resource data model.
type ProjectNotificationSchemeResourceModel struct {
	ProjectKey types.String `tfsdk:"project_key"`
	SchemeID   types.String `tfsdk:"scheme_id"`
}

// projectNotificationSchemeResponse is the response from GET /rest/api/3/project/{key}/notificationscheme.
type projectNotificationSchemeResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// NewProjectNotificationSchemeResource returns a new resource factory function.
func NewProjectNotificationSchemeResource() resource.Resource {
	return &ProjectNotificationSchemeResource{}
}

func (r *ProjectNotificationSchemeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_project_notification_scheme"
}

func (r *ProjectNotificationSchemeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Assigns a notification scheme to a Jira project.",
		Attributes: map[string]schema.Attribute{
			"project_key": schema.StringAttribute{
				Description: "The key of the project.",
				Required:    true,
			},
			"scheme_id": schema.StringAttribute{
				Description: "The ID of the notification scheme to assign.",
				Required:    true,
			},
		},
	}
}

func (r *ProjectNotificationSchemeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectNotificationSchemeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectNotificationSchemeResourceModel
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

func (r *ProjectNotificationSchemeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectNotificationSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp projectNotificationSchemeResponse
	err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/project/%s/notificationscheme", state.ProjectKey.ValueString()), &apiResp)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Project Notification Scheme",
			"An error occurred while calling the Jira API.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state.SchemeID = types.StringValue(fmt.Sprintf("%d", apiResp.ID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProjectNotificationSchemeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectNotificationSchemeResourceModel
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

func (r *ProjectNotificationSchemeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectNotificationSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remove the notification scheme assignment by setting it to the default (ID -1 means none)
	// Jira assigns the default notification scheme when notificationScheme is set to -1
	if diags := r.assignScheme(ctx, state.ProjectKey.ValueString(), "-1"); diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

func (r *ProjectNotificationSchemeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by project key
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_key"), req.ID)...)

	var apiResp projectNotificationSchemeResponse
	err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/project/%s/notificationscheme", req.ID), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Import Project Notification Scheme",
			"An error occurred while calling the Jira API.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scheme_id"), fmt.Sprintf("%d", apiResp.ID))...)
}

func (r *ProjectNotificationSchemeResource) assignScheme(ctx context.Context, projectKey, schemeID string) diag.Diagnostics {
	var diags diag.Diagnostics

	var schemeIDInt int
	if _, err := fmt.Sscanf(schemeID, "%d", &schemeIDInt); err != nil {
		diags.AddError(
			"Invalid Scheme ID",
			fmt.Sprintf("The scheme_id %q is not a valid integer: %s", schemeID, err.Error()),
		)
		return diags
	}

	// Use project update endpoint to set the notification scheme
	updateReq := map[string]any{
		"notificationScheme": schemeIDInt,
	}
	if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/project/%s", projectKey), updateReq, nil); err != nil {
		diags.AddError(
			"Unable to Assign Notification Scheme",
			"An error occurred while calling the Jira API to assign the notification scheme.\n\n"+
				"Error: "+err.Error(),
		)
	}
	return diags
}
