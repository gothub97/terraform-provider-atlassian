package jira

import (
	"context"
	"encoding/json"
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
	_ resource.Resource                = &NotificationSchemeResource{}
	_ resource.ResourceWithImportState = &NotificationSchemeResource{}
)

// NotificationSchemeResource implements the atlassian_jira_notification_scheme resource.
type NotificationSchemeResource struct {
	client *atlassian.Client
}

// NotificationSchemeResourceModel describes the resource data model.
type NotificationSchemeResourceModel struct {
	ID            types.String                   `tfsdk:"id"`
	Name          types.String                   `tfsdk:"name"`
	Description   types.String                   `tfsdk:"description"`
	Notifications []NotificationSchemeEventModel `tfsdk:"notification"`
}

// NotificationSchemeEventModel describes a single event notification.
type NotificationSchemeEventModel struct {
	ID               types.String `tfsdk:"id"`
	EventID          types.String `tfsdk:"event_id"`
	NotificationType types.String `tfsdk:"notification_type"`
	Parameter        types.String `tfsdk:"parameter"`
}

// --- API types ---

type notificationSchemeCreateRequest struct {
	Name                     string                             `json:"name"`
	Description              string                             `json:"description,omitempty"`
	NotificationSchemeEvents []notificationSchemeEventRequest   `json:"notificationSchemeEvents,omitempty"`
}

type notificationSchemeEventRequest struct {
	Event         notificationEventRef       `json:"event"`
	Notifications []notificationRecipientReq `json:"notifications"`
}

type notificationEventRef struct {
	ID string `json:"id"`
}

type notificationRecipientReq struct {
	NotificationType string `json:"notificationType"`
	Parameter        string `json:"parameter,omitempty"`
}

type notificationSchemeUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type notificationSchemeAPIResponse struct {
	ID                       json.Number                        `json:"id"`
	Name                     string                             `json:"name"`
	Description              string                             `json:"description"`
	NotificationSchemeEvents []notificationSchemeEventResponse  `json:"notificationSchemeEvents"`
}

type notificationSchemeEventResponse struct {
	Event         notificationEventResponse   `json:"event"`
	Notifications []notificationRecipientResp `json:"notifications"`
}

type notificationEventResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type notificationRecipientResp struct {
	ID               int    `json:"id"`
	NotificationType string `json:"notificationType"`
	Parameter        string `json:"parameter"`
}

type notificationAddPayload struct {
	NotificationSchemeEvents []notificationSchemeEventRequest `json:"notificationSchemeEvents"`
}

// NewNotificationSchemeResource returns a new resource factory function.
func NewNotificationSchemeResource() resource.Resource {
	return &NotificationSchemeResource{}
}

func (r *NotificationSchemeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_notification_scheme"
}

func (r *NotificationSchemeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira notification scheme.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the notification scheme.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the notification scheme.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the notification scheme.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"notification": schema.ListNestedBlock{
				Description: "Event notifications in the scheme.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the notification entry.",
							Computed:    true,
						},
						"event_id": schema.StringAttribute{
							Description: "The event ID (e.g. 1 for Issue Created).",
							Required:    true,
						},
						"notification_type": schema.StringAttribute{
							Description: "The notification recipient type (e.g. CurrentAssignee, Reporter, User, Group, ProjectRole, EmailAddress, AllWatchers, etc).",
							Required:    true,
						},
						"parameter": schema.StringAttribute{
							Description: "The parameter value (user account ID, group ID, role ID, email address, etc). Not needed for types like CurrentAssignee, Reporter, CurrentUser, AllWatchers.",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (r *NotificationSchemeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NotificationSchemeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NotificationSchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := notificationSchemeCreateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}
	apiReq.NotificationSchemeEvents = buildNotificationEvents(plan.Notifications)

	var createResp struct {
		ID json.Number `json:"id"`
	}
	if err := r.client.Post(ctx, "/rest/api/3/notificationscheme", apiReq, &createResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Notification Scheme",
			"An error occurred while calling the Jira API to create the notification scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(createResp.ID.String())

	// Read back to get notification IDs
	diags := r.readScheme(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NotificationSchemeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NotificationSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := r.readScheme(ctx, &state)
	if diags.HasError() {
		// Check for not found (state already handled by readScheme returning special diag)
		for _, d := range diags {
			if d.Summary() == notFoundSentinel {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NotificationSchemeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NotificationSchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state NotificationSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID
	schemeID := state.ID.ValueString()

	// Update name/description
	updateReq := notificationSchemeUpdateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}
	if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/notificationscheme/%s", schemeID), updateReq, nil); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Notification Scheme",
			"An error occurred while calling the Jira API to update the notification scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Read current notifications to get IDs for deletion
	var apiResp notificationSchemeAPIResponse
	if err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/notificationscheme/%s?expand=notificationSchemeEvents", schemeID), &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Notification Scheme",
			"An error occurred while reading current notifications.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Delete all existing notifications
	for _, event := range apiResp.NotificationSchemeEvents {
		for _, notif := range event.Notifications {
			err := r.client.Delete(ctx, fmt.Sprintf("/rest/api/3/notificationscheme/%s/notification/%d", schemeID, notif.ID), nil)
			if err != nil {
				if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
					continue
				}
				resp.Diagnostics.AddError(
					"Unable to Remove Notification",
					"An error occurred while removing an existing notification.\n\n"+
						"Error: "+err.Error(),
				)
				return
			}
		}
	}

	// Add new notifications from plan
	if len(plan.Notifications) > 0 {
		addPayload := notificationAddPayload{
			NotificationSchemeEvents: buildNotificationEvents(plan.Notifications),
		}
		if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/notificationscheme/%s/notification", schemeID), addPayload, nil); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Add Notifications",
				"An error occurred while adding notifications.\n\n"+
					"Error: "+err.Error(),
			)
			return
		}
	}

	// Read back to get notification IDs
	diags := r.readScheme(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NotificationSchemeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NotificationSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/rest/api/3/notificationscheme/%s", state.ID.ValueString()), nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Notification Scheme",
			"An error occurred while calling the Jira API to delete the notification scheme.\n\n"+
				"Error: "+err.Error(),
		)
	}
}

func (r *NotificationSchemeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

const notFoundSentinel = "__not_found__"

// readScheme reads the notification scheme from the API and maps it to the model.
// Returns a special diagnostic with Summary() == notFoundSentinel if the scheme was deleted out of band.
func (r *NotificationSchemeResource) readScheme(ctx context.Context, state *NotificationSchemeResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	var apiResp notificationSchemeAPIResponse
	err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/notificationscheme/%s?expand=notificationSchemeEvents", state.ID.ValueString()), &apiResp)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			diags.AddError(notFoundSentinel, "Resource not found")
			return diags
		}
		diags.AddError(
			"Unable to Read Notification Scheme",
			"An error occurred while calling the Jira API.\n\n"+
				"Error: "+err.Error(),
		)
		return diags
	}

	state.ID = types.StringValue(apiResp.ID.String())
	state.Name = types.StringValue(apiResp.Name)

	if apiResp.Description != "" {
		state.Description = types.StringValue(apiResp.Description)
	} else {
		state.Description = types.StringNull()
	}

	// Flatten event notifications into a flat list
	apiNotifs := []NotificationSchemeEventModel{}
	for _, event := range apiResp.NotificationSchemeEvents {
		for _, notif := range event.Notifications {
			entry := NotificationSchemeEventModel{
				ID:               types.StringValue(strconv.Itoa(notif.ID)),
				EventID:          types.StringValue(strconv.Itoa(event.Event.ID)),
				NotificationType: types.StringValue(notif.NotificationType),
			}
			if notif.Parameter != "" {
				entry.Parameter = types.StringValue(notif.Parameter)
			} else {
				entry.Parameter = types.StringNull()
			}
			apiNotifs = append(apiNotifs, entry)
		}
	}

	// Reorder API results to match the existing state order to avoid spurious diffs.
	if len(state.Notifications) > 0 && len(apiNotifs) > 0 {
		ordered := make([]NotificationSchemeEventModel, 0, len(apiNotifs))
		used := make([]bool, len(apiNotifs))

		// First, match entries from state in order
		for _, planned := range state.Notifications {
			for j, api := range apiNotifs {
				if !used[j] &&
					api.EventID.ValueString() == planned.EventID.ValueString() &&
					api.NotificationType.ValueString() == planned.NotificationType.ValueString() &&
					api.Parameter.ValueString() == planned.Parameter.ValueString() {
					ordered = append(ordered, api)
					used[j] = true
					break
				}
			}
		}
		// Append any remaining unmatched entries (new from API)
		for j, api := range apiNotifs {
			if !used[j] {
				ordered = append(ordered, api)
			}
		}
		state.Notifications = ordered
	} else {
		state.Notifications = apiNotifs
	}

	return diags
}

// buildNotificationEvents groups flat notification entries by event ID for the create API.
func buildNotificationEvents(notifications []NotificationSchemeEventModel) []notificationSchemeEventRequest {
	if len(notifications) == 0 {
		return nil
	}

	eventMap := make(map[string][]notificationRecipientReq)
	eventOrder := make([]string, 0)

	for _, n := range notifications {
		eventID := n.EventID.ValueString()
		if _, exists := eventMap[eventID]; !exists {
			eventOrder = append(eventOrder, eventID)
		}

		recipient := notificationRecipientReq{
			NotificationType: n.NotificationType.ValueString(),
		}
		if !n.Parameter.IsNull() && !n.Parameter.IsUnknown() {
			recipient.Parameter = n.Parameter.ValueString()
		}

		eventMap[eventID] = append(eventMap[eventID], recipient)
	}

	result := make([]notificationSchemeEventRequest, 0, len(eventMap))
	for _, eventID := range eventOrder {
		result = append(result, notificationSchemeEventRequest{
			Event:         notificationEventRef{ID: eventID},
			Notifications: eventMap[eventID],
		})
	}

	return result
}
