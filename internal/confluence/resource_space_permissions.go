package confluence

import (
	"context"
	"encoding/json"
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
	_ resource.Resource                = &SpacePermissionsResource{}
	_ resource.ResourceWithImportState = &SpacePermissionsResource{}
)

// SpacePermissionsResource implements the atlassian_confluence_space_permissions
// resource. It binds a single group to a single space with a customizable set of
// granular operation grants, managed through the Confluence v1 REST API.
type SpacePermissionsResource struct {
	client *atlassian.Client
}

// SpacePermissionsResourceModel describes the resource data model.
type SpacePermissionsResourceModel struct {
	ID          types.String                `tfsdk:"id"`
	SpaceKey    types.String                `tfsdk:"space_key"`
	GroupID     types.String                `tfsdk:"group_id"`
	Permissions []SpacePermissionGrantModel `tfsdk:"permission"`
}

// SpacePermissionGrantModel describes a single granular permission grant.
type SpacePermissionGrantModel struct {
	Operation types.String `tfsdk:"operation"`
	Target    types.String `tfsdk:"target"`
	ID        types.String `tfsdk:"id"`
}

// --- API types ---

// spacePermissionRequest is the body for POST {apiV1Base}/space/{key}/permission.
type spacePermissionRequest struct {
	Subject   spacePermissionSubject   `json:"subject"`
	Operation spacePermissionOperation `json:"operation"`
}

type spacePermissionSubject struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}

type spacePermissionOperation struct {
	Key    string `json:"key"`
	Target string `json:"target"`
}

// spacePermissionResponse is the body returned when creating a grant.
type spacePermissionResponse struct {
	ID json.Number `json:"id"`
}

// spaceWithPermissionsResponse is the GET {apiV1Base}/space/{key}?expand=permissions body.
type spaceWithPermissionsResponse struct {
	Key         string                `json:"key"`
	Permissions []spaceReadPermission `json:"permissions"`
}

type spaceReadPermission struct {
	ID        json.Number `json:"id"`
	Operation struct {
		Operation  string `json:"operation"`
		TargetType string `json:"targetType"`
	} `json:"operation"`
	Subjects struct {
		Group struct {
			Results []spaceReadSubject `json:"results"`
		} `json:"group"`
	} `json:"subjects"`
}

type spaceReadSubject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NewSpacePermissionsResource returns a new resource factory function.
func NewSpacePermissionsResource() resource.Resource {
	return &SpacePermissionsResource{}
}

func (r *SpacePermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_confluence_space_permissions"
}

func (r *SpacePermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the granular Confluence space permissions granted to a single group for a single space (Confluence v1 REST API).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier in the form \"{space_key}/{group_id}\".",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_key": schema.StringAttribute{
				Description: "The key of the Confluence space.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_id": schema.StringAttribute{
				Description: "The Atlassian group ID that the permissions are granted to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"permission": schema.ListNestedBlock{
				Description: "A granular permission grant for the group on the space.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"operation": schema.StringAttribute{
							Description: "The operation key, e.g. read, create, delete, export, administer, restrict_content, archive.",
							Required:    true,
						},
						"target": schema.StringAttribute{
							Description: "The operation target, e.g. page, blogpost, comment, attachment, space.",
							Required:    true,
						},
						"id": schema.StringAttribute{
							Description: "The ID of the permission grant returned by Confluence.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (r *SpacePermissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, errMsg, ok := configureClient(req.ProviderData)
	if !ok {
		if errMsg != "" {
			resp.Diagnostics.AddError("Unexpected Resource Configure Type", errMsg)
		}
		return
	}
	r.client = client
}

func (r *SpacePermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SpacePermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spaceKey := plan.SpaceKey.ValueString()
	groupID := plan.GroupID.ValueString()

	grants := make([]SpacePermissionGrantModel, 0, len(plan.Permissions))
	for _, g := range plan.Permissions {
		id, err := r.createGrant(ctx, spaceKey, groupID, g)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Space Permission",
				"An error occurred while granting a Confluence space permission.\n\n"+
					"Error: "+err.Error(),
			)
			return
		}
		g.ID = types.StringValue(id)
		grants = append(grants, g)
	}

	plan.Permissions = grants
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", spaceKey, groupID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SpacePermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SpacePermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spaceKey := state.SpaceKey.ValueString()
	groupID := state.GroupID.ValueString()

	var apiResp spaceWithPermissionsResponse
	err := r.client.Get(ctx, fmt.Sprintf("%s/space/%s?expand=permissions", apiV1Base, spaceKey), &apiResp)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Space Permissions",
			"An error occurred while reading Confluence space permissions.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	mapSpacePermissionsAPIToState(&state, &apiResp, groupID, state.Permissions)

	// If none of the configured grants for this group remain, the binding is gone.
	if len(state.Permissions) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SpacePermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SpacePermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state SpacePermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spaceKey := state.SpaceKey.ValueString()
	groupID := state.GroupID.ValueString()

	// Index current grants by operation+target so we can diff against the plan.
	current := make(map[string]SpacePermissionGrantModel, len(state.Permissions))
	for _, g := range state.Permissions {
		current[grantKey(g)] = g
	}

	desired := make(map[string]struct{}, len(plan.Permissions))
	result := make([]SpacePermissionGrantModel, 0, len(plan.Permissions))

	// Add new grants and keep unchanged ones.
	for _, g := range plan.Permissions {
		key := grantKey(g)
		desired[key] = struct{}{}
		if existing, ok := current[key]; ok {
			g.ID = existing.ID
			result = append(result, g)
			continue
		}
		id, err := r.createGrant(ctx, spaceKey, groupID, g)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Add Space Permission",
				"An error occurred while granting a Confluence space permission.\n\n"+
					"Error: "+err.Error(),
			)
			return
		}
		g.ID = types.StringValue(id)
		result = append(result, g)
	}

	// Remove grants that are no longer desired.
	for key, g := range current {
		if _, ok := desired[key]; ok {
			continue
		}
		if err := r.deleteGrant(ctx, spaceKey, g.ID.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Remove Space Permission",
				"An error occurred while revoking a Confluence space permission.\n\n"+
					"Error: "+err.Error(),
			)
			return
		}
	}

	plan.Permissions = result
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", spaceKey, groupID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SpacePermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SpacePermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spaceKey := state.SpaceKey.ValueString()
	for _, g := range state.Permissions {
		if err := r.deleteGrant(ctx, spaceKey, g.ID.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Delete Space Permission",
				"An error occurred while revoking a Confluence space permission.\n\n"+
					"Error: "+err.Error(),
			)
			return
		}
	}
}

func (r *SpacePermissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := splitCompositeID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("space_key"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// createGrant POSTs a single granular permission and returns its ID.
func (r *SpacePermissionsResource) createGrant(ctx context.Context, spaceKey, groupID string, g SpacePermissionGrantModel) (string, error) {
	body := spacePermissionRequest{
		Subject: spacePermissionSubject{
			Type:       "group",
			Identifier: groupID,
		},
		Operation: spacePermissionOperation{
			Key:    g.Operation.ValueString(),
			Target: g.Target.ValueString(),
		},
	}

	var apiResp spacePermissionResponse
	if err := r.client.Post(ctx, fmt.Sprintf("%s/space/%s/permission", apiV1Base, spaceKey), body, &apiResp); err != nil {
		return "", err
	}
	return apiResp.ID.String(), nil
}

// deleteGrant DELETEs a single granular permission by ID, ignoring 404s.
func (r *SpacePermissionsResource) deleteGrant(ctx context.Context, spaceKey, id string) error {
	if id == "" {
		return nil
	}
	err := r.client.Delete(ctx, fmt.Sprintf("%s/space/%s/permission/%s", apiV1Base, spaceKey, id), nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return nil
		}
		return err
	}
	return nil
}

// grantKey returns the stable identity of a grant for diffing (operation+target).
func grantKey(g SpacePermissionGrantModel) string {
	return g.Operation.ValueString() + "\x00" + g.Target.ValueString()
}

// mapSpacePermissionsAPIToState maps the space?expand=permissions response into the
// state model, keeping only grants for groupID and reordering them to match the prior
// state order to avoid spurious diffs (mirrors mapPermissionSchemeAPIToState).
func mapSpacePermissionsAPIToState(state *SpacePermissionsResourceModel, apiResp *spaceWithPermissionsResponse, groupID string, oldPermissions []SpacePermissionGrantModel) {
	// Preserve prior IDs keyed by operation+target; the expand=permissions payload
	// does not always echo the granular grant ID.
	priorIDs := make(map[string]string, len(oldPermissions))
	for _, g := range oldPermissions {
		priorIDs[grantKey(g)] = g.ID.ValueString()
	}

	grants := make([]SpacePermissionGrantModel, 0, len(apiResp.Permissions))
	for _, p := range apiResp.Permissions {
		matched := false
		for _, sub := range p.Subjects.Group.Results {
			if sub.ID == groupID || sub.Name == groupID {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}

		grant := SpacePermissionGrantModel{
			Operation: types.StringValue(p.Operation.Operation),
			Target:    types.StringValue(p.Operation.TargetType),
		}

		id := p.ID.String()
		if id == "" || id == "0" {
			id = priorIDs[grantKey(grant)]
		}
		grant.ID = types.StringValue(id)

		grants = append(grants, grant)
	}

	// Reorder API results to match the existing state order to avoid spurious diffs.
	if len(oldPermissions) > 0 {
		ordered := make([]SpacePermissionGrantModel, 0, len(grants))
		used := make([]bool, len(grants))
		for _, planned := range oldPermissions {
			for j, api := range grants {
				if !used[j] &&
					api.Operation.ValueString() == planned.Operation.ValueString() &&
					api.Target.ValueString() == planned.Target.ValueString() {
					ordered = append(ordered, api)
					used[j] = true
					break
				}
			}
		}
		for j, api := range grants {
			if !used[j] {
				ordered = append(ordered, api)
			}
		}
		grants = ordered
	}

	state.Permissions = grants
}
