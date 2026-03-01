package jira

import (
	"context"
	"fmt"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &StatusResource{}
	_ resource.ResourceWithImportState = &StatusResource{}
)

// StatusResource implements the atlassian_jira_status resource.
type StatusResource struct {
	client *atlassian.Client
}

// StatusResourceModel describes the resource data model.
type StatusResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	StatusCategory types.String `tfsdk:"status_category"`
	ScopeType      types.String `tfsdk:"scope_type"`
	ScopeProjectID types.String `tfsdk:"scope_project_id"`
}

// statusCreateRequest represents the POST /rest/api/3/statuses request body.
type statusCreateRequest struct {
	Scope    statusScope          `json:"scope"`
	Statuses []statusCreateEntry  `json:"statuses"`
}

type statusScope struct {
	Type    string         `json:"type"`
	Project *scopeProject  `json:"project,omitempty"`
}

type scopeProject struct {
	ID string `json:"id"`
}

type statusCreateEntry struct {
	Name           string `json:"name"`
	StatusCategory string `json:"statusCategory"`
	Description    string `json:"description,omitempty"`
}

// statusUpdateRequest represents the PUT /rest/api/3/statuses request body.
type statusUpdateRequest struct {
	Statuses []statusUpdateEntry `json:"statuses"`
}

type statusUpdateEntry struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	StatusCategory string `json:"statusCategory"`
	Description    string `json:"description,omitempty"`
}

// statusAPIResponse represents a status object returned by the API.
type statusAPIResponse struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	StatusCategory string           `json:"statusCategory"`
	Scope          statusScopeResp  `json:"scope"`
}

type statusScopeResp struct {
	Type    string             `json:"type"`
	Project *scopeProjectResp  `json:"project,omitempty"`
}

type scopeProjectResp struct {
	ID string `json:"id"`
}

// statusSearchResponse represents the paginated response from GET /rest/api/3/statuses/search.
type statusSearchResponse struct {
	Values []statusAPIResponse `json:"values"`
}

// NewStatusResource returns a new resource factory function.
func NewStatusResource() resource.Resource {
	return &StatusResource{}
}

func (r *StatusResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_status"
}

func (r *StatusResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira status.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the status.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the status.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the status.",
				Optional:    true,
			},
			"status_category": schema.StringAttribute{
				Description: "The category of the status. Must be one of: TODO, IN_PROGRESS, DONE.",
				Required:    true,
			},
			"scope_type": schema.StringAttribute{
				Description: "The scope of the status. Must be one of: GLOBAL, PROJECT.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scope_project_id": schema.StringAttribute{
				Description: "The project ID for the status scope. Required when scope_type is PROJECT.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *StatusResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StatusResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StatusResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate status_category
	if !isValidStatusCategory(plan.StatusCategory.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			path.Root("status_category"),
			"Invalid Status Category",
			fmt.Sprintf("status_category must be one of: TODO, IN_PROGRESS, DONE. Got: %s", plan.StatusCategory.ValueString()),
		)
	}

	// Validate scope_type
	if !isValidScopeType(plan.ScopeType.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			path.Root("scope_type"),
			"Invalid Scope Type",
			fmt.Sprintf("scope_type must be one of: GLOBAL, PROJECT. Got: %s", plan.ScopeType.ValueString()),
		)
	}

	// Validate scope_project_id is set when scope_type is PROJECT
	if plan.ScopeType.ValueString() == "PROJECT" && (plan.ScopeProjectID.IsNull() || plan.ScopeProjectID.ValueString() == "") {
		resp.Diagnostics.AddAttributeError(
			path.Root("scope_project_id"),
			"Missing Scope Project ID",
			"scope_project_id is required when scope_type is PROJECT.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Build the scope
	scope := statusScope{
		Type: plan.ScopeType.ValueString(),
	}
	if plan.ScopeType.ValueString() == "PROJECT" {
		scope.Project = &scopeProject{
			ID: plan.ScopeProjectID.ValueString(),
		}
	}

	// Build the create request
	createReq := statusCreateRequest{
		Scope: scope,
		Statuses: []statusCreateEntry{
			{
				Name:           plan.Name.ValueString(),
				StatusCategory: plan.StatusCategory.ValueString(),
				Description:    plan.Description.ValueString(),
			},
		},
	}

	var created []statusAPIResponse
	if err := r.client.Post(ctx, "/rest/api/3/statuses", createReq, &created); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Status",
			"An error occurred while calling the Jira API to create the status.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	if len(created) == 0 {
		resp.Diagnostics.AddError(
			"Unexpected API Response",
			"The Jira API returned an empty array when creating the status.",
		)
		return
	}

	// Map the response to the state
	plan.ID = types.StringValue(created[0].ID)
	plan.Name = types.StringValue(created[0].Name)
	plan.Description = types.StringValue(created[0].Description)
	plan.StatusCategory = types.StringValue(created[0].StatusCategory)
	plan.ScopeType = types.StringValue(created[0].Scope.Type)
	if created[0].Scope.Project != nil {
		plan.ScopeProjectID = types.StringValue(created[0].Scope.Project.ID)
	} else {
		plan.ScopeProjectID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *StatusResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StatusResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var statuses []statusAPIResponse
	getPath := fmt.Sprintf("/rest/api/3/statuses?id=%s", state.ID.ValueString())
	if err := r.client.Get(ctx, getPath, &statuses); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Status",
			"An error occurred while calling the Jira API to read the status.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	if len(statuses) == 0 {
		// Status was deleted outside of Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	status := statuses[0]
	state.ID = types.StringValue(status.ID)
	state.Name = types.StringValue(status.Name)
	state.Description = types.StringValue(status.Description)
	state.StatusCategory = types.StringValue(status.StatusCategory)
	state.ScopeType = types.StringValue(status.Scope.Type)
	if status.Scope.Project != nil {
		state.ScopeProjectID = types.StringValue(status.Scope.Project.ID)
	} else {
		state.ScopeProjectID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *StatusResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StatusResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state StatusResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate status_category
	if !isValidStatusCategory(plan.StatusCategory.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			path.Root("status_category"),
			"Invalid Status Category",
			fmt.Sprintf("status_category must be one of: TODO, IN_PROGRESS, DONE. Got: %s", plan.StatusCategory.ValueString()),
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := statusUpdateRequest{
		Statuses: []statusUpdateEntry{
			{
				ID:             state.ID.ValueString(),
				Name:           plan.Name.ValueString(),
				StatusCategory: plan.StatusCategory.ValueString(),
				Description:    plan.Description.ValueString(),
			},
		},
	}

	if err := r.client.Put(ctx, "/rest/api/3/statuses", updateReq, nil); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Status",
			"An error occurred while calling the Jira API to update the status.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Preserve the ID and ForceNew fields from state
	plan.ID = state.ID
	plan.ScopeType = state.ScopeType
	plan.ScopeProjectID = state.ScopeProjectID

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *StatusResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StatusResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deletePath := fmt.Sprintf("/rest/api/3/statuses?id=%s", state.ID.ValueString())
	if err := r.client.Delete(ctx, deletePath, nil); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Status",
			"An error occurred while calling the Jira API to delete the status.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}
}

func (r *StatusResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func isValidStatusCategory(v string) bool {
	return v == "TODO" || v == "IN_PROGRESS" || v == "DONE"
}

func isValidScopeType(v string) bool {
	return v == "GLOBAL" || v == "PROJECT"
}
