package jira

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &GroupResource{}
	_ resource.ResourceWithImportState = &GroupResource{}
)

// GroupResource implements the atlassian_jira_group resource.
type GroupResource struct {
	client *atlassian.Client
}

// GroupResourceModel describes the resource data model.
type GroupResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Self types.String `tfsdk:"self"`
}

// groupCreateRequest is the JSON body for POST /rest/api/3/group.
type groupCreateRequest struct {
	Name string `json:"name"`
}

// groupCreateResponse represents the JSON response from POST /rest/api/3/group.
type groupCreateResponse struct {
	Name    string `json:"name"`
	GroupID string `json:"groupId"`
	Self    string `json:"self"`
}

// groupBulkItem represents a single group in the group bulk API response.
type groupBulkItem struct {
	GroupID string `json:"groupId"`
	Name    string `json:"name"`
}

// NewGroupResource returns a new resource factory function.
func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

func (r *GroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_group"
}

func (r *GroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the group (UUID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the group. Cannot be changed after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"self": schema.StringAttribute{
				Description: "The URL of the group.",
				Computed:    true,
			},
		},
	}
}

func (r *GroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := groupCreateRequest{
		Name: plan.Name.ValueString(),
	}

	var createResp groupCreateResponse
	if err := r.client.Post(ctx, "/rest/api/3/group", createReq, &createResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Group",
			"An error occurred while calling the Jira API to create the group.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(createResp.GroupID)
	plan.Name = types.StringValue(createResp.Name)
	plan.Self = types.StringValue(createResp.Self)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, err := atlassian.Paginate[groupBulkItem](ctx, r.client, "/rest/api/3/group/bulk?groupId="+url.QueryEscape(state.ID.ValueString()))
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Group",
			"An error occurred while calling the Jira API to read the group.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	if len(items) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(items[0].GroupID)
	state.Name = types.StringValue(items[0].Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GroupResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Unable to Update Group",
		"Groups cannot be updated. The group name is immutable; Terraform should destroy and recreate the resource.",
	)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "/rest/api/3/group?groupId="+url.QueryEscape(state.ID.ValueString()), nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			// Already deleted, nothing to do.
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Group",
			"An error occurred while calling the Jira API to delete the group.\n\n"+
				"Error: "+err.Error(),
		)
	}
}

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
