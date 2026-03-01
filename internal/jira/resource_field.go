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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &FieldResource{}
	_ resource.ResourceWithImportState = &FieldResource{}
)

// FieldResource implements the atlassian_jira_field resource.
type FieldResource struct {
	client *atlassian.Client
}

// FieldResourceModel describes the resource data model.
type FieldResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
	SearcherKey types.String `tfsdk:"searcher_key"`
}

// fieldCreateRequest is the JSON body for POST /rest/api/3/field.
type fieldCreateRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	SearcherKey string `json:"searcherKey,omitempty"`
}

// fieldUpdateRequest is the JSON body for PUT /rest/api/3/field/{fieldId}.
type fieldUpdateRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	SearcherKey string `json:"searcherKey,omitempty"`
}

// fieldCreateResponse represents the JSON response from POST /rest/api/3/field.
type fieldCreateResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// fieldSearchItem represents a single field in the field search API response.
type fieldSearchItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Schema      struct {
		Type     string `json:"type"`
		Custom   string `json:"custom"`
		CustomID int64  `json:"customId"`
	} `json:"schema"`
}

// NewFieldResource returns a new resource factory function.
func NewFieldResource() resource.Resource {
	return &FieldResource{}
}

func (r *FieldResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_field"
}

func (r *FieldResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira custom field.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the field (e.g., customfield_10001).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the custom field.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the custom field (e.g., com.atlassian.jira.plugin.system.customfieldtypes:float). Cannot be changed after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description of the custom field.",
				Optional:    true,
			},
			"searcher_key": schema.StringAttribute{
				Description: "The searcher key for the custom field (e.g., com.atlassian.jira.plugin.system.customfieldtypes:exactnumber).",
				Optional:    true,
			},
		},
	}
}

func (r *FieldResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := fieldCreateRequest{
		Name: plan.Name.ValueString(),
		Type: plan.Type.ValueString(),
	}
	if !plan.Description.IsNull() {
		createReq.Description = plan.Description.ValueString()
	}
	if !plan.SearcherKey.IsNull() {
		createReq.SearcherKey = plan.SearcherKey.ValueString()
	}

	var createResp fieldCreateResponse
	if err := r.client.Post(ctx, "/rest/api/3/field", createReq, &createResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Field",
			"An error occurred while calling the Jira API to create the field.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(createResp.ID)
	plan.Name = types.StringValue(createResp.Name)
	if createResp.Description != "" {
		plan.Description = types.StringValue(createResp.Description)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	field, diags := r.readField(ctx, state.ID.ValueString())
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if field == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	mapFieldSearchItemToState(field, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state FieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := fieldUpdateRequest{}
	if !plan.Name.IsNull() {
		updateReq.Name = plan.Name.ValueString()
	}
	if !plan.Description.IsNull() {
		updateReq.Description = plan.Description.ValueString()
	}
	if !plan.SearcherKey.IsNull() {
		updateReq.SearcherKey = plan.SearcherKey.ValueString()
	}

	apiPath := fmt.Sprintf("/rest/api/3/field/%s", state.ID.ValueString())
	if err := r.client.Put(ctx, apiPath, updateReq, nil); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Field",
			"An error occurred while calling the Jira API to update the field.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Read back the field to get updated state
	field, diags := r.readField(ctx, state.ID.ValueString())
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if field == nil {
		resp.Diagnostics.AddError(
			"Unable to Read Updated Field",
			"The field was updated but could not be found when reading it back.",
		)
		return
	}

	mapFieldSearchItemToState(field, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPath := fmt.Sprintf("/rest/api/3/field/%s", state.ID.ValueString())
	err := r.client.Delete(ctx, apiPath, nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			// Already deleted, nothing to do.
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Field",
			"An error occurred while calling the Jira API to delete the field.\n\n"+
				"Error: "+err.Error(),
		)
	}
}

func (r *FieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readField reads a single field by ID using the field search API.
// Returns nil if the field is not found (404 or empty results).
func (r *FieldResource) readField(ctx context.Context, fieldID string) (*fieldSearchItem, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiPath := fmt.Sprintf("/rest/api/3/field/search?id=%s", fieldID)

	var searchResp atlassian.PaginatedResponse[fieldSearchItem]
	if err := r.client.Get(ctx, apiPath, &searchResp); err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return nil, diags
		}
		diags.AddError(
			"Unable to Read Field",
			"An error occurred while calling the Jira API to read the field.\n\n"+
				"Error: "+err.Error(),
		)
		return nil, diags
	}

	if len(searchResp.Values) == 0 {
		return nil, diags
	}

	return &searchResp.Values[0], diags
}

// mapFieldSearchItemToState maps a field search API response to the Terraform state model.
func mapFieldSearchItemToState(field *fieldSearchItem, state *FieldResourceModel) {
	state.ID = types.StringValue(field.ID)
	state.Name = types.StringValue(field.Name)
	if field.Schema.Custom != "" {
		state.Type = types.StringValue(field.Schema.Custom)
	}
	if field.Description != "" {
		state.Description = types.StringValue(field.Description)
	} else {
		state.Description = types.StringNull()
	}
}
