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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &FieldConfigurationSchemeResource{}
	_ resource.ResourceWithImportState = &FieldConfigurationSchemeResource{}
)

// FieldConfigurationSchemeResource implements the atlassian_jira_field_configuration_scheme resource.
type FieldConfigurationSchemeResource struct {
	client *atlassian.Client
}

// FieldConfigurationSchemeResourceModel describes the resource data model.
type FieldConfigurationSchemeResourceModel struct {
	ID          types.String                         `tfsdk:"id"`
	Name        types.String                         `tfsdk:"name"`
	Description types.String                         `tfsdk:"description"`
	Mappings    []FieldConfigurationSchemeMappingModel `tfsdk:"mapping"`
}

// FieldConfigurationSchemeMappingModel describes a single mapping in the field configuration scheme.
type FieldConfigurationSchemeMappingModel struct {
	IssueTypeID          types.String `tfsdk:"issue_type_id"`
	FieldConfigurationID types.String `tfsdk:"field_configuration_id"`
}

// fieldConfigSchemeCreateRequest is the JSON body for POST /rest/api/3/fieldconfigurationscheme.
type fieldConfigSchemeCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// fieldConfigSchemeUpdateRequest is the JSON body for PUT /rest/api/3/fieldconfigurationscheme/{id}.
type fieldConfigSchemeUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// fieldConfigSchemeCreateResponse represents the JSON response from POST /rest/api/3/fieldconfigurationscheme.
type fieldConfigSchemeCreateResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// fieldConfigSchemeAPIResponse represents a field configuration scheme entry from the paginated GET endpoint.
type fieldConfigSchemeAPIResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// fieldConfigSchemeMappingAPIResponse represents a mapping entry from the paginated GET endpoint.
type fieldConfigSchemeMappingAPIResponse struct {
	IssueTypeID          string `json:"issueTypeId"`
	FieldConfigurationID string `json:"fieldConfigurationId"`
}

// fieldConfigSchemeMappingsUpdateRequest is the JSON body for PUT /rest/api/3/fieldconfigurationscheme/{id}/mapping.
type fieldConfigSchemeMappingsUpdateRequest struct {
	Mappings []fieldConfigSchemeMappingAPIRequest `json:"mappings"`
}

// fieldConfigSchemeMappingAPIRequest represents a single mapping in the update request.
type fieldConfigSchemeMappingAPIRequest struct {
	IssueTypeID          string `json:"issueTypeId"`
	FieldConfigurationID string `json:"fieldConfigurationId"`
}

// fieldConfigSchemeMappingsDeleteRequest is the JSON body for POST /rest/api/3/fieldconfigurationscheme/{id}/mapping/delete.
type fieldConfigSchemeMappingsDeleteRequest struct {
	IssueTypeIDs []string `json:"issueTypeIds"`
}

// fieldConfigSchemeNameValidator validates that a string is at most 255 characters.
type fieldConfigSchemeNameValidator struct{}

func (v fieldConfigSchemeNameValidator) Description(_ context.Context) string {
	return "string length must be at most 255"
}

func (v fieldConfigSchemeNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v fieldConfigSchemeNameValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
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

// fieldConfigSchemeDescValidator validates that a string is at most 1024 characters.
type fieldConfigSchemeDescValidator struct{}

func (v fieldConfigSchemeDescValidator) Description(_ context.Context) string {
	return "string length must be at most 1024"
}

func (v fieldConfigSchemeDescValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v fieldConfigSchemeDescValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	if len(value) > 1024 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid String Length",
			fmt.Sprintf("Value must be at most 1024 characters, got %d.", len(value)),
		)
	}
}

// NewFieldConfigurationSchemeResource returns a new resource factory function.
func NewFieldConfigurationSchemeResource() resource.Resource {
	return &FieldConfigurationSchemeResource{}
}

func (r *FieldConfigurationSchemeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_field_configuration_scheme"
}

func (r *FieldConfigurationSchemeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira field configuration scheme.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The numeric ID of the field configuration scheme.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the field configuration scheme. Maximum 255 characters.",
				Required:    true,
				Validators: []validator.String{
					fieldConfigSchemeNameValidator{},
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the field configuration scheme. Maximum 1024 characters.",
				Optional:    true,
				Validators: []validator.String{
					fieldConfigSchemeDescValidator{},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"mapping": schema.ListNestedBlock{
				Description: "A mapping of issue type to field configuration.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"issue_type_id": schema.StringAttribute{
							Description: "The ID of the issue type. Use \"default\" for the default mapping.",
							Required:    true,
						},
						"field_configuration_id": schema.StringAttribute{
							Description: "The ID of the field configuration.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *FieldConfigurationSchemeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FieldConfigurationSchemeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FieldConfigurationSchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := fieldConfigSchemeCreateRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() {
		createReq.Description = plan.Description.ValueString()
	}

	var createResp fieldConfigSchemeCreateResponse
	if err := r.client.Post(ctx, "/rest/api/3/fieldconfigurationscheme", createReq, &createResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Field Configuration Scheme",
			"An error occurred while calling the Jira API to create the field configuration scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(createResp.ID)

	// If mappings are specified, set them
	if len(plan.Mappings) > 0 {
		diags := r.setMappings(ctx, plan.ID.ValueString(), plan.Mappings)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Read back the full resource
	readDiags := r.readFieldConfigurationScheme(ctx, &plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FieldConfigurationSchemeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FieldConfigurationSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readDiags := r.readFieldConfigurationScheme(ctx, &state)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FieldConfigurationSchemeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FieldConfigurationSchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state FieldConfigurationSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the scheme metadata
	updateReq := fieldConfigSchemeUpdateRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() {
		updateReq.Description = plan.Description.ValueString()
	}

	if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/fieldconfigurationscheme/%s", state.ID.ValueString()), updateReq, nil); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Field Configuration Scheme",
			"An error occurred while calling the Jira API to update the field configuration scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = state.ID

	// Update mappings: first delete existing non-default mappings, then set desired mappings
	diags := r.reconcileMappings(ctx, plan.ID.ValueString(), plan.Mappings)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read back the full resource
	readDiags := r.readFieldConfigurationScheme(ctx, &plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FieldConfigurationSchemeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FieldConfigurationSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/rest/api/3/fieldconfigurationscheme/%s", state.ID.ValueString()), nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			// Already deleted, nothing to do.
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Field Configuration Scheme",
			"An error occurred while calling the Jira API to delete the field configuration scheme.\n\n"+
				"Error: "+err.Error(),
		)
	}
}

func (r *FieldConfigurationSchemeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readFieldConfigurationScheme fetches the scheme and its mappings from the API and updates the model.
func (r *FieldConfigurationSchemeResource) readFieldConfigurationScheme(ctx context.Context, model *FieldConfigurationSchemeResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Fetch the scheme metadata
	schemes, err := atlassian.Paginate[fieldConfigSchemeAPIResponse](ctx, r.client, fmt.Sprintf("/rest/api/3/fieldconfigurationscheme?id=%s", model.ID.ValueString()))
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			diags.AddError(
				"Field Configuration Scheme Not Found",
				fmt.Sprintf("The field configuration scheme with ID %s was not found.", model.ID.ValueString()),
			)
			return diags
		}
		diags.AddError(
			"Unable to Read Field Configuration Scheme",
			"An error occurred while calling the Jira API to read the field configuration scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return diags
	}

	if len(schemes) == 0 {
		diags.AddError(
			"Field Configuration Scheme Not Found",
			fmt.Sprintf("The field configuration scheme with ID %s was not found.", model.ID.ValueString()),
		)
		return diags
	}

	scheme := schemes[0]
	model.ID = types.StringValue(scheme.ID)
	model.Name = types.StringValue(scheme.Name)
	if scheme.Description != "" {
		model.Description = types.StringValue(scheme.Description)
	} else if !model.Description.IsNull() {
		model.Description = types.StringNull()
	}

	// Fetch mappings (paginated)
	mappings, err := atlassian.Paginate[fieldConfigSchemeMappingAPIResponse](ctx, r.client, fmt.Sprintf("/rest/api/3/fieldconfigurationscheme/mapping?fieldConfigurationSchemeId=%s", model.ID.ValueString()))
	if err != nil {
		diags.AddError(
			"Unable to Read Field Configuration Scheme Mappings",
			"An error occurred while calling the Jira API to read the field configuration scheme mappings.\n\n"+
				"Error: "+err.Error(),
		)
		return diags
	}

	// Only populate mappings if the user has configured them (to avoid importing
	// API-created default mappings that weren't in the user's config)
	if model.Mappings != nil && len(model.Mappings) > 0 {
		modelMappings := make([]FieldConfigurationSchemeMappingModel, 0, len(mappings))
		for _, m := range mappings {
			modelMappings = append(modelMappings, FieldConfigurationSchemeMappingModel{
				IssueTypeID:          types.StringValue(m.IssueTypeID),
				FieldConfigurationID: types.StringValue(m.FieldConfigurationID),
			})
		}
		model.Mappings = modelMappings
	}

	return diags
}

// setMappings sends a PUT request to set the mappings for a field configuration scheme.
func (r *FieldConfigurationSchemeResource) setMappings(ctx context.Context, schemeID string, mappings []FieldConfigurationSchemeMappingModel) diag.Diagnostics {
	var diags diag.Diagnostics

	apiMappings := make([]fieldConfigSchemeMappingAPIRequest, 0, len(mappings))
	for _, m := range mappings {
		apiMappings = append(apiMappings, fieldConfigSchemeMappingAPIRequest{
			IssueTypeID:          m.IssueTypeID.ValueString(),
			FieldConfigurationID: m.FieldConfigurationID.ValueString(),
		})
	}

	updateReq := fieldConfigSchemeMappingsUpdateRequest{
		Mappings: apiMappings,
	}

	if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/fieldconfigurationscheme/%s/mapping", schemeID), updateReq, nil); err != nil {
		diags.AddError(
			"Unable to Set Field Configuration Scheme Mappings",
			"An error occurred while calling the Jira API to set the field configuration scheme mappings.\n\n"+
				"Error: "+err.Error(),
		)
	}

	return diags
}

// reconcileMappings deletes existing non-default mappings and then sets the desired mappings.
func (r *FieldConfigurationSchemeResource) reconcileMappings(ctx context.Context, schemeID string, desiredMappings []FieldConfigurationSchemeMappingModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get current mappings from the API
	currentMappings, err := atlassian.Paginate[fieldConfigSchemeMappingAPIResponse](ctx, r.client, fmt.Sprintf("/rest/api/3/fieldconfigurationscheme/mapping?fieldConfigurationSchemeId=%s", schemeID))
	if err != nil {
		diags.AddError(
			"Unable to Read Current Mappings",
			"An error occurred while reading the current field configuration scheme mappings.\n\n"+
				"Error: "+err.Error(),
		)
		return diags
	}

	// Collect non-"default" issue type IDs from current mappings to delete
	var nonDefaultIssueTypeIDs []string
	for _, m := range currentMappings {
		if m.IssueTypeID != "default" {
			nonDefaultIssueTypeIDs = append(nonDefaultIssueTypeIDs, m.IssueTypeID)
		}
	}

	// Delete all non-default mappings if any exist
	if len(nonDefaultIssueTypeIDs) > 0 {
		deleteReq := fieldConfigSchemeMappingsDeleteRequest{
			IssueTypeIDs: nonDefaultIssueTypeIDs,
		}
		if err := r.client.Post(ctx, fmt.Sprintf("/rest/api/3/fieldconfigurationscheme/%s/mapping/delete", schemeID), deleteReq, nil); err != nil {
			diags.AddError(
				"Unable to Delete Existing Mappings",
				"An error occurred while deleting existing field configuration scheme mappings.\n\n"+
					"Error: "+err.Error(),
			)
			return diags
		}
	}

	// Set the desired mappings if any
	if len(desiredMappings) > 0 {
		setDiags := r.setMappings(ctx, schemeID, desiredMappings)
		diags.Append(setDiags...)
	}

	return diags
}
