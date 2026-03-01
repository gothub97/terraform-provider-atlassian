package jira

import (
	"context"
	"fmt"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &IssueTypeResource{}
	_ resource.ResourceWithImportState = &IssueTypeResource{}
)

// IssueTypeResource implements the atlassian_jira_issue_type resource.
type IssueTypeResource struct {
	client *atlassian.Client
}

// IssueTypeResourceModel describes the resource data model.
type IssueTypeResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	HierarchyLevel types.Int64  `tfsdk:"hierarchy_level"`
	AvatarID       types.Int64  `tfsdk:"avatar_id"`
	IconURL        types.String `tfsdk:"icon_url"`
	Subtask        types.Bool   `tfsdk:"subtask"`
	Self           types.String `tfsdk:"self"`
}

// issueTypeAPIResponse represents the JSON response from the Jira issue type API.
type issueTypeAPIResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	HierarchyLevel int64  `json:"hierarchyLevel"`
	AvatarID       int64  `json:"avatarId"`
	IconURL        string `json:"iconUrl"`
	Subtask        bool   `json:"subtask"`
	Self           string `json:"self"`
}

// issueTypeCreateRequest represents the request body for creating an issue type.
type issueTypeCreateRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	HierarchyLevel int64  `json:"hierarchyLevel"`
}

// issueTypeUpdateRequest represents the request body for updating an issue type.
type issueTypeUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	AvatarID    *int64 `json:"avatarId,omitempty"`
}

// NewIssueTypeResource returns a new resource factory function.
func NewIssueTypeResource() resource.Resource {
	return &IssueTypeResource{}
}

func (r *IssueTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_issue_type"
}

// nameMaxLengthValidator validates that a string is at most maxLen characters.
type nameMaxLengthValidator struct {
	maxLen int
}

func (v nameMaxLengthValidator) Description(_ context.Context) string {
	return fmt.Sprintf("string length must be at most %d", v.maxLen)
}

func (v nameMaxLengthValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v nameMaxLengthValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	if len(value) > v.maxLen {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Name Length",
			fmt.Sprintf("Name must be at most %d characters, got %d.", v.maxLen, len(value)),
		)
	}
}

// hierarchyLevelValidator validates that hierarchy_level is 0 or -1.
type hierarchyLevelValidator struct{}

func (v hierarchyLevelValidator) Description(_ context.Context) string {
	return "value must be 0 (standard) or -1 (subtask)"
}

func (v hierarchyLevelValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v hierarchyLevelValidator) ValidateInt64(_ context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueInt64()
	if value != 0 && value != -1 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Hierarchy Level",
			fmt.Sprintf("Hierarchy level must be 0 (standard) or -1 (subtask), got %d.", value),
		)
	}
}

func (r *IssueTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira issue type.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the issue type.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the issue type. Maximum 60 characters.",
				Required:    true,
				Validators: []validator.String{
					nameMaxLengthValidator{maxLen: 60},
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the issue type.",
				Optional:    true,
			},
			"hierarchy_level": schema.Int64Attribute{
				Description: "The hierarchy level of the issue type. Use 0 for standard issue types or -1 for subtask issue types.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					hierarchyLevelValidator{},
				},
			},
			"avatar_id": schema.Int64Attribute{
				Description: "The ID of the avatar for the issue type.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"icon_url": schema.StringAttribute{
				Description: "The URL of the issue type icon.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subtask": schema.BoolAttribute{
				Description: "Whether the issue type is a subtask type.",
				Computed:    true,
			},
			"self": schema.StringAttribute{
				Description: "The URL of the issue type.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *IssueTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IssueTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IssueTypeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hierarchyLevel := int64(0)
	if !plan.HierarchyLevel.IsNull() && !plan.HierarchyLevel.IsUnknown() {
		hierarchyLevel = plan.HierarchyLevel.ValueInt64()
	}

	createReq := issueTypeCreateRequest{
		Name:           plan.Name.ValueString(),
		HierarchyLevel: hierarchyLevel,
	}
	if !plan.Description.IsNull() {
		createReq.Description = plan.Description.ValueString()
	}

	var apiResp issueTypeAPIResponse
	if err := r.client.Post(ctx, "/rest/api/3/issuetype", createReq, &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Issue Type",
			"An error occurred while calling the Jira API to create the issue type.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	mapAPIResponseToState(&apiResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IssueTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IssueTypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp issueTypeAPIResponse
	apiPath := fmt.Sprintf("/rest/api/3/issuetype/%s", state.ID.ValueString())
	if err := r.client.Get(ctx, apiPath, &apiResp); err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Issue Type",
			"An error occurred while calling the Jira API to read the issue type.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	mapAPIResponseToState(&apiResp, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *IssueTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan IssueTypeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state IssueTypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := issueTypeUpdateRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() {
		updateReq.Description = plan.Description.ValueString()
	}
	if !plan.AvatarID.IsNull() && !plan.AvatarID.IsUnknown() {
		avatarID := plan.AvatarID.ValueInt64()
		updateReq.AvatarID = &avatarID
	}

	var apiResp issueTypeAPIResponse
	apiPath := fmt.Sprintf("/rest/api/3/issuetype/%s", state.ID.ValueString())
	if err := r.client.Put(ctx, apiPath, updateReq, &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Issue Type",
			"An error occurred while calling the Jira API to update the issue type.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	mapAPIResponseToState(&apiResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IssueTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state IssueTypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPath := fmt.Sprintf("/rest/api/3/issuetype/%s", state.ID.ValueString())
	if err := r.client.Delete(ctx, apiPath, nil); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Issue Type",
			"An error occurred while calling the Jira API to delete the issue type.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}
}

func (r *IssueTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapAPIResponseToState maps the API response fields to the Terraform state model.
func mapAPIResponseToState(apiResp *issueTypeAPIResponse, model *IssueTypeResourceModel) {
	model.ID = types.StringValue(apiResp.ID)
	model.Name = types.StringValue(apiResp.Name)
	model.Description = types.StringValue(apiResp.Description)
	model.HierarchyLevel = types.Int64Value(apiResp.HierarchyLevel)
	model.AvatarID = types.Int64Value(apiResp.AvatarID)
	model.IconURL = types.StringValue(apiResp.IconURL)
	model.Subtask = types.BoolValue(apiResp.Subtask)
	model.Self = types.StringValue(apiResp.Self)
}
