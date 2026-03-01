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
	_ resource.Resource                = &IssueTypeScreenSchemeResource{}
	_ resource.ResourceWithImportState = &IssueTypeScreenSchemeResource{}
)

// IssueTypeScreenSchemeResource implements the atlassian_jira_issue_type_screen_scheme resource.
type IssueTypeScreenSchemeResource struct {
	client *atlassian.Client
}

// IssueTypeScreenSchemeResourceModel describes the resource data model.
type IssueTypeScreenSchemeResourceModel struct {
	ID          types.String                        `tfsdk:"id"`
	Name        types.String                        `tfsdk:"name"`
	Description types.String                        `tfsdk:"description"`
	Mappings    []IssueTypeScreenSchemeMappingModel `tfsdk:"mapping"`
}

// IssueTypeScreenSchemeMappingModel describes a single mapping in the issue type screen scheme.
type IssueTypeScreenSchemeMappingModel struct {
	IssueTypeID    types.String `tfsdk:"issue_type_id"`
	ScreenSchemeID types.String `tfsdk:"screen_scheme_id"`
}

// issueTypeScreenSchemeCreateRequest is the JSON body for POST /rest/api/3/issuetypescreenscheme.
type issueTypeScreenSchemeCreateRequest struct {
	Name              string                           `json:"name"`
	Description       string                           `json:"description,omitempty"`
	IssueTypeMappings []issueTypeScreenSchemeMappingAPI `json:"issueTypeMappings"`
}

// issueTypeScreenSchemeMappingAPI represents a mapping in API requests and responses.
type issueTypeScreenSchemeMappingAPI struct {
	IssueTypeID    string `json:"issueTypeId"`
	ScreenSchemeID string `json:"screenSchemeId"`
}

// issueTypeScreenSchemeCreateResponse represents the JSON response from POST /rest/api/3/issuetypescreenscheme.
type issueTypeScreenSchemeCreateResponse struct {
	ID string `json:"id"`
}

// issueTypeScreenSchemeAPIResponse represents the scheme metadata from the paginated API.
type issueTypeScreenSchemeAPIResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// issueTypeScreenSchemeUpdateRequest is the JSON body for PUT /rest/api/3/issuetypescreenscheme/{id}.
type issueTypeScreenSchemeUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// issueTypeScreenSchemeMappingResponse represents a mapping from the paginated mappings API.
type issueTypeScreenSchemeMappingResponse struct {
	IssueTypeID              string `json:"issueTypeId"`
	ScreenSchemeID           string `json:"screenSchemeId"`
	IssueTypeScreenSchemeID  string `json:"issueTypeScreenSchemeId"`
}

// issueTypeScreenSchemeAddMappingsRequest is the JSON body for PUT /rest/api/3/issuetypescreenscheme/{id}/mapping.
type issueTypeScreenSchemeAddMappingsRequest struct {
	IssueTypeMappings []issueTypeScreenSchemeMappingAPI `json:"issueTypeMappings"`
}

// issueTypeScreenSchemeRemoveMappingsRequest is the JSON body for POST /rest/api/3/issuetypescreenscheme/{id}/mapping/remove.
type issueTypeScreenSchemeRemoveMappingsRequest struct {
	IssueTypeIDs []string `json:"issueTypeIds"`
}

// issueTypeScreenSchemeUpdateDefaultRequest is the JSON body for PUT /rest/api/3/issuetypescreenscheme/{id}/mapping/default.
type issueTypeScreenSchemeUpdateDefaultRequest struct {
	ScreenSchemeID string `json:"screenSchemeId"`
}

// NewIssueTypeScreenSchemeResource returns a new resource factory function.
func NewIssueTypeScreenSchemeResource() resource.Resource {
	return &IssueTypeScreenSchemeResource{}
}

func (r *IssueTypeScreenSchemeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_issue_type_screen_scheme"
}

func (r *IssueTypeScreenSchemeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira issue type screen scheme.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The numeric ID of the issue type screen scheme.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the issue type screen scheme.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the issue type screen scheme.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"mapping": schema.ListNestedBlock{
				Description: "A mapping of an issue type to a screen scheme.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"issue_type_id": schema.StringAttribute{
							Description: "The ID of the issue type, or 'default' for the default mapping.",
							Required:    true,
						},
						"screen_scheme_id": schema.StringAttribute{
							Description: "The ID of the screen scheme.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *IssueTypeScreenSchemeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IssueTypeScreenSchemeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IssueTypeScreenSchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the mappings
	mappings := make([]issueTypeScreenSchemeMappingAPI, len(plan.Mappings))
	for i, m := range plan.Mappings {
		mappings[i] = issueTypeScreenSchemeMappingAPI{
			IssueTypeID:    m.IssueTypeID.ValueString(),
			ScreenSchemeID: m.ScreenSchemeID.ValueString(),
		}
	}

	createReq := issueTypeScreenSchemeCreateRequest{
		Name:              plan.Name.ValueString(),
		IssueTypeMappings: mappings,
	}
	if !plan.Description.IsNull() {
		createReq.Description = plan.Description.ValueString()
	}

	var createResp issueTypeScreenSchemeCreateResponse
	if err := r.client.Post(ctx, "/rest/api/3/issuetypescreenscheme", createReq, &createResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Issue Type Screen Scheme",
			"An error occurred while calling the Jira API to create the issue type screen scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(createResp.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IssueTypeScreenSchemeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IssueTypeScreenSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemeID := state.ID.ValueString()

	// Read scheme metadata
	schemes, err := atlassian.Paginate[issueTypeScreenSchemeAPIResponse](ctx, r.client, fmt.Sprintf("/rest/api/3/issuetypescreenscheme?id=%s", schemeID))
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Issue Type Screen Scheme",
			"An error occurred while calling the Jira API to read the issue type screen scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	if len(schemes) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	scheme := schemes[0]
	state.ID = types.StringValue(scheme.ID)
	state.Name = types.StringValue(scheme.Name)
	if scheme.Description == "" {
		state.Description = types.StringNull()
	} else {
		state.Description = types.StringValue(scheme.Description)
	}

	// Read mappings
	mappingsResp, err := atlassian.Paginate[issueTypeScreenSchemeMappingResponse](ctx, r.client, fmt.Sprintf("/rest/api/3/issuetypescreenscheme/mapping?issueTypeScreenSchemeId=%s", schemeID))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Issue Type Screen Scheme Mappings",
			"An error occurred while calling the Jira API to read the issue type screen scheme mappings.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state.Mappings = make([]IssueTypeScreenSchemeMappingModel, len(mappingsResp))
	for i, m := range mappingsResp {
		state.Mappings[i] = IssueTypeScreenSchemeMappingModel{
			IssueTypeID:    types.StringValue(m.IssueTypeID),
			ScreenSchemeID: types.StringValue(m.ScreenSchemeID),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *IssueTypeScreenSchemeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan IssueTypeScreenSchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state IssueTypeScreenSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemeID := state.ID.ValueString()

	// Step 1: Update scheme metadata (name, description)
	updateReq := issueTypeScreenSchemeUpdateRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() {
		updateReq.Description = plan.Description.ValueString()
	}

	if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/issuetypescreenscheme/%s", schemeID), updateReq, nil); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Issue Type Screen Scheme",
			"An error occurred while calling the Jira API to update the issue type screen scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Step 2: Sync mappings
	// Read current mappings from API
	currentMappings, err := atlassian.Paginate[issueTypeScreenSchemeMappingResponse](ctx, r.client, fmt.Sprintf("/rest/api/3/issuetypescreenscheme/mapping?issueTypeScreenSchemeId=%s", schemeID))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Current Mappings",
			"An error occurred while reading the current mappings for update.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Build desired mappings map
	desiredMap := make(map[string]string)
	for _, m := range plan.Mappings {
		desiredMap[m.IssueTypeID.ValueString()] = m.ScreenSchemeID.ValueString()
	}

	// Build current mappings map
	currentMap := make(map[string]string)
	for _, m := range currentMappings {
		currentMap[m.IssueTypeID] = m.ScreenSchemeID
	}

	// Handle default mapping update
	if desiredDefault, ok := desiredMap["default"]; ok {
		if currentDefault, exists := currentMap["default"]; !exists || currentDefault != desiredDefault {
			defaultReq := issueTypeScreenSchemeUpdateDefaultRequest{
				ScreenSchemeID: desiredDefault,
			}
			if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/issuetypescreenscheme/%s/mapping/default", schemeID), defaultReq, nil); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Update Default Mapping",
					"An error occurred while updating the default mapping.\n\n"+
						"Error: "+err.Error(),
				)
				return
			}
		}
	}

	// Find non-default mappings to remove (in current but not in desired, or changed)
	var removeIDs []string
	for issueTypeID := range currentMap {
		if issueTypeID == "default" {
			continue
		}
		if _, exists := desiredMap[issueTypeID]; !exists {
			removeIDs = append(removeIDs, issueTypeID)
		}
	}

	// Also remove mappings that need to be updated (changed screen scheme)
	var addMappings []issueTypeScreenSchemeMappingAPI
	for issueTypeID, desiredSchemeID := range desiredMap {
		if issueTypeID == "default" {
			continue
		}
		currentSchemeID, exists := currentMap[issueTypeID]
		if !exists {
			// New mapping
			addMappings = append(addMappings, issueTypeScreenSchemeMappingAPI{
				IssueTypeID:    issueTypeID,
				ScreenSchemeID: desiredSchemeID,
			})
		} else if currentSchemeID != desiredSchemeID {
			// Changed mapping — need to remove and re-add
			removeIDs = append(removeIDs, issueTypeID)
			addMappings = append(addMappings, issueTypeScreenSchemeMappingAPI{
				IssueTypeID:    issueTypeID,
				ScreenSchemeID: desiredSchemeID,
			})
		}
	}

	// Remove mappings
	if len(removeIDs) > 0 {
		removeReq := issueTypeScreenSchemeRemoveMappingsRequest{
			IssueTypeIDs: removeIDs,
		}
		if err := r.client.Post(ctx, fmt.Sprintf("/rest/api/3/issuetypescreenscheme/%s/mapping/remove", schemeID), removeReq, nil); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Remove Mappings",
				"An error occurred while removing mappings from the issue type screen scheme.\n\n"+
					"Error: "+err.Error(),
			)
			return
		}
	}

	// Add mappings
	if len(addMappings) > 0 {
		addReq := issueTypeScreenSchemeAddMappingsRequest{
			IssueTypeMappings: addMappings,
		}
		if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/issuetypescreenscheme/%s/mapping", schemeID), addReq, nil); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Add Mappings",
				"An error occurred while adding mappings to the issue type screen scheme.\n\n"+
					"Error: "+err.Error(),
			)
			return
		}
	}

	plan.ID = state.ID

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IssueTypeScreenSchemeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state IssueTypeScreenSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/rest/api/3/issuetypescreenscheme/%s", state.ID.ValueString()), nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			// Already deleted, nothing to do.
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Issue Type Screen Scheme",
			"An error occurred while calling the Jira API to delete the issue type screen scheme.\n\n"+
				"Error: "+err.Error(),
		)
	}
}

func (r *IssueTypeScreenSchemeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
