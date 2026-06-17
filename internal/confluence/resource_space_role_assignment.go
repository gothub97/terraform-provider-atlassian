package confluence

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &SpaceRoleAssignmentResource{}
	_ resource.ResourceWithImportState = &SpaceRoleAssignmentResource{}
)

// SpaceRoleAssignmentResource implements the atlassian_confluence_space_role_assignment
// resource, managing v2 space role assignments for a single principal.
type SpaceRoleAssignmentResource struct {
	client *atlassian.Client
}

// SpaceRoleAssignmentResourceModel describes the resource data model.
type SpaceRoleAssignmentResourceModel struct {
	ID            types.String `tfsdk:"id"`
	SpaceID       types.String `tfsdk:"space_id"`
	PrincipalType types.String `tfsdk:"principal_type"`
	PrincipalID   types.String `tfsdk:"principal_id"`
	RoleID        types.String `tfsdk:"role_id"`
}

// --- API types ---

// roleAssignmentPrincipal identifies the subject of a role assignment.
type roleAssignmentPrincipal struct {
	PrincipalType string `json:"principalType"`
	PrincipalID   string `json:"principalId"`
}

// roleAssignmentRequest is one element of the POST role-assignments body. RoleID is
// omitted when clearing an assignment (delete semantics).
type roleAssignmentRequest struct {
	Principal roleAssignmentPrincipal `json:"principal"`
	RoleID    string                  `json:"roleId,omitempty"`
}

// roleAssignmentsListResponse is the GET role-assignments response envelope.
type roleAssignmentsListResponse struct {
	Results []roleAssignmentResult `json:"results"`
	Links   struct {
		Next string `json:"next"`
	} `json:"_links"`
}

type roleAssignmentResult struct {
	Principal roleAssignmentPrincipal `json:"principal"`
	RoleID    string                  `json:"roleId"`
}

// NewSpaceRoleAssignmentResource returns a new resource factory function.
func NewSpaceRoleAssignmentResource() resource.Resource {
	return &SpaceRoleAssignmentResource{}
}

func (r *SpaceRoleAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_confluence_space_role_assignment"
}

func (r *SpaceRoleAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Confluence space role assignment for a single principal (Confluence v2 REST API).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier in the form \"{space_id}/{principal_id}\".",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "The ID of the Confluence space.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"principal_type": schema.StringAttribute{
				Description: "The type of the principal: USER, GROUP, or ACCESS_CLASS. Defaults to GROUP.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("GROUP"),
			},
			"principal_id": schema.StringAttribute{
				Description: "The ID of the principal (e.g. the group ID).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_id": schema.StringAttribute{
				Description: "The ID of the space role to assign to the principal.",
				Required:    true,
			},
		},
	}
}

func (r *SpaceRoleAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, errMsg, ok := configureClient(req.ProviderData)
	if !ok {
		if errMsg != "" {
			resp.Diagnostics.AddError("Unexpected Resource Configure Type", errMsg)
		}
		return
	}
	r.client = client
}

func (r *SpaceRoleAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SpaceRoleAssignmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.assignRole(ctx, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Space Role Assignment",
			"An error occurred while assigning a Confluence space role.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.SpaceID.ValueString(), plan.PrincipalID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SpaceRoleAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SpaceRoleAssignmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	match, found, err := r.findAssignment(ctx, state.SpaceID.ValueString(), state.PrincipalID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Space Role Assignment",
			"An error occurred while reading Confluence space role assignments.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	state.RoleID = types.StringValue(match.RoleID)
	if match.Principal.PrincipalType != "" {
		state.PrincipalType = types.StringValue(match.Principal.PrincipalType)
	}
	state.ID = types.StringValue(fmt.Sprintf("%s/%s", state.SpaceID.ValueString(), state.PrincipalID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SpaceRoleAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SpaceRoleAssignmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.assignRole(ctx, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Space Role Assignment",
			"An error occurred while updating a Confluence space role assignment.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.SpaceID.ValueString(), plan.PrincipalID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SpaceRoleAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SpaceRoleAssignmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// v2 semantics: re-POST the principal WITHOUT a roleId to clear the assignment.
	body := []roleAssignmentRequest{
		{
			Principal: roleAssignmentPrincipal{
				PrincipalType: state.PrincipalType.ValueString(),
				PrincipalID:   state.PrincipalID.ValueString(),
			},
		},
	}

	err := r.client.Post(ctx, fmt.Sprintf("%s/spaces/%s/role-assignments", apiV2Base, state.SpaceID.ValueString()), body, nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Space Role Assignment",
			"An error occurred while clearing a Confluence space role assignment.\n\n"+
				"Error: "+err.Error(),
		)
	}
}

func (r *SpaceRoleAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := splitCompositeID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("space_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("principal_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// assignRole POSTs the principal+role to the v2 role-assignments endpoint.
func (r *SpaceRoleAssignmentResource) assignRole(ctx context.Context, plan *SpaceRoleAssignmentResourceModel) error {
	body := []roleAssignmentRequest{
		{
			Principal: roleAssignmentPrincipal{
				PrincipalType: plan.PrincipalType.ValueString(),
				PrincipalID:   plan.PrincipalID.ValueString(),
			},
			RoleID: plan.RoleID.ValueString(),
		},
	}
	return r.client.Post(ctx, fmt.Sprintf("%s/spaces/%s/role-assignments", apiV2Base, plan.SpaceID.ValueString()), body, nil)
}

// findAssignment paginates the v2 role-assignments endpoint and returns the
// assignment matching principalID, following the cursor-based _links.next.
func (r *SpaceRoleAssignmentResource) findAssignment(ctx context.Context, spaceID, principalID string) (roleAssignmentResult, bool, error) {
	path := fmt.Sprintf("%s/spaces/%s/role-assignments?limit=250", apiV2Base, spaceID)

	for path != "" {
		var page roleAssignmentsListResponse
		if err := r.client.Get(ctx, path, &page); err != nil {
			return roleAssignmentResult{}, false, err
		}
		for _, a := range page.Results {
			if a.Principal.PrincipalID == principalID {
				return a, true, nil
			}
		}
		path = page.Links.Next
	}

	return roleAssignmentResult{}, false, nil
}
