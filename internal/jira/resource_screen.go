package jira

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &ScreenResource{}
	_ resource.ResourceWithImportState = &ScreenResource{}
)

// ScreenResource implements the atlassian_jira_screen resource.
type ScreenResource struct {
	client *atlassian.Client
}

// ScreenResourceModel describes the resource data model.
type ScreenResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Description types.String   `tfsdk:"description"`
	Tabs        []ScreenTabModel `tfsdk:"tab"`
}

// ScreenTabModel describes a single tab in the screen.
type ScreenTabModel struct {
	ID     types.String   `tfsdk:"id"`
	Name   types.String   `tfsdk:"name"`
	Fields []types.String `tfsdk:"fields"`
}

// screenCreateRequest is the JSON body for POST /rest/api/3/screens.
type screenCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// screenUpdateRequest is the JSON body for PUT /rest/api/3/screens/{id}.
type screenUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// screenAPIResponse represents the JSON response from the screen API.
type screenAPIResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// screenTabAPIResponse represents a tab returned by the API.
type screenTabAPIResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// screenTabCreateRequest is the JSON body for POST /rest/api/3/screens/{id}/tabs.
type screenTabCreateRequest struct {
	Name string `json:"name"`
}

// screenTabUpdateRequest is the JSON body for PUT /rest/api/3/screens/{id}/tabs/{tabId}.
type screenTabUpdateRequest struct {
	Name string `json:"name"`
}

// screenTabFieldAPIResponse represents a field on a tab returned by the API.
type screenTabFieldAPIResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// screenTabFieldAddRequest is the JSON body for POST /rest/api/3/screens/{id}/tabs/{tabId}/fields.
type screenTabFieldAddRequest struct {
	FieldID string `json:"fieldId"`
}

// screenTabFieldMoveRequest is the JSON body for POST /rest/api/3/screens/{id}/tabs/{tabId}/fields/{fieldId}/move.
type screenTabFieldMoveRequest struct {
	Position string `json:"position,omitempty"`
	After    string `json:"after,omitempty"`
}

// NewScreenResource returns a new resource factory function.
func NewScreenResource() resource.Resource {
	return &ScreenResource{}
}

func (r *ScreenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_screen"
}

func (r *ScreenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira screen with tabs and fields.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The numeric ID of the screen.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the screen.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the screen.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"tab": schema.ListNestedBlock{
				Description: "A tab on the screen. Tabs are ordered by their position in the configuration.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The numeric ID of the tab.",
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							Description: "The name of the tab.",
							Required:    true,
						},
						"fields": schema.ListAttribute{
							Description: "Ordered list of field IDs on this tab.",
							Optional:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (r *ScreenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ScreenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ScreenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 1: Create the screen
	createReq := screenCreateRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() {
		createReq.Description = plan.Description.ValueString()
	}

	var screenResp screenAPIResponse
	if err := r.client.Post(ctx, "/rest/api/3/screens", createReq, &screenResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Screen",
			"An error occurred while calling the Jira API to create the screen.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	screenID := strconv.Itoa(screenResp.ID)
	plan.ID = types.StringValue(screenID)

	// Step 2: List the default tabs created by Jira
	var defaultTabs []screenTabAPIResponse
	if err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs", screenID), &defaultTabs); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Screen Tabs",
			"The screen was created but an error occurred while reading its default tabs.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Step 3: Configure tabs
	if len(plan.Tabs) > 0 {
		// Rename the first default tab to match the first desired tab
		if len(defaultTabs) > 0 {
			defaultTabID := strconv.Itoa(defaultTabs[0].ID)
			renameReq := screenTabUpdateRequest{Name: plan.Tabs[0].Name.ValueString()}
			if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs/%s", screenID, defaultTabID), renameReq, nil); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Rename Default Tab",
					"An error occurred while renaming the default tab on the screen.\n\n"+
						"Error: "+err.Error(),
				)
				return
			}
			plan.Tabs[0].ID = types.StringValue(defaultTabID)

			// Add fields to the first tab
			if err := r.addFieldsToTab(ctx, screenID, defaultTabID, plan.Tabs[0].Fields); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Add Fields to Tab",
					"An error occurred while adding fields to the first tab.\n\n"+
						"Error: "+err.Error(),
				)
				return
			}

			// Delete any extra default tabs
			for i := 1; i < len(defaultTabs); i++ {
				extraTabID := strconv.Itoa(defaultTabs[i].ID)
				if err := r.client.Delete(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs/%s", screenID, extraTabID), nil); err != nil {
					resp.Diagnostics.AddError(
						"Unable to Delete Extra Default Tab",
						"An error occurred while deleting an extra default tab.\n\n"+
							"Error: "+err.Error(),
					)
					return
				}
			}
		}

		// Create additional tabs
		for i := 1; i < len(plan.Tabs); i++ {
			tabCreateReq := screenTabCreateRequest{Name: plan.Tabs[i].Name.ValueString()}
			var tabResp screenTabAPIResponse
			if err := r.client.Post(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs", screenID), tabCreateReq, &tabResp); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Create Tab",
					fmt.Sprintf("An error occurred while creating tab %q.\n\n"+
						"Error: %s", plan.Tabs[i].Name.ValueString(), err.Error()),
				)
				return
			}

			tabID := strconv.Itoa(tabResp.ID)
			plan.Tabs[i].ID = types.StringValue(tabID)

			// Add fields to this tab
			if err := r.addFieldsToTab(ctx, screenID, tabID, plan.Tabs[i].Fields); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Add Fields to Tab",
					fmt.Sprintf("An error occurred while adding fields to tab %q.\n\n"+
						"Error: %s", plan.Tabs[i].Name.ValueString(), err.Error()),
				)
				return
			}
		}

		// If the first tab was created from a default tab, it should already be at position 0.
		// No reordering needed since we create tabs in order.
	} else {
		// No tabs desired — delete all default tabs
		for _, dt := range defaultTabs {
			dtID := strconv.Itoa(dt.ID)
			if err := r.client.Delete(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs/%s", screenID, dtID), nil); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Delete Default Tab",
					"An error occurred while deleting a default tab on the screen.\n\n"+
						"Error: "+err.Error(),
				)
				return
			}
		}
	}

	plan.Name = types.StringValue(screenResp.Name)
	plan.Description = types.StringValue(screenResp.Description)
	if screenResp.Description == "" && (plan.Description.IsNull() || plan.Description.ValueString() == "") {
		plan.Description = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ScreenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ScreenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	screenID := state.ID.ValueString()

	// Fetch the screen via paginated search
	screens, err := atlassian.Paginate[screenAPIResponse](ctx, r.client, fmt.Sprintf("/rest/api/3/screens?id=%s", screenID))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Screen",
			"An error occurred while calling the Jira API to read the screen.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	if len(screens) == 0 {
		// Screen was deleted outside of Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	screen := screens[0]
	state.ID = types.StringValue(strconv.Itoa(screen.ID))
	state.Name = types.StringValue(screen.Name)
	if screen.Description == "" {
		state.Description = types.StringNull()
	} else {
		state.Description = types.StringValue(screen.Description)
	}

	// Read tabs
	var apiTabs []screenTabAPIResponse
	if err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs", screenID), &apiTabs); err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Screen Tabs",
			"An error occurred while calling the Jira API to read the screen tabs.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state.Tabs = make([]ScreenTabModel, len(apiTabs))
	for i, tab := range apiTabs {
		tabID := strconv.Itoa(tab.ID)
		state.Tabs[i] = ScreenTabModel{
			ID:   types.StringValue(tabID),
			Name: types.StringValue(tab.Name),
		}

		// Read fields for this tab
		var apiFields []screenTabFieldAPIResponse
		if err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs/%s/fields", screenID, tabID), &apiFields); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Screen Tab Fields",
				fmt.Sprintf("An error occurred while reading fields for tab %q.\n\n"+
					"Error: %s", tab.Name, err.Error()),
			)
			return
		}

		if len(apiFields) > 0 {
			fields := make([]types.String, len(apiFields))
			for j, f := range apiFields {
				fields[j] = types.StringValue(f.ID)
			}
			state.Tabs[i].Fields = fields
		} else {
			state.Tabs[i].Fields = nil
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ScreenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ScreenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ScreenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	screenID := state.ID.ValueString()

	// Step 1: Update screen name and description
	updateReq := screenUpdateRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() {
		updateReq.Description = plan.Description.ValueString()
	}

	if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/screens/%s", screenID), updateReq, nil); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Screen",
			"An error occurred while calling the Jira API to update the screen.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Step 2: Read current tabs from the API
	var currentTabs []screenTabAPIResponse
	if err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs", screenID), &currentTabs); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Screen Tabs",
			"An error occurred while reading the current screen tabs for update.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Build a map of current tab IDs to their data
	currentTabMap := make(map[string]screenTabAPIResponse)
	for _, t := range currentTabs {
		currentTabMap[strconv.Itoa(t.ID)] = t
	}

	// Build a set of desired tab IDs (from plan state tabs that have IDs)
	desiredTabIDs := make(map[string]bool)
	for _, t := range plan.Tabs {
		if !t.ID.IsNull() && !t.ID.IsUnknown() && t.ID.ValueString() != "" {
			desiredTabIDs[t.ID.ValueString()] = true
		}
	}

	// Step 3: Match desired tabs to existing tabs by ID where possible
	// For tabs in state that have an ID matching a current tab, update them.
	// For tabs without an ID (new tabs), create them.
	// Delete current tabs that aren't in the desired set.

	// First, build a map from old state tab IDs
	oldTabIDs := make(map[string]bool)
	for _, t := range state.Tabs {
		if !t.ID.IsNull() && t.ID.ValueString() != "" {
			oldTabIDs[t.ID.ValueString()] = true
		}
	}

	// Delete tabs that exist on the API but are not in the desired plan
	// We match by position: desired tabs at the same index keep the state's tab ID
	// Strategy: delete all existing tabs and recreate, OR do a diff-based approach.
	// For correctness, let's use a position-based approach.

	// Build the desired tabs with IDs from old state where positions match
	desiredTabs := make([]struct {
		id     string // existing tab ID if reusing, empty if new
		name   string
		fields []string
	}, len(plan.Tabs))

	for i, t := range plan.Tabs {
		desiredTabs[i].name = t.Name.ValueString()
		for _, f := range t.Fields {
			desiredTabs[i].fields = append(desiredTabs[i].fields, f.ValueString())
		}
		// Try to match with old state by index
		if i < len(state.Tabs) && !state.Tabs[i].ID.IsNull() {
			desiredTabs[i].id = state.Tabs[i].ID.ValueString()
		}
	}

	// Collect tab IDs to keep
	keepIDs := make(map[string]bool)
	for _, dt := range desiredTabs {
		if dt.id != "" {
			keepIDs[dt.id] = true
		}
	}

	// Delete tabs that are not being kept
	for _, ct := range currentTabs {
		ctID := strconv.Itoa(ct.ID)
		if !keepIDs[ctID] {
			if err := r.client.Delete(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs/%s", screenID, ctID), nil); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Delete Tab",
					fmt.Sprintf("An error occurred while deleting tab %d.\n\n"+
						"Error: %s", ct.ID, err.Error()),
				)
				return
			}
		}
	}

	// Update or create tabs in order
	for i := range desiredTabs {
		if desiredTabs[i].id != "" {
			// Update existing tab name
			tabID := desiredTabs[i].id
			renameReq := screenTabUpdateRequest{Name: desiredTabs[i].name}
			if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs/%s", screenID, tabID), renameReq, nil); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Update Tab",
					fmt.Sprintf("An error occurred while updating tab %q.\n\n"+
						"Error: %s", desiredTabs[i].name, err.Error()),
				)
				return
			}

			// Sync fields for this tab
			if err := r.syncTabFields(ctx, screenID, tabID, desiredTabs[i].fields); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Update Tab Fields",
					fmt.Sprintf("An error occurred while syncing fields for tab %q.\n\n"+
						"Error: %s", desiredTabs[i].name, err.Error()),
				)
				return
			}

			plan.Tabs[i].ID = types.StringValue(tabID)
		} else {
			// Create new tab
			tabCreateReq := screenTabCreateRequest{Name: desiredTabs[i].name}
			var tabResp screenTabAPIResponse
			if err := r.client.Post(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs", screenID), tabCreateReq, &tabResp); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Create Tab",
					fmt.Sprintf("An error occurred while creating tab %q.\n\n"+
						"Error: %s", desiredTabs[i].name, err.Error()),
				)
				return
			}

			tabID := strconv.Itoa(tabResp.ID)
			plan.Tabs[i].ID = types.StringValue(tabID)

			// Add fields to new tab
			if err := r.addFieldsToTab(ctx, screenID, tabID, plan.Tabs[i].Fields); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Add Fields to Tab",
					fmt.Sprintf("An error occurred while adding fields to tab %q.\n\n"+
						"Error: %s", desiredTabs[i].name, err.Error()),
				)
				return
			}
		}
	}

	// Reorder tabs to match desired order
	for i := range plan.Tabs {
		tabID := plan.Tabs[i].ID.ValueString()
		if err := r.client.Post(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs/%s/move/%d", screenID, tabID, i), nil, nil); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Reorder Tab",
				fmt.Sprintf("An error occurred while moving tab %q to position %d.\n\n"+
					"Error: %s", plan.Tabs[i].Name.ValueString(), i, err.Error()),
			)
			return
		}
	}

	plan.ID = state.ID
	if plan.Description.IsNull() || plan.Description.ValueString() == "" {
		plan.Description = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ScreenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ScreenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/rest/api/3/screens/%s", state.ID.ValueString()), nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			// Already deleted, nothing to do.
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Screen",
			"An error occurred while calling the Jira API to delete the screen.\n\n"+
				"Error: "+err.Error(),
		)
	}
}

func (r *ScreenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// addFieldsToTab adds the given fields to a tab in order.
func (r *ScreenResource) addFieldsToTab(ctx context.Context, screenID, tabID string, fields []types.String) error {
	for _, f := range fields {
		fieldID := f.ValueString()
		addReq := screenTabFieldAddRequest{FieldID: fieldID}
		if err := r.client.Post(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs/%s/fields", screenID, tabID), addReq, nil); err != nil {
			return fmt.Errorf("adding field %q: %w", fieldID, err)
		}
	}

	// Reorder fields to match desired order
	for i, f := range fields {
		fieldID := f.ValueString()
		var moveReq screenTabFieldMoveRequest
		if i == 0 {
			moveReq.Position = "First"
		} else {
			moveReq.After = fields[i-1].ValueString()
		}
		if err := r.client.Post(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs/%s/fields/%s/move", screenID, tabID, fieldID), moveReq, nil); err != nil {
			return fmt.Errorf("moving field %q: %w", fieldID, err)
		}
	}

	return nil
}

// syncTabFields synchronizes the fields on a tab to match the desired list.
func (r *ScreenResource) syncTabFields(ctx context.Context, screenID, tabID string, desiredFields []string) error {
	// Read current fields
	var currentFields []screenTabFieldAPIResponse
	if err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs/%s/fields", screenID, tabID), &currentFields); err != nil {
		return fmt.Errorf("reading current fields: %w", err)
	}

	// Build sets for diff
	currentFieldSet := make(map[string]bool)
	for _, f := range currentFields {
		currentFieldSet[f.ID] = true
	}

	desiredFieldSet := make(map[string]bool)
	for _, f := range desiredFields {
		desiredFieldSet[f] = true
	}

	// Remove fields not in desired set
	for _, f := range currentFields {
		if !desiredFieldSet[f.ID] {
			if err := r.client.Delete(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs/%s/fields/%s", screenID, tabID, f.ID), nil); err != nil {
				return fmt.Errorf("removing field %q: %w", f.ID, err)
			}
		}
	}

	// Add fields that are in desired but not current
	for _, f := range desiredFields {
		if !currentFieldSet[f] {
			addReq := screenTabFieldAddRequest{FieldID: f}
			if err := r.client.Post(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs/%s/fields", screenID, tabID), addReq, nil); err != nil {
				return fmt.Errorf("adding field %q: %w", f, err)
			}
		}
	}

	// Reorder fields to match desired order
	for i, f := range desiredFields {
		var moveReq screenTabFieldMoveRequest
		if i == 0 {
			moveReq.Position = "First"
		} else {
			moveReq.After = desiredFields[i-1]
		}
		if err := r.client.Post(ctx, fmt.Sprintf("/rest/api/3/screens/%s/tabs/%s/fields/%s/move", screenID, tabID, f), moveReq, nil); err != nil {
			return fmt.Errorf("moving field %q: %w", f, err)
		}
	}

	return nil
}
