package jira

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

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
	_ resource.Resource                = &ScreenSchemeResource{}
	_ resource.ResourceWithImportState = &ScreenSchemeResource{}
)

// ScreenSchemeResource implements the atlassian_jira_screen_scheme resource.
type ScreenSchemeResource struct {
	client *atlassian.Client
}

// ScreenSchemeResourceModel describes the resource data model.
type ScreenSchemeResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	DefaultScreenID types.String `tfsdk:"default_screen_id"`
	CreateScreenID  types.String `tfsdk:"create_screen_id"`
	EditScreenID    types.String `tfsdk:"edit_screen_id"`
	ViewScreenID    types.String `tfsdk:"view_screen_id"`
}

// screenSchemeCreateRequest is the JSON body for POST /rest/api/3/screenscheme.
type screenSchemeCreateRequest struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description,omitempty"`
	Screens     screenSchemeScreensRequest `json:"screens"`
}

// screenSchemeScreensRequest represents the screens mapping in API requests.
type screenSchemeScreensRequest struct {
	Default int  `json:"default"`
	Create  *int `json:"create,omitempty"`
	Edit    *int `json:"edit,omitempty"`
	View    *int `json:"view,omitempty"`
}

// screenSchemeUpdateRequest is the JSON body for PUT /rest/api/3/screenscheme/{id}.
type screenSchemeUpdateRequest struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description,omitempty"`
	Screens     screenSchemeScreensRequest `json:"screens"`
}

// screenSchemeCreateResponse represents the JSON response from POST /rest/api/3/screenscheme.
type screenSchemeCreateResponse struct {
	ID int `json:"id"`
}

// screenSchemeAPIResponse represents a screen scheme from the paginated API.
type screenSchemeAPIResponse struct {
	ID          int                        `json:"id"`
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Screens     screenSchemeScreensAPIResp `json:"screens"`
}

// screenSchemeScreensAPIResp represents the screens mapping in API responses.
type screenSchemeScreensAPIResp struct {
	Default int `json:"default"`
	Create  int `json:"create,omitempty"`
	Edit    int `json:"edit,omitempty"`
	View    int `json:"view,omitempty"`
}

// NewScreenSchemeResource returns a new resource factory function.
func NewScreenSchemeResource() resource.Resource {
	return &ScreenSchemeResource{}
}

func (r *ScreenSchemeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_screen_scheme"
}

func (r *ScreenSchemeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira screen scheme.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The numeric ID of the screen scheme.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the screen scheme.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the screen scheme.",
				Optional:    true,
			},
			"default_screen_id": schema.StringAttribute{
				Description: "The ID of the default screen.",
				Required:    true,
			},
			"create_screen_id": schema.StringAttribute{
				Description: "The ID of the screen used for creating issues.",
				Optional:    true,
			},
			"edit_screen_id": schema.StringAttribute{
				Description: "The ID of the screen used for editing issues.",
				Optional:    true,
			},
			"view_screen_id": schema.StringAttribute{
				Description: "The ID of the screen used for viewing issues.",
				Optional:    true,
			},
		},
	}
}

func (r *ScreenSchemeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ScreenSchemeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ScreenSchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	screens, diags := r.buildScreensRequest(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := screenSchemeCreateRequest{
		Name:    plan.Name.ValueString(),
		Screens: screens,
	}
	if !plan.Description.IsNull() {
		createReq.Description = plan.Description.ValueString()
	}

	var createResp screenSchemeCreateResponse
	if err := r.client.Post(ctx, "/rest/api/3/screenscheme", createReq, &createResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Screen Scheme",
			"An error occurred while calling the Jira API to create the screen scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(strconv.Itoa(createResp.ID))

	// Read back to get the full state
	apiResp, err := r.readScreenScheme(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Created Screen Scheme",
			"The screen scheme was created but an error occurred while reading it back.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}
	if apiResp == nil {
		resp.Diagnostics.AddError(
			"Unable to Read Created Screen Scheme",
			"The screen scheme was created but could not be found when reading it back.",
		)
		return
	}

	r.mapScreenSchemeToState(&plan, apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ScreenSchemeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ScreenSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.readScreenScheme(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Screen Scheme",
			"An error occurred while calling the Jira API to read the screen scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	if apiResp == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.mapScreenSchemeToState(&state, apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ScreenSchemeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ScreenSchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ScreenSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	screens, diags := r.buildScreensRequest(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := screenSchemeUpdateRequest{
		Name:    plan.Name.ValueString(),
		Screens: screens,
	}
	if !plan.Description.IsNull() {
		updateReq.Description = plan.Description.ValueString()
	}

	if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/screenscheme/%s", state.ID.ValueString()), updateReq, nil); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Screen Scheme",
			"An error occurred while calling the Jira API to update the screen scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = state.ID

	// Read back to get the full state
	apiResp, err := r.readScreenScheme(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Updated Screen Scheme",
			"The screen scheme was updated but an error occurred while reading it back.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}
	if apiResp == nil {
		resp.Diagnostics.AddError(
			"Unable to Read Updated Screen Scheme",
			"The screen scheme was updated but could not be found when reading it back.",
		)
		return
	}

	r.mapScreenSchemeToState(&plan, apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ScreenSchemeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ScreenSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/rest/api/3/screenscheme/%s", state.ID.ValueString()), nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			// Already deleted, nothing to do.
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Screen Scheme",
			"An error occurred while calling the Jira API to delete the screen scheme.\n\n"+
				"Error: "+err.Error(),
		)
	}
}

func (r *ScreenSchemeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// buildScreensRequest converts the Terraform model screen IDs to the API request format.
func (r *ScreenSchemeResource) buildScreensRequest(plan ScreenSchemeResourceModel) (screenSchemeScreensRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	var screens screenSchemeScreensRequest

	defaultID, err := strconv.Atoi(plan.DefaultScreenID.ValueString())
	if err != nil {
		diags.AddAttributeError(
			path.Root("default_screen_id"),
			"Invalid Default Screen ID",
			fmt.Sprintf("default_screen_id must be a numeric string, got: %s", plan.DefaultScreenID.ValueString()),
		)
		return screens, diags
	}
	screens.Default = defaultID

	if !plan.CreateScreenID.IsNull() && plan.CreateScreenID.ValueString() != "" {
		createID, err := strconv.Atoi(plan.CreateScreenID.ValueString())
		if err != nil {
			diags.AddAttributeError(
				path.Root("create_screen_id"),
				"Invalid Create Screen ID",
				fmt.Sprintf("create_screen_id must be a numeric string, got: %s", plan.CreateScreenID.ValueString()),
			)
			return screens, diags
		}
		screens.Create = &createID
	}

	if !plan.EditScreenID.IsNull() && plan.EditScreenID.ValueString() != "" {
		editID, err := strconv.Atoi(plan.EditScreenID.ValueString())
		if err != nil {
			diags.AddAttributeError(
				path.Root("edit_screen_id"),
				"Invalid Edit Screen ID",
				fmt.Sprintf("edit_screen_id must be a numeric string, got: %s", plan.EditScreenID.ValueString()),
			)
			return screens, diags
		}
		screens.Edit = &editID
	}

	if !plan.ViewScreenID.IsNull() && plan.ViewScreenID.ValueString() != "" {
		viewID, err := strconv.Atoi(plan.ViewScreenID.ValueString())
		if err != nil {
			diags.AddAttributeError(
				path.Root("view_screen_id"),
				"Invalid View Screen ID",
				fmt.Sprintf("view_screen_id must be a numeric string, got: %s", plan.ViewScreenID.ValueString()),
			)
			return screens, diags
		}
		screens.View = &viewID
	}

	return screens, diags
}

// readScreenScheme reads a screen scheme by ID from the API.
func (r *ScreenSchemeResource) readScreenScheme(ctx context.Context, id string) (*screenSchemeAPIResponse, error) {
	schemes, err := atlassian.Paginate[screenSchemeAPIResponse](ctx, r.client, fmt.Sprintf("/rest/api/3/screenscheme?id=%s", id))
	if err != nil {
		return nil, err
	}

	if len(schemes) == 0 {
		return nil, nil
	}

	return &schemes[0], nil
}

// mapScreenSchemeToState maps an API response to the Terraform state model.
func (r *ScreenSchemeResource) mapScreenSchemeToState(state *ScreenSchemeResourceModel, apiResp *screenSchemeAPIResponse) {
	state.ID = types.StringValue(strconv.Itoa(apiResp.ID))
	state.Name = types.StringValue(apiResp.Name)
	if apiResp.Description == "" {
		state.Description = types.StringNull()
	} else {
		state.Description = types.StringValue(apiResp.Description)
	}
	state.DefaultScreenID = types.StringValue(strconv.Itoa(apiResp.Screens.Default))

	if apiResp.Screens.Create != 0 {
		state.CreateScreenID = types.StringValue(strconv.Itoa(apiResp.Screens.Create))
	} else {
		state.CreateScreenID = types.StringNull()
	}

	if apiResp.Screens.Edit != 0 {
		state.EditScreenID = types.StringValue(strconv.Itoa(apiResp.Screens.Edit))
	} else {
		state.EditScreenID = types.StringNull()
	}

	if apiResp.Screens.View != 0 {
		state.ViewScreenID = types.StringValue(strconv.Itoa(apiResp.Screens.View))
	} else {
		state.ViewScreenID = types.StringNull()
	}
}
