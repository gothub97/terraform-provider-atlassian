package jira

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &GroupMembershipResource{}
	_ resource.ResourceWithImportState = &GroupMembershipResource{}
)

// GroupMembershipResource implements the atlassian_jira_group_membership resource.
type GroupMembershipResource struct {
	client *atlassian.Client
}

// GroupMembershipResourceModel describes the resource data model.
type GroupMembershipResourceModel struct {
	ID        types.String `tfsdk:"id"`
	GroupID   types.String `tfsdk:"group_id"`
	AccountID types.String `tfsdk:"account_id"`
}

// groupMemberAddRequest is the JSON body for POST /rest/api/3/group/user.
type groupMemberAddRequest struct {
	AccountID string `json:"accountId"`
}

// groupMemberItem represents a single member in the group member API response.
type groupMemberItem struct {
	AccountID string `json:"accountId"`
}

// NewGroupMembershipResource returns a new resource factory function.
func NewGroupMembershipResource() resource.Resource {
	return &GroupMembershipResource{}
}

func (r *GroupMembershipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_group_membership"
}

func (r *GroupMembershipResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages membership of a user in a Jira group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The composite ID of the group membership (groupId/accountId).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_id": schema.StringAttribute{
				Description: "The ID of the group (UUID).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"account_id": schema.StringAttribute{
				Description: "The account ID of the user.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *GroupMembershipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GroupMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupMembershipResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addReq := groupMemberAddRequest{
		AccountID: plan.AccountID.ValueString(),
	}

	apiPath := "/rest/api/3/group/user?groupId=" + url.QueryEscape(plan.GroupID.ValueString())
	if err := r.client.Post(ctx, apiPath, addReq, nil); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Add User to Group",
			"An error occurred while calling the Jira API to add the user to the group.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(plan.GroupID.ValueString() + "/" + plan.AccountID.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GroupMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupMembershipResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPath := "/rest/api/3/group/member?groupId=" + url.QueryEscape(state.GroupID.ValueString())
	members, err := atlassian.Paginate[groupMemberItem](ctx, r.client, apiPath)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Group Membership",
			"An error occurred while calling the Jira API to read group members.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	found := false
	for _, member := range members {
		if member.AccountID == state.AccountID.ValueString() {
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GroupMembershipResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Unable to Update Group Membership",
		"Group memberships cannot be updated. Both group_id and account_id are immutable; Terraform should destroy and recreate the resource.",
	)
}

func (r *GroupMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GroupMembershipResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPath := "/rest/api/3/group/user?groupId=" + url.QueryEscape(state.GroupID.ValueString()) +
		"&accountId=" + url.QueryEscape(state.AccountID.ValueString())
	err := r.client.Delete(ctx, apiPath, nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			// Already removed, nothing to do.
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Remove User from Group",
			"An error occurred while calling the Jira API to remove the user from the group.\n\n"+
				"Error: "+err.Error(),
		)
	}
}

func (r *GroupMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in the format 'groupId/accountId', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("account_id"), parts[1])...)
}
