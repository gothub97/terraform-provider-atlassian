package jira

import (
	"context"
	"fmt"
	"net/url"
	"regexp"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &PriorityResource{}
	_ resource.ResourceWithImportState    = &PriorityResource{}
	_ resource.ResourceWithValidateConfig = &PriorityResource{}
)

// PriorityResource implements the atlassian_jira_priority resource.
type PriorityResource struct {
	client *atlassian.Client
}

// PriorityResourceModel describes the resource data model.
type PriorityResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	StatusColor types.String `tfsdk:"status_color"`
	IconURL     types.String `tfsdk:"icon_url"`
	AvatarID    types.Int64  `tfsdk:"avatar_id"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
	Self        types.String `tfsdk:"self"`
}

// priorityAPIRequest represents the JSON request body for create/update.
type priorityAPIRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	StatusColor string `json:"statusColor"`
	IconURL     string `json:"iconUrl,omitempty"`
}

// priorityCreateResponse represents the JSON response from POST /rest/api/3/priority.
type priorityCreateResponse struct {
	ID string `json:"id"`
}

// priorityAPIResponse represents the JSON response from GET /rest/api/3/priority/{id}.
type priorityAPIResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	StatusColor string `json:"statusColor"`
	IconURL     string `json:"iconUrl"`
	IsDefault   bool   `json:"isDefault"`
	Self        string `json:"self"`
}

// statusColorValidator validates that a string matches the hex color pattern.
type statusColorValidator struct{}

func (v statusColorValidator) Description(_ context.Context) string {
	return "value must be a valid hex color (e.g., #FFF or #FF0000)"
}

func (v statusColorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v statusColorValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	matched, _ := regexp.MatchString(`^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$`, value)
	if !matched {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Status Color",
			fmt.Sprintf("status_color must be a valid hex color (e.g., #FFF or #FF0000), got: %s", value),
		)
	}
}

// NewPriorityResource returns a new resource factory function.
func NewPriorityResource() resource.Resource {
	return &PriorityResource{}
}

func (r *PriorityResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_priority"
}

func (r *PriorityResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira priority.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the priority.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the priority.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the priority.",
				Optional:    true,
			},
			"status_color": schema.StringAttribute{
				Description: "The color of the priority status in hex format (e.g., #FF0000).",
				Required:    true,
				Validators: []validator.String{
					statusColorValidator{},
				},
			},
			"icon_url": schema.StringAttribute{
				Description: "The URL of the icon for the priority. Conflicts with avatar_id.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"avatar_id": schema.Int64Attribute{
				Description: "The ID of the avatar for the priority. Conflicts with icon_url.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_default": schema.BoolAttribute{
				Description: "Whether the priority is the default priority.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"self": schema.StringAttribute{
				Description: "The URL of the priority.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *PriorityResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PriorityResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data PriorityResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.IconURL.IsNull() && !data.AvatarID.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("icon_url"),
			"Conflicting Attributes",
			"icon_url and avatar_id are mutually exclusive. Please specify only one.",
		)
	}
}

func (r *PriorityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PriorityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := priorityAPIRequest{
		Name:        plan.Name.ValueString(),
		StatusColor: plan.StatusColor.ValueString(),
	}
	if !plan.Description.IsNull() {
		apiReq.Description = plan.Description.ValueString()
	}
	if !plan.IconURL.IsNull() {
		apiReq.IconURL = plan.IconURL.ValueString()
	}

	var createResp priorityCreateResponse
	if err := r.client.Post(ctx, "/rest/api/3/priority", apiReq, &createResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Priority",
			"An error occurred while creating the Jira priority.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(createResp.ID)

	// Read back the full resource to get computed fields
	diags := r.readPriority(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PriorityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PriorityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := r.readPriority(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PriorityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PriorityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PriorityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := priorityAPIRequest{
		Name:        plan.Name.ValueString(),
		StatusColor: plan.StatusColor.ValueString(),
	}
	if !plan.Description.IsNull() {
		apiReq.Description = plan.Description.ValueString()
	}
	if !plan.IconURL.IsNull() {
		apiReq.IconURL = plan.IconURL.ValueString()
	}

	if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/priority/%s", state.ID.ValueString()), apiReq, nil); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Priority",
			"An error occurred while updating the Jira priority.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = state.ID

	// Read back to get computed fields
	diags := r.readPriority(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PriorityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PriorityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The Jira API requires a replaceWith parameter when deleting a priority.
	// Find another priority to use as the replacement.
	var priorities []priorityAPIResponse
	if err := r.client.Get(ctx, "/rest/api/3/priority", &priorities); err != nil {
		resp.Diagnostics.AddError(
			"Unable to List Priorities",
			"An error occurred while listing priorities to find a replacement for deletion.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	var replaceWithID string
	for _, p := range priorities {
		if p.ID != state.ID.ValueString() {
			replaceWithID = p.ID
			break
		}
	}

	if replaceWithID == "" {
		resp.Diagnostics.AddError(
			"Unable to Delete Priority",
			"Cannot delete the last remaining priority. At least one other priority must exist.",
		)
		return
	}

	// Remove the priority from all priority schemes before deletion.
	// Jira Cloud auto-adds new priorities to the default scheme, which blocks deletion.
	if diags := r.removePriorityFromSchemes(ctx, state.ID.ValueString()); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	location, err := r.client.DeleteWithRedirect(ctx, fmt.Sprintf("/rest/api/3/priority/%s?replaceWith=%s", state.ID.ValueString(), replaceWithID))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Priority",
			"An error occurred while deleting the Jira priority.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// If we got a redirect location, we need to poll the task
	if location != "" {
		taskPath, err := extractTaskPath(location)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Parse Task URL",
				"An error occurred while parsing the task URL from the delete response.\n\n"+
					"Error: "+err.Error(),
			)
			return
		}

		if err := r.client.WaitForTask(ctx, taskPath, 0); err != nil {
			resp.Diagnostics.AddError(
				"Delete Task Failed",
				"The async delete task failed.\n\n"+
					"Error: "+err.Error(),
			)
			return
		}
	}
}

func (r *PriorityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readPriority fetches the priority from the API and updates the model.
func (r *PriorityResource) readPriority(ctx context.Context, model *PriorityResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	var apiResp priorityAPIResponse
	if err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/priority/%s", model.ID.ValueString()), &apiResp); err != nil {
		diags.AddError(
			"Unable to Read Priority",
			"An error occurred while reading the Jira priority.\n\n"+
				"Error: "+err.Error(),
		)
		return diags
	}

	model.ID = types.StringValue(apiResp.ID)
	model.Name = types.StringValue(apiResp.Name)
	if apiResp.Description != "" {
		model.Description = types.StringValue(apiResp.Description)
	} else if !model.Description.IsNull() {
		model.Description = types.StringNull()
	}
	model.StatusColor = types.StringValue(apiResp.StatusColor)
	model.IconURL = types.StringValue(apiResp.IconURL)
	// The API does not return avatarId in the GET response, so preserve existing value or set null.
	if model.AvatarID.IsUnknown() {
		model.AvatarID = types.Int64Null()
	}
	model.IsDefault = types.BoolValue(apiResp.IsDefault)
	model.Self = types.StringValue(apiResp.Self)

	return diags
}

// extractTaskPath extracts the path component from a full task URL.
func extractTaskPath(locationURL string) (string, error) {
	u, err := url.Parse(locationURL)
	if err != nil {
		return "", fmt.Errorf("parsing task URL %q: %w", locationURL, err)
	}
	return u.Path, nil
}

// prioritySchemeListResponse represents the paginated list of priority schemes.
type prioritySchemeListResponse struct {
	Values []prioritySchemeEntry `json:"values"`
	IsLast bool                  `json:"isLast"`
}

type prioritySchemeEntry struct {
	ID string `json:"id"`
}

// prioritySchemePrioritiesResponse represents the paginated list of priorities in a scheme.
type prioritySchemePrioritiesResponse struct {
	Values []struct {
		ID string `json:"id"`
	} `json:"values"`
	IsLast bool `json:"isLast"`
}

// prioritySchemeUpdateRequest represents the PUT body for updating a priority scheme.
type prioritySchemeUpdateRequest struct {
	Priorities *prioritySchemeChanges `json:"priorities,omitempty"`
}

type prioritySchemeChanges struct {
	Remove *priorityIDList `json:"remove,omitempty"`
}

type priorityIDList struct {
	IDs []int64 `json:"ids"`
}

// removePriorityFromSchemes removes the priority from all priority schemes.
func (r *PriorityResource) removePriorityFromSchemes(ctx context.Context, priorityID string) diag.Diagnostics {
	var diags diag.Diagnostics

	var schemes prioritySchemeListResponse
	if err := r.client.Get(ctx, "/rest/api/3/priorityscheme", &schemes); err != nil {
		diags.AddError(
			"Unable to List Priority Schemes",
			"An error occurred while listing priority schemes.\n\n"+
				"Error: "+err.Error(),
		)
		return diags
	}

	for _, scheme := range schemes.Values {
		// Check if this scheme contains the priority
		var schemePriorities prioritySchemePrioritiesResponse
		if err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/priorityscheme/%s/priorities", scheme.ID), &schemePriorities); err != nil {
			continue
		}

		found := false
		for _, p := range schemePriorities.Values {
			if p.ID == priorityID {
				found = true
				break
			}
		}

		if !found {
			continue
		}

		// Parse the priority ID as int64 for the API
		var pidInt int64
		if _, err := fmt.Sscanf(priorityID, "%d", &pidInt); err != nil {
			diags.AddError(
				"Invalid Priority ID",
				fmt.Sprintf("Could not parse priority ID %q as integer: %v", priorityID, err),
			)
			return diags
		}

		updateReq := prioritySchemeUpdateRequest{
			Priorities: &prioritySchemeChanges{
				Remove: &priorityIDList{
					IDs: []int64{pidInt},
				},
			},
		}

		if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/priorityscheme/%s", scheme.ID), updateReq, nil); err != nil {
			diags.AddError(
				"Unable to Remove Priority from Scheme",
				fmt.Sprintf("An error occurred while removing priority %s from scheme %s.\n\nError: %s", priorityID, scheme.ID, err.Error()),
			)
			return diags
		}
	}

	return diags
}
