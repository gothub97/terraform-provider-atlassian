package jira

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &FieldConfigurationResource{}
	_ resource.ResourceWithImportState = &FieldConfigurationResource{}
)

// FieldConfigurationResource implements the atlassian_jira_field_configuration resource.
type FieldConfigurationResource struct {
	client *atlassian.Client
}

// FieldConfigurationResourceModel describes the resource data model.
type FieldConfigurationResourceModel struct {
	ID          types.String                  `tfsdk:"id"`
	Name        types.String                  `tfsdk:"name"`
	Description types.String                  `tfsdk:"description"`
	FieldItems  []FieldConfigurationItemModel `tfsdk:"field_item"`
}

// FieldConfigurationItemModel describes a single field item in the field configuration.
type FieldConfigurationItemModel struct {
	FieldID     types.String `tfsdk:"field_id"`
	IsRequired  types.Bool   `tfsdk:"is_required"`
	IsHidden    types.Bool   `tfsdk:"is_hidden"`
	Description types.String `tfsdk:"description"`
	Renderer    types.String `tfsdk:"renderer"`
}

// fieldConfigurationCreateRequest is the JSON body for POST /rest/api/3/fieldconfiguration.
type fieldConfigurationCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// fieldConfigurationUpdateRequest is the JSON body for PUT /rest/api/3/fieldconfiguration/{id}.
type fieldConfigurationUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// fieldConfigurationCreateResponse represents the JSON response from POST /rest/api/3/fieldconfiguration.
type fieldConfigurationCreateResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// fieldConfigurationAPIResponse represents a field configuration entry returned from the paginated GET endpoint.
type fieldConfigurationAPIResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// fieldConfigurationItemAPIResponse represents a field item in a field configuration.
type fieldConfigurationItemAPIResponse struct {
	ID          string `json:"id"`
	IsRequired  bool   `json:"isRequired"`
	IsHidden    bool   `json:"isHidden"`
	Description string `json:"description"`
	Renderer    string `json:"renderer"`
}

// fieldConfigurationItemsUpdateRequest is the JSON body for PUT /rest/api/3/fieldconfiguration/{id}/fields.
type fieldConfigurationItemsUpdateRequest struct {
	FieldConfigurationItems []fieldConfigurationItemAPIRequest `json:"fieldConfigurationItems"`
}

// fieldConfigurationItemAPIRequest represents a single item in the update request.
type fieldConfigurationItemAPIRequest struct {
	ID          string `json:"id"`
	IsRequired  bool   `json:"isRequired"`
	IsHidden    bool   `json:"isHidden"`
	Description string `json:"description,omitempty"`
	Renderer    string `json:"renderer,omitempty"`
}

// fieldConfigNameValidator validates that a string is at most 255 characters.
type fieldConfigNameValidator struct{}

func (v fieldConfigNameValidator) Description(_ context.Context) string {
	return "string length must be at most 255"
}

func (v fieldConfigNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v fieldConfigNameValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	if len(value) > 255 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid String Length",
			fmt.Sprintf("Value must be at most 255 characters, got %d.", len(value)),
		)
	}
}

// NewFieldConfigurationResource returns a new resource factory function.
func NewFieldConfigurationResource() resource.Resource {
	return &FieldConfigurationResource{}
}

func (r *FieldConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_field_configuration"
}

func (r *FieldConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira field configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The numeric ID of the field configuration.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the field configuration. Maximum 255 characters.",
				Required:    true,
				Validators: []validator.String{
					fieldConfigNameValidator{},
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the field configuration. Maximum 255 characters.",
				Optional:    true,
				Validators: []validator.String{
					fieldConfigNameValidator{},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"field_item": schema.ListNestedBlock{
				Description: "A field item in the field configuration.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"field_id": schema.StringAttribute{
							Description: "The ID of the field (e.g., summary, customfield_10001).",
							Required:    true,
						},
						"is_required": schema.BoolAttribute{
							Description: "Whether the field is required.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"is_hidden": schema.BoolAttribute{
							Description: "Whether the field is hidden.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"description": schema.StringAttribute{
							Description: "The description of the field in this configuration.",
							Optional:    true,
						},
						"renderer": schema.StringAttribute{
							Description: "The renderer type for the field.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (r *FieldConfigurationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FieldConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FieldConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := fieldConfigurationCreateRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() {
		createReq.Description = plan.Description.ValueString()
	}

	var createResp fieldConfigurationCreateResponse
	if err := r.client.Post(ctx, "/rest/api/3/fieldconfiguration", createReq, &createResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Field Configuration",
			"An error occurred while calling the Jira API to create the field configuration.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d", createResp.ID))

	// If field items are specified, update them
	if len(plan.FieldItems) > 0 {
		diags := r.updateFieldItems(ctx, plan.ID.ValueString(), plan.FieldItems)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Read back the full resource to get the current state
	readDiags := r.readFieldConfiguration(ctx, &plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FieldConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FieldConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readDiags := r.readFieldConfiguration(ctx, &state)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FieldConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FieldConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state FieldConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the field configuration metadata
	updateReq := fieldConfigurationUpdateRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() {
		updateReq.Description = plan.Description.ValueString()
	}

	if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/fieldconfiguration/%s", state.ID.ValueString()), updateReq, nil); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Field Configuration",
			"An error occurred while calling the Jira API to update the field configuration.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = state.ID

	// Update field items if specified
	if len(plan.FieldItems) > 0 {
		diags := r.updateFieldItems(ctx, plan.ID.ValueString(), plan.FieldItems)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Read back the full resource
	readDiags := r.readFieldConfiguration(ctx, &plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FieldConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FieldConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/rest/api/3/fieldconfiguration/%s", state.ID.ValueString()), nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			// Already deleted, nothing to do.
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Field Configuration",
			"An error occurred while calling the Jira API to delete the field configuration.\n\n"+
				"Error: "+err.Error(),
		)
	}
}

func (r *FieldConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readFieldConfiguration fetches the field configuration and its items from the API and updates the model.
func (r *FieldConfigurationResource) readFieldConfiguration(ctx context.Context, model *FieldConfigurationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Fetch the field configuration metadata
	configs, err := atlassian.Paginate[fieldConfigurationAPIResponse](ctx, r.client, fmt.Sprintf("/rest/api/3/fieldconfiguration?id=%s", model.ID.ValueString()))
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			diags.AddError(
				"Field Configuration Not Found",
				fmt.Sprintf("The field configuration with ID %s was not found.", model.ID.ValueString()),
			)
			return diags
		}
		diags.AddError(
			"Unable to Read Field Configuration",
			"An error occurred while calling the Jira API to read the field configuration.\n\n"+
				"Error: "+err.Error(),
		)
		return diags
	}

	if len(configs) == 0 {
		diags.AddError(
			"Field Configuration Not Found",
			fmt.Sprintf("The field configuration with ID %s was not found.", model.ID.ValueString()),
		)
		return diags
	}

	config := configs[0]
	model.ID = types.StringValue(fmt.Sprintf("%d", config.ID))
	model.Name = types.StringValue(config.Name)
	if config.Description != "" {
		model.Description = types.StringValue(config.Description)
	} else if !model.Description.IsNull() {
		model.Description = types.StringNull()
	}

	// Fetch field items (paginated)
	items, err := atlassian.Paginate[fieldConfigurationItemAPIResponse](ctx, r.client, fmt.Sprintf("/rest/api/3/fieldconfiguration/%s/fields", model.ID.ValueString()))
	if err != nil {
		diags.AddError(
			"Unable to Read Field Configuration Items",
			"An error occurred while calling the Jira API to read the field configuration items.\n\n"+
				"Error: "+err.Error(),
		)
		return diags
	}

	// Only populate field_items if the user has configured them (to avoid importing all default items)
	if model.FieldItems != nil {
		// Build a map of API items by field ID for quick lookup
		apiItemMap := make(map[string]fieldConfigurationItemAPIResponse, len(items))
		for _, item := range items {
			apiItemMap[item.ID] = item
		}

		// Update the model's field items with values from the API
		updatedItems := make([]FieldConfigurationItemModel, 0, len(model.FieldItems))
		for _, planItem := range model.FieldItems {
			fieldID := planItem.FieldID.ValueString()
			if apiItem, ok := apiItemMap[fieldID]; ok {
				updatedItems = append(updatedItems, FieldConfigurationItemModel{
					FieldID:    types.StringValue(apiItem.ID),
					IsRequired: types.BoolValue(apiItem.IsRequired),
					IsHidden:   types.BoolValue(apiItem.IsHidden),
					Description: func() types.String {
						if apiItem.Description != "" {
							return types.StringValue(apiItem.Description)
						}
						if !planItem.Description.IsNull() {
							return types.StringNull()
						}
						return types.StringNull()
					}(),
					Renderer: func() types.String {
						if !planItem.Renderer.IsNull() && apiItem.Renderer != "" {
							return types.StringValue(apiItem.Renderer)
						}
						return types.StringNull()
					}(),
				})
			} else {
				// Field not found in the API response; keep the plan value
				updatedItems = append(updatedItems, planItem)
			}
		}
		model.FieldItems = updatedItems
	}

	return diags
}

// updateFieldItems sends a PUT request to update the field items for a field configuration.
func (r *FieldConfigurationResource) updateFieldItems(ctx context.Context, configID string, items []FieldConfigurationItemModel) diag.Diagnostics {
	var diags diag.Diagnostics

	apiItems := make([]fieldConfigurationItemAPIRequest, 0, len(items))
	for _, item := range items {
		apiItem := fieldConfigurationItemAPIRequest{
			ID:         item.FieldID.ValueString(),
			IsRequired: item.IsRequired.ValueBool(),
			IsHidden:   item.IsHidden.ValueBool(),
		}
		if !item.Description.IsNull() {
			apiItem.Description = item.Description.ValueString()
		}
		if !item.Renderer.IsNull() {
			apiItem.Renderer = item.Renderer.ValueString()
		}
		apiItems = append(apiItems, apiItem)
	}

	updateReq := fieldConfigurationItemsUpdateRequest{
		FieldConfigurationItems: apiItems,
	}

	if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/fieldconfiguration/%s/fields", configID), updateReq, nil); err != nil {
		diags.AddError(
			"Unable to Update Field Configuration Items",
			"An error occurred while calling the Jira API to update the field configuration items.\n\n"+
				"Error: "+err.Error(),
		)
	}

	return diags
}
