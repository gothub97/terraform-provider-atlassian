package confluence

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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
	_ resource.Resource                = &SpaceResource{}
	_ resource.ResourceWithImportState = &SpaceResource{}
)

// SpaceResource implements the atlassian_confluence_space resource. It uses a
// hybrid of the Confluence v2 REST API (create/read) and v1 REST API
// (update/delete), since v2 does not expose space update or delete operations.
type SpaceResource struct {
	client *atlassian.Client
}

// SpaceResourceModel describes the resource data model.
type SpaceResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Status      types.String `tfsdk:"status"`
	HomepageID  types.String `tfsdk:"homepage_id"`
	URL         types.String `tfsdk:"url"`
}

// spaceV2 is the v2 SpaceBulk JSON shape returned by the Confluence v2 API
// (POST/GET /wiki/api/v2/spaces). It is the package-shared representation of a
// space and is reused by the data source — keep the field/tag set stable and
// coordinate with the team before changing it.
type spaceV2 struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	HomepageID  string `json:"homepageId"`
	Description struct {
		Plain struct {
			Value          string `json:"value"`
			Representation string `json:"representation"`
		} `json:"plain"`
	} `json:"description"`
	Links struct {
		Base  string `json:"base"`
		WebUI string `json:"webui"`
	} `json:"_links"`
}

// spaceV2CreateRequest is the body for POST /wiki/api/v2/spaces.
type spaceV2CreateRequest struct {
	Key         string                   `json:"key"`
	Name        string                   `json:"name"`
	Description *spaceV2DescriptionInput `json:"description,omitempty"`
}

type spaceV2DescriptionInput struct {
	Value          string `json:"value"`
	Representation string `json:"representation"`
}

// spaceV1UpdateRequest is the body for PUT /wiki/rest/api/space/{key}.
type spaceV1UpdateRequest struct {
	Key         string                   `json:"key"`
	Name        string                   `json:"name"`
	Description *spaceV1DescriptionInput `json:"description,omitempty"`
}

type spaceV1DescriptionInput struct {
	Plain spaceV1DescriptionPlain `json:"plain"`
}

type spaceV1DescriptionPlain struct {
	Value          string `json:"value"`
	Representation string `json:"representation"`
}

// NewSpaceResource returns a new resource factory function.
func NewSpaceResource() resource.Resource {
	return &SpaceResource{}
}

func (r *SpaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_confluence_space"
}

func (r *SpaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Confluence space.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The numeric ID of the space.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "The key of the space. Changing this forces a new space to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the space.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the space.",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the space (e.g. \"global\").",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The status of the space (\"current\" or \"archived\").",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"homepage_id": schema.StringAttribute{
				Description: "The ID of the space's homepage.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"url": schema.StringAttribute{
				Description: "The URL of the space.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *SpaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, errMsg, ok := configureClient(req.ProviderData)
	if !ok {
		if errMsg != "" {
			resp.Diagnostics.AddError("Unexpected Resource Configure Type", errMsg)
		}
		return
	}
	r.client = client
}

func (r *SpaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SpaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := spaceV2CreateRequest{
		Key:  plan.Key.ValueString(),
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() {
		apiReq.Description = &spaceV2DescriptionInput{
			Value:          plan.Description.ValueString(),
			Representation: "plain",
		}
	}

	var created spaceV2
	if err := r.client.Post(ctx, fmt.Sprintf("%s/spaces", apiV2Base), apiReq, &created); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Space",
			"An error occurred while creating the Confluence space.\n\nError: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(created.ID)

	// Read back the full resource to populate computed fields.
	resp.Diagnostics.Append(r.readSpace(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SpaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SpaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sp, err := r.getSpaceByID(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Space",
			"An error occurred while reading the Confluence space.\n\nError: "+err.Error(),
		)
		return
	}

	applySpaceToModel(&state, sp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SpaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SpaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state SpaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The v2 API has no space update; use the v1 API keyed by space key.
	apiReq := spaceV1UpdateRequest{
		Key:  state.Key.ValueString(),
		Name: plan.Name.ValueString(),
	}
	descValue := ""
	if !plan.Description.IsNull() {
		descValue = plan.Description.ValueString()
	}
	apiReq.Description = &spaceV1DescriptionInput{
		Plain: spaceV1DescriptionPlain{
			Value:          descValue,
			Representation: "plain",
		},
	}

	if err := r.client.Put(ctx, fmt.Sprintf("%s/space/%s", apiV1Base, state.Key.ValueString()), apiReq, nil); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Space",
			"An error occurred while updating the Confluence space.\n\nError: "+err.Error(),
		)
		return
	}

	plan.ID = state.ID

	// Read back to refresh computed fields.
	resp.Diagnostics.Append(r.readSpace(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SpaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SpaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The v1 delete returns 202 Accepted with a long-running task descriptor.
	var task longTaskResponse
	if err := r.client.Delete(ctx, fmt.Sprintf("%s/space/%s", apiV1Base, state.Key.ValueString()), &task); err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			// Already gone.
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Space",
			"An error occurred while deleting the Confluence space.\n\nError: "+err.Error(),
		)
		return
	}

	taskID := task.ID
	if taskID == "" && task.Links.Status != "" {
		segments := strings.Split(strings.Trim(task.Links.Status, "/"), "/")
		taskID = segments[len(segments)-1]
	}

	if taskID != "" {
		if err := waitForLongTask(ctx, r.client, taskID); err != nil {
			resp.Diagnostics.AddError(
				"Space Delete Task Failed",
				"The asynchronous space delete task did not complete successfully.\n\nError: "+err.Error(),
			)
			return
		}
	}
}

func (r *SpaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by space key. Resolve the numeric v2 id from the key, then set both
	// id and key into state; Read populates the remaining attributes.
	key := req.ID

	var listResp struct {
		Results []spaceV2 `json:"results"`
	}
	listPath := fmt.Sprintf("%s/spaces?keys=%s&description-format=plain", apiV2Base, url.QueryEscape(key))
	if err := r.client.Get(ctx, listPath, &listResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Import Space",
			"An error occurred while looking up the Confluence space by key.\n\nError: "+err.Error(),
		)
		return
	}

	if len(listResp.Results) == 0 {
		resp.Diagnostics.AddError(
			"Unable to Import Space",
			fmt.Sprintf("No Confluence space found with key %q.", key),
		)
		return
	}

	sp := listResp.Results[0]
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), sp.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), sp.Key)...)
}

// getSpaceByID fetches a space by its numeric v2 id (description rendered as plain).
func (r *SpaceResource) getSpaceByID(ctx context.Context, id string) (*spaceV2, error) {
	var sp spaceV2
	reqPath := fmt.Sprintf("%s/spaces/%s?description-format=plain", apiV2Base, id)
	if err := r.client.Get(ctx, reqPath, &sp); err != nil {
		return nil, err
	}
	return &sp, nil
}

// readSpace fetches the space by the model's id and maps it onto the model.
func (r *SpaceResource) readSpace(ctx context.Context, model *SpaceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	sp, err := r.getSpaceByID(ctx, model.ID.ValueString())
	if err != nil {
		diags.AddError(
			"Unable to Read Space",
			"An error occurred while reading the Confluence space.\n\nError: "+err.Error(),
		)
		return diags
	}

	applySpaceToModel(model, sp)
	return diags
}

// applySpaceToModel maps a spaceV2 onto the resource model.
func applySpaceToModel(model *SpaceResourceModel, sp *spaceV2) {
	model.ID = types.StringValue(sp.ID)
	model.Key = types.StringValue(sp.Key)
	model.Name = types.StringValue(sp.Name)
	model.Type = types.StringValue(sp.Type)
	model.Status = types.StringValue(sp.Status)

	if sp.HomepageID != "" {
		model.HomepageID = types.StringValue(sp.HomepageID)
	} else {
		model.HomepageID = types.StringNull()
	}

	if sp.Description.Plain.Value != "" {
		model.Description = types.StringValue(sp.Description.Plain.Value)
	} else if !model.Description.IsNull() {
		model.Description = types.StringNull()
	}

	model.URL = types.StringValue(spaceURL(sp))
}

// spaceURL derives the human-facing space URL from the v2 _links data.
func spaceURL(sp *spaceV2) string {
	if sp.Links.Base == "" {
		return ""
	}
	if sp.Links.WebUI != "" {
		return sp.Links.Base + sp.Links.WebUI
	}
	return sp.Links.Base + "/spaces/" + sp.Key
}
