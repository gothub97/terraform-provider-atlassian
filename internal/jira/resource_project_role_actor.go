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
	_ resource.Resource                = &ProjectRoleActorResource{}
	_ resource.ResourceWithImportState = &ProjectRoleActorResource{}
)

// ProjectRoleActorResource implements the atlassian_jira_project_role_actor resource.
type ProjectRoleActorResource struct {
	client *atlassian.Client
}

// ProjectRoleActorResourceModel describes the resource data model.
type ProjectRoleActorResourceModel struct {
	ID         types.String `tfsdk:"id"`
	ProjectKey types.String `tfsdk:"project_key"`
	RoleID     types.String `tfsdk:"role_id"`
	ActorType  types.String `tfsdk:"actor_type"`
	ActorValue types.String `tfsdk:"actor_value"`
}

// projectRoleWithActorsResponse represents the JSON response from GET /rest/api/3/project/{projectKey}/role/{roleId}.
type projectRoleWithActorsResponse struct {
	ID     int              `json:"id"`
	Name   string           `json:"name"`
	Actors []roleActorEntry `json:"actors"`
}

// roleActorEntry represents a single actor in the project role.
type roleActorEntry struct {
	ID         int         `json:"id"`
	Type       string      `json:"type"`
	ActorUser  *actorUser  `json:"actorUser"`
	ActorGroup *actorGroup `json:"actorGroup"`
}

// actorUser represents the user details in an actor entry.
type actorUser struct {
	AccountID string `json:"accountId"`
}

// actorGroup represents the group details in an actor entry.
type actorGroup struct {
	GroupID string `json:"groupId"`
	Name    string `json:"name"`
}

// addActorRequest is the JSON body for POST /rest/api/3/project/{projectKey}/role/{roleId}.
type addActorRequest struct {
	User    []string `json:"user,omitempty"`
	GroupID []string `json:"groupId,omitempty"`
}

// NewProjectRoleActorResource returns a new resource factory function.
func NewProjectRoleActorResource() resource.Resource {
	return &ProjectRoleActorResource{}
}

func (r *ProjectRoleActorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_project_role_actor"
}

func (r *ProjectRoleActorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira project role actor (user or group membership).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID: projectKey/roleId/actorType/actorValue.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_key": schema.StringAttribute{
				Description: "The project key.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_id": schema.StringAttribute{
				Description: "The numeric ID of the project role.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"actor_type": schema.StringAttribute{
				Description: "The type of actor: \"user\" or \"group\".",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"actor_value": schema.StringAttribute{
				Description: "The account ID (for user) or group ID (for group).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *ProjectRoleActorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectRoleActorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectRoleActorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := plan.ProjectKey.ValueString()
	roleID := plan.RoleID.ValueString()
	actorType := plan.ActorType.ValueString()
	actorValue := plan.ActorValue.ValueString()

	apiPath := fmt.Sprintf("/rest/api/3/project/%s/role/%s", projectKey, roleID)

	var addReq addActorRequest
	if actorType == "user" {
		addReq.User = []string{actorValue}
	} else {
		addReq.GroupID = []string{actorValue}
	}

	if err := r.client.Post(ctx, apiPath, addReq, nil); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Add Project Role Actor",
			"An error occurred while calling the Jira API to add the project role actor.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s/%s", projectKey, roleID, actorType, actorValue))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectRoleActorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectRoleActorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := state.ProjectKey.ValueString()
	roleID := state.RoleID.ValueString()
	actorType := state.ActorType.ValueString()
	actorValue := state.ActorValue.ValueString()

	apiPath := fmt.Sprintf("/rest/api/3/project/%s/role/%s", projectKey, roleID)

	var apiResp projectRoleWithActorsResponse
	err := r.client.Get(ctx, apiPath, &apiResp)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Project Role Actors",
			"An error occurred while calling the Jira API to read the project role actors.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Search for the matching actor
	found := false
	for _, actor := range apiResp.Actors {
		if actorType == "user" && actor.Type == "atlassian-user-role-actor" && actor.ActorUser != nil && actor.ActorUser.AccountID == actorValue {
			found = true
			break
		}
		if actorType == "group" && actor.Type == "atlassian-group-role-actor" && actor.ActorGroup != nil && actor.ActorGroup.GroupID == actorValue {
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

func (r *ProjectRoleActorResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes require replace, so Update should never be called.
	resp.Diagnostics.AddError(
		"Unexpected Update",
		"All attributes of this resource require replacement. Update should not be called.",
	)
}

func (r *ProjectRoleActorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectRoleActorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := state.ProjectKey.ValueString()
	roleID := state.RoleID.ValueString()
	actorType := state.ActorType.ValueString()
	actorValue := state.ActorValue.ValueString()

	var apiPath string
	if actorType == "user" {
		apiPath = fmt.Sprintf("/rest/api/3/project/%s/role/%s?user=%s", projectKey, roleID, url.QueryEscape(actorValue))
	} else {
		apiPath = fmt.Sprintf("/rest/api/3/project/%s/role/%s?groupId=%s", projectKey, roleID, url.QueryEscape(actorValue))
	}

	err := r.client.Delete(ctx, apiPath, nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			// Already deleted, nothing to do.
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Project Role Actor",
			"An error occurred while calling the Jira API to delete the project role actor.\n\n"+
				"Error: "+err.Error(),
		)
	}
}

func (r *ProjectRoleActorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 4)
	if len(parts) != 4 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: projectKey/roleId/actorType/actorValue, got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_key"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("actor_type"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("actor_value"), parts[3])...)
}
