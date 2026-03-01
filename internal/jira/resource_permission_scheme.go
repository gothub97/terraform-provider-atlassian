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
	_ resource.Resource                = &PermissionSchemeResource{}
	_ resource.ResourceWithImportState = &PermissionSchemeResource{}
)

// PermissionSchemeResource implements the atlassian_jira_permission_scheme resource.
type PermissionSchemeResource struct {
	client *atlassian.Client
}

// PermissionSchemeResourceModel describes the resource data model.
type PermissionSchemeResourceModel struct {
	ID          types.String                `tfsdk:"id"`
	Name        types.String                `tfsdk:"name"`
	Description types.String                `tfsdk:"description"`
	Self        types.String                `tfsdk:"self"`
	Permissions []PermissionSchemeGrantModel `tfsdk:"permission"`
}

// PermissionSchemeGrantModel describes a single permission grant.
type PermissionSchemeGrantModel struct {
	ID          types.String `tfsdk:"id"`
	Permission  types.String `tfsdk:"permission"`
	HolderType  types.String `tfsdk:"holder_type"`
	HolderValue types.String `tfsdk:"holder_value"`
}

// --- API types ---

// permissionSchemeCreateRequest is the body for POST/PUT /rest/api/3/permissionscheme.
type permissionSchemeCreateRequest struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description,omitempty"`
	Permissions []permissionGrantRequest `json:"permissions,omitempty"`
}

type permissionGrantRequest struct {
	Permission string                  `json:"permission"`
	Holder     permissionHolderRequest `json:"holder"`
}

type permissionHolderRequest struct {
	Type  string `json:"type"`
	Value string `json:"value,omitempty"`
}

// permissionSchemeAPIResponse represents the response from the permission scheme API.
type permissionSchemeAPIResponse struct {
	ID          int                       `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Self        string                    `json:"self"`
	Permissions []permissionGrantResponse `json:"permissions"`
}

type permissionGrantResponse struct {
	ID         int                      `json:"id"`
	Permission string                   `json:"permission"`
	Holder     permissionHolderResponse `json:"holder"`
}

type permissionHolderResponse struct {
	Type      string `json:"type"`
	Parameter string `json:"parameter"`
	Value     string `json:"value"`
}

// NewPermissionSchemeResource returns a new resource factory function.
func NewPermissionSchemeResource() resource.Resource {
	return &PermissionSchemeResource{}
}

func (r *PermissionSchemeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_permission_scheme"
}

func (r *PermissionSchemeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira permission scheme.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The numeric ID of the permission scheme.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the permission scheme.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the permission scheme.",
				Optional:    true,
			},
			"self": schema.StringAttribute{
				Description: "The URL of the permission scheme.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"permission": schema.ListNestedBlock{
				Description: "Permission grants in the scheme.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the permission grant.",
							Computed:    true,
						},
						"permission": schema.StringAttribute{
							Description: "The permission key, e.g. ADMINISTER_PROJECTS.",
							Required:    true,
						},
						"holder_type": schema.StringAttribute{
							Description: "The type of the permission holder, e.g. group, user, projectRole, anyone, reporter, assignee, projectLead.",
							Required:    true,
						},
						"holder_value": schema.StringAttribute{
							Description: "The value for the holder (group ID, user account ID, role ID, etc). Not needed for holder types like anyone, reporter, assignee, projectLead, currentUser.",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (r *PermissionSchemeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PermissionSchemeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PermissionSchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := permissionSchemeCreateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	// Build permissions array
	apiReq.Permissions = buildPermissionGrantRequests(plan.Permissions)

	var createResp permissionSchemeAPIResponse
	if err := r.client.Post(ctx, "/rest/api/3/permissionscheme", apiReq, &createResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Permission Scheme",
			"An error occurred while calling the Jira API to create the permission scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Jira adds default permissions on POST; PUT with exact grants to overwrite them.
	schemeID := fmt.Sprintf("%d", createResp.ID)
	var apiResp permissionSchemeAPIResponse
	if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/permissionscheme/%s", schemeID), apiReq, &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Permission Scheme After Create",
			"The permission scheme was created but Jira added default grants. "+
				"An error occurred while overwriting them with the configured grants.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	mapPermissionSchemeAPIToState(&plan, &apiResp, plan.Permissions)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PermissionSchemeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PermissionSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	oldPermissions := state.Permissions

	var apiResp permissionSchemeAPIResponse
	err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/permissionscheme/%s?expand=all", state.ID.ValueString()), &apiResp)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Permission Scheme",
			"An error occurred while calling the Jira API to read the permission scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	mapPermissionSchemeAPIToState(&state, &apiResp, oldPermissions)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PermissionSchemeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PermissionSchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PermissionSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := permissionSchemeCreateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	// Build permissions array from the plan (PUT overwrites all grants)
	apiReq.Permissions = buildPermissionGrantRequests(plan.Permissions)

	var apiResp permissionSchemeAPIResponse
	if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/permissionscheme/%s", state.ID.ValueString()), apiReq, &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Permission Scheme",
			"An error occurred while calling the Jira API to update the permission scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	mapPermissionSchemeAPIToState(&plan, &apiResp, plan.Permissions)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PermissionSchemeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PermissionSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/rest/api/3/permissionscheme/%s", state.ID.ValueString()), nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			// Already deleted, nothing to do.
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Permission Scheme",
			"An error occurred while calling the Jira API to delete the permission scheme.\n\n"+
				"Error: "+err.Error(),
		)
	}
}

func (r *PermissionSchemeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// holderTypeNeedsValue returns true if the holder type requires a value parameter.
func holderTypeNeedsValue(holderType string) bool {
	switch holderType {
	case "anyone", "reporter", "assignee", "projectLead", "currentUser":
		return false
	default:
		return true
	}
}

// buildPermissionGrantRequests converts the Terraform model grants to API request grants.
func buildPermissionGrantRequests(grants []PermissionSchemeGrantModel) []permissionGrantRequest {
	if len(grants) == 0 {
		return nil
	}

	result := make([]permissionGrantRequest, 0, len(grants))
	for _, g := range grants {
		grant := permissionGrantRequest{
			Permission: g.Permission.ValueString(),
			Holder: permissionHolderRequest{
				Type: g.HolderType.ValueString(),
			},
		}

		if !g.HolderValue.IsNull() && !g.HolderValue.IsUnknown() {
			grant.Holder.Value = g.HolderValue.ValueString()
		}

		result = append(result, grant)
	}
	return result
}

// mapPermissionSchemeAPIToState maps a permission scheme API response to the Terraform state model.
// oldPermissions is the prior state ordering used to reorder API results and avoid spurious diffs.
func mapPermissionSchemeAPIToState(state *PermissionSchemeResourceModel, apiResp *permissionSchemeAPIResponse, oldPermissions []PermissionSchemeGrantModel) {
	state.ID = types.StringValue(fmt.Sprintf("%d", apiResp.ID))
	state.Name = types.StringValue(apiResp.Name)
	state.Self = types.StringValue(apiResp.Self)

	if apiResp.Description != "" {
		state.Description = types.StringValue(apiResp.Description)
	} else {
		state.Description = types.StringNull()
	}

	// Map permission grants
	if len(apiResp.Permissions) > 0 {
		state.Permissions = make([]PermissionSchemeGrantModel, 0, len(apiResp.Permissions))
		for _, p := range apiResp.Permissions {
			grant := PermissionSchemeGrantModel{
				ID:         types.StringValue(fmt.Sprintf("%d", p.ID)),
				Permission: types.StringValue(p.Permission),
				HolderType: types.StringValue(p.Holder.Type),
			}

			// Set holder_value from response's holder.value (preferred) or holder.parameter (fallback)
			holderValue := p.Holder.Value
			if holderValue == "" {
				holderValue = p.Holder.Parameter
			}

			if holderValue != "" && holderTypeNeedsValue(p.Holder.Type) {
				grant.HolderValue = types.StringValue(holderValue)
			} else {
				grant.HolderValue = types.StringNull()
			}

			state.Permissions = append(state.Permissions, grant)
		}

		// Reorder API results to match the existing state order to avoid spurious diffs.
		if len(oldPermissions) > 0 {
			ordered := make([]PermissionSchemeGrantModel, 0, len(state.Permissions))
			used := make([]bool, len(state.Permissions))
			for _, planned := range oldPermissions {
				for j, api := range state.Permissions {
					if !used[j] &&
						api.Permission.ValueString() == planned.Permission.ValueString() &&
						api.HolderType.ValueString() == planned.HolderType.ValueString() &&
						api.HolderValue.ValueString() == planned.HolderValue.ValueString() {
						ordered = append(ordered, api)
						used[j] = true
						break
					}
				}
			}
			for j, api := range state.Permissions {
				if !used[j] {
					ordered = append(ordered, api)
				}
			}
			state.Permissions = ordered
		}
	} else {
		state.Permissions = []PermissionSchemeGrantModel{}
	}
}
