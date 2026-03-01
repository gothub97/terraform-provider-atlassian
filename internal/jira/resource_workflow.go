package jira

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
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
	_ resource.Resource                = &WorkflowResource{}
	_ resource.ResourceWithImportState = &WorkflowResource{}
)

// WorkflowResource implements the atlassian_jira_workflow resource.
type WorkflowResource struct {
	client *atlassian.Client
}

// WorkflowResourceModel describes the resource data model.
type WorkflowResourceModel struct {
	ID          types.String              `tfsdk:"id"`
	Name        types.String              `tfsdk:"name"`
	Description types.String              `tfsdk:"description"`
	Statuses    []WorkflowStatusModel     `tfsdk:"status"`
	Transitions []WorkflowTransitionModel `tfsdk:"transition"`
}

// WorkflowStatusModel describes a status entry within the workflow.
type WorkflowStatusModel struct {
	Name            types.String `tfsdk:"name"`
	StatusReference types.String `tfsdk:"status_reference"`
	StatusCategory  types.String `tfsdk:"status_category"`
	StatusID        types.String `tfsdk:"status_id"`
}

// WorkflowTransitionModel describes a transition entry within the workflow.
type WorkflowTransitionModel struct {
	Name                types.String                      `tfsdk:"name"`
	FromStatusReference types.String                      `tfsdk:"from_status_reference"`
	ToStatusReference   types.String                      `tfsdk:"to_status_reference"`
	Type                types.String                      `tfsdk:"type"`
	Validators          []WorkflowRuleModel               `tfsdk:"validator"`
	Condition           *WorkflowTransitionConditionModel `tfsdk:"condition"`
	PostFunctions       []WorkflowRuleModel               `tfsdk:"post_function"`
}

// WorkflowRuleModel describes a validator or post-function rule.
type WorkflowRuleModel struct {
	RuleKey    types.String `tfsdk:"rule_key"`
	Parameters types.Map    `tfsdk:"parameters"`
}

// WorkflowTransitionConditionModel describes the condition block on a transition.
type WorkflowTransitionConditionModel struct {
	Operator types.String        `tfsdk:"operator"`
	Rules    []WorkflowRuleModel `tfsdk:"rule"`
}

// --- API types for create/update ---

type workflowCreateRequest struct {
	Statuses  []workflowStatusDef `json:"statuses"`
	Workflows []workflowDef       `json:"workflows"`
	Scope     workflowScope       `json:"scope"`
}

type workflowScope struct {
	Type string `json:"type"`
}

type workflowStatusDef struct {
	ID              string `json:"id,omitempty"`
	Name            string `json:"name"`
	StatusReference string `json:"statusReference"`
	StatusCategory  string `json:"statusCategory"`
}

type workflowDef struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Version     *workflowVersion       `json:"version,omitempty"`
	Statuses    []workflowStatusLayout `json:"statuses"`
	Transitions []workflowTransDef     `json:"transitions"`
}

type workflowVersion struct {
	VersionNumber int    `json:"versionNumber"`
	ID            string `json:"id"`
}

type workflowStatusLayout struct {
	StatusReference string            `json:"statusReference"`
	Properties      map[string]string `json:"properties"`
}

type workflowTransDef struct {
	ID                string              `json:"id"`
	Name              string              `json:"name"`
	ToStatusReference string              `json:"toStatusReference"`
	Links             []workflowTransLink `json:"links"`
	Type              string              `json:"type"`
	Validators        []workflowRuleDef   `json:"validators,omitempty"`
	Conditions        *workflowCondDef    `json:"conditions,omitempty"`
	Actions           []workflowRuleDef   `json:"actions,omitempty"`
}

type workflowTransLink struct {
	FromStatusReference string `json:"fromStatusReference"`
}

type workflowRuleDef struct {
	RuleKey    string            `json:"ruleKey"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

type workflowCondDef struct {
	Operation       string            `json:"operation,omitempty"`
	Conditions      []workflowRuleDef `json:"conditions,omitempty"`
	ConditionGroups []workflowCondDef `json:"conditionGroups,omitempty"`
}

// --- API types for create/update response ---

type workflowCreateResponse struct {
	Workflows []workflowCreateEntry `json:"workflows"`
}

type workflowCreateEntry struct {
	ID       string                      `json:"id"`
	Name     string                      `json:"name"`
	Statuses []workflowCreateStatusEntry `json:"statuses"`
}

type workflowCreateStatusEntry struct {
	StatusReference string `json:"statusReference"`
	StatusID        string `json:"statusId"`
}

// --- API types for read ---

type workflowReadRequest struct {
	WorkflowIDs []string `json:"workflowIds"`
}

type workflowReadResponse struct {
	Statuses  []workflowReadGlobalStatus `json:"statuses"`
	Workflows []workflowReadEntry        `json:"workflows"`
}

type workflowReadGlobalStatus struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	StatusCategory  string `json:"statusCategory"`
	StatusReference string `json:"statusReference"`
}

type workflowReadEntry struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Statuses    []workflowReadStatusRef  `json:"statuses"`
	Transitions []workflowReadTransition `json:"transitions"`
	Version     *workflowVersion         `json:"version"`
}

type workflowReadStatusRef struct {
	StatusReference string            `json:"statusReference"`
	Properties      map[string]string `json:"properties"`
}

type workflowReadTransition struct {
	Name              string              `json:"name"`
	ToStatusReference string              `json:"toStatusReference"`
	Links             []workflowTransLink `json:"links"`
	Type              string              `json:"type"`
	Validators        []workflowRuleDef   `json:"validators"`
	Conditions        *workflowCondDef    `json:"conditions"`
	Actions           []workflowRuleDef   `json:"actions"`
}

// --- Update request ---

type workflowUpdateRequest struct {
	Statuses  []workflowStatusDef `json:"statuses"`
	Workflows []workflowDef       `json:"workflows"`
}

// NewWorkflowResource returns a new resource factory function.
func NewWorkflowResource() resource.Resource {
	return &WorkflowResource{}
}

func (r *WorkflowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_workflow"
}

func (r *WorkflowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira workflow.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The entity ID (UUID) of the workflow.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the workflow.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the workflow.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"status": schema.ListNestedBlock{
				Description: "The statuses used in this workflow. At least one status is required.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the status.",
							Required:    true,
						},
						"status_reference": schema.StringAttribute{
							Description: "A local reference string used to link statuses and transitions within the workflow.",
							Required:    true,
						},
						"status_category": schema.StringAttribute{
							Description: "The category of the status. Must be one of: TODO, IN_PROGRESS, DONE.",
							Required:    true,
						},
						"status_id": schema.StringAttribute{
							Description: "The Jira status ID, resolved after creation.",
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"transition": schema.ListNestedBlock{
				Description: "The transitions in this workflow. At least one transition is required.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the transition.",
							Required:    true,
						},
						"from_status_reference": schema.StringAttribute{
							Description: "The status reference the transition originates from. Omit for initial or global transitions.",
							Optional:    true,
						},
						"to_status_reference": schema.StringAttribute{
							Description: "The status reference the transition goes to.",
							Required:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of transition. Must be one of: initial, directed, global.",
							Required:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"validator": schema.ListNestedBlock{
							Description: "Validators for this transition.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"rule_key": schema.StringAttribute{
										Description: "The rule key for the validator.",
										Required:    true,
									},
									"parameters": schema.MapAttribute{
										Description: "The parameters for the validator rule.",
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
						"condition": schema.SingleNestedBlock{
							Description: "Condition for this transition.",
							Attributes: map[string]schema.Attribute{
								"operator": schema.StringAttribute{
									Description: "The logical operator. Must be one of: ALL, ANY.",
									Optional:    true,
								},
							},
							Blocks: map[string]schema.Block{
								"rule": schema.ListNestedBlock{
									Description: "Condition rules.",
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"rule_key": schema.StringAttribute{
												Description: "The rule key for the condition.",
												Required:    true,
											},
											"parameters": schema.MapAttribute{
												Description: "The parameters for the condition rule.",
												Optional:    true,
												ElementType: types.StringType,
											},
										},
									},
								},
							},
						},
						"post_function": schema.ListNestedBlock{
							Description: "Post-functions for this transition.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"rule_key": schema.StringAttribute{
										Description: "The rule key for the post-function.",
										Required:    true,
									},
									"parameters": schema.MapAttribute{
										Description: "The parameters for the post-function rule.",
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *WorkflowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WorkflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkflowResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate UUIDs for each user-provided status reference
	refToUUID := make(map[string]string)
	statusDefs := make([]workflowStatusDef, len(plan.Statuses))
	statusLayouts := make([]workflowStatusLayout, len(plan.Statuses))
	for i, s := range plan.Statuses {
		uuid := generateUUID()
		userRef := s.StatusReference.ValueString()
		refToUUID[userRef] = uuid
		statusDefs[i] = workflowStatusDef{
			Name:            s.Name.ValueString(),
			StatusReference: uuid,
			StatusCategory:  s.StatusCategory.ValueString(),
		}
		statusLayouts[i] = workflowStatusLayout{
			StatusReference: uuid,
			Properties:      map[string]string{},
		}
	}

	// Build transitions using UUID references
	transitions := r.buildTransitionDefs(ctx, plan.Transitions, refToUUID)

	createReq := workflowCreateRequest{
		Statuses: statusDefs,
		Workflows: []workflowDef{
			{
				Name:        plan.Name.ValueString(),
				Description: plan.Description.ValueString(),
				Statuses:    statusLayouts,
				Transitions: transitions,
			},
		},
		Scope: workflowScope{Type: "GLOBAL"},
	}

	var createResp workflowCreateResponse
	if err := r.client.Post(ctx, "/rest/api/3/workflows/create", createReq, &createResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Workflow",
			"An error occurred while calling the Jira API to create the workflow.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	if len(createResp.Workflows) == 0 {
		resp.Diagnostics.AddError(
			"Unexpected API Response",
			"The Jira API returned an empty workflows array when creating the workflow.",
		)
		return
	}

	wf := createResp.Workflows[0]
	plan.ID = types.StringValue(wf.ID)

	// Map status IDs from create response via UUID→userRef reverse map
	uuidToRef := make(map[string]string)
	for ref, uuid := range refToUUID {
		uuidToRef[uuid] = ref
	}
	statusIDMap := make(map[string]string) // userRef → statusID
	for _, s := range wf.Statuses {
		if userRef, ok := uuidToRef[s.StatusReference]; ok {
			statusIDMap[userRef] = s.StatusID
		}
	}
	for i := range plan.Statuses {
		ref := plan.Statuses[i].StatusReference.ValueString()
		if id, ok := statusIDMap[ref]; ok {
			plan.Statuses[i].StatusID = types.StringValue(id)
		} else {
			plan.Statuses[i].StatusID = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WorkflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkflowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readResp, diags := r.readWorkflow(ctx, state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readResp == nil || len(readResp.Workflows) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	mapWorkflowReadToState(&state, readResp, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WorkflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkflowResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state WorkflowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read current workflow to get version information
	currentResp, diags := r.readWorkflow(ctx, state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if currentResp == nil || len(currentResp.Workflows) == 0 {
		resp.Diagnostics.AddError(
			"Workflow Not Found",
			"The workflow could not be found for update.",
		)
		return
	}

	currentWF := currentResp.Workflows[0]

	// Build map of existing status name → statusReference UUID and statusID
	existingNameToRef := make(map[string]string)
	existingNameToID := make(map[string]string)
	wfStatusRefs := make(map[string]bool)
	for _, s := range currentWF.Statuses {
		wfStatusRefs[s.StatusReference] = true
	}
	for _, gs := range currentResp.Statuses {
		if wfStatusRefs[gs.StatusReference] {
			existingNameToRef[gs.Name] = gs.StatusReference
			existingNameToID[gs.Name] = gs.ID
		}
	}

	// Map plan statuses: reuse existing refs or generate new UUIDs for new statuses
	// The Jira update API requires the top-level statuses array to be non-empty.
	// For existing statuses, include their ID so the API recognizes them; use a fresh UUID as statusReference.
	// For new statuses, generate a UUID without an ID.
	refToAPIRef := make(map[string]string) // userRef → API statusReference (UUID)
	allStatusDefs := make([]workflowStatusDef, 0, len(plan.Statuses))
	statusLayouts := make([]workflowStatusLayout, len(plan.Statuses))
	for i, s := range plan.Statuses {
		userRef := s.StatusReference.ValueString()
		statusName := s.Name.ValueString()
		uuid := generateUUID()

		if existingID, ok := existingNameToID[statusName]; ok {
			// Existing status — include ID so API links to existing, UUID as reference
			allStatusDefs = append(allStatusDefs, workflowStatusDef{
				ID:              existingID,
				Name:            statusName,
				StatusReference: uuid,
				StatusCategory:  s.StatusCategory.ValueString(),
			})
		} else {
			// New status — no ID, API will create it
			allStatusDefs = append(allStatusDefs, workflowStatusDef{
				Name:            statusName,
				StatusReference: uuid,
				StatusCategory:  s.StatusCategory.ValueString(),
			})
		}
		refToAPIRef[userRef] = uuid
		statusLayouts[i] = workflowStatusLayout{
			StatusReference: uuid,
			Properties:      map[string]string{},
		}
	}

	// Build transitions using API references
	transitions := r.buildTransitionDefs(ctx, plan.Transitions, refToAPIRef)

	updateReq := workflowUpdateRequest{
		Statuses: allStatusDefs,
		Workflows: []workflowDef{
			{
				ID:          state.ID.ValueString(),
				Name:        plan.Name.ValueString(),
				Description: plan.Description.ValueString(),
				Version:     currentWF.Version,
				Statuses:    statusLayouts,
				Transitions: transitions,
			},
		},
	}

	var updateResp workflowCreateResponse
	if err := r.client.Post(ctx, "/rest/api/3/workflows/update", updateReq, &updateResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Workflow",
			"An error occurred while calling the Jira API to update the workflow.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = state.ID

	// Map status IDs: from update response for new statuses, from existing for reused
	apiRefToUserRef := make(map[string]string)
	for ref, apiRef := range refToAPIRef {
		apiRefToUserRef[apiRef] = ref
	}
	statusIDMap := make(map[string]string)
	// First, populate from existing statuses
	for i, s := range plan.Statuses {
		name := s.Name.ValueString()
		if id, ok := existingNameToID[name]; ok {
			statusIDMap[plan.Statuses[i].StatusReference.ValueString()] = id
		}
	}
	// Then, override with any new status IDs from update response
	if len(updateResp.Workflows) > 0 {
		for _, s := range updateResp.Workflows[0].Statuses {
			if userRef, ok := apiRefToUserRef[s.StatusReference]; ok {
				statusIDMap[userRef] = s.StatusID
			}
		}
	}
	for i := range plan.Statuses {
		ref := plan.Statuses[i].StatusReference.ValueString()
		if id, ok := statusIDMap[ref]; ok {
			plan.Statuses[i].StatusID = types.StringValue(id)
		} else {
			plan.Statuses[i].StatusID = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WorkflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkflowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/rest/api/3/workflow/%s", state.ID.ValueString()), nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			// Already deleted, nothing to do.
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Workflow",
			"An error occurred while calling the Jira API to delete the workflow.\n\n"+
				"Error: "+err.Error(),
		)
	}
}

func (r *WorkflowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readWorkflow reads a single workflow by entity ID. Returns nil if not found.
func (r *WorkflowResource) readWorkflow(ctx context.Context, entityID string) (*workflowReadResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	readReq := workflowReadRequest{
		WorkflowIDs: []string{entityID},
	}

	var readResp workflowReadResponse
	if err := r.client.Post(ctx, "/rest/api/3/workflows", readReq, &readResp); err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return nil, diags
		}
		diags.AddError(
			"Unable to Read Workflow",
			"An error occurred while calling the Jira API to read the workflow.\n\n"+
				"Error: "+err.Error(),
		)
		return nil, diags
	}

	if len(readResp.Workflows) == 0 {
		return nil, diags
	}

	return &readResp, diags
}

// mapWorkflowReadToState maps a workflow read response to the Terraform state model.
func mapWorkflowReadToState(state *WorkflowResourceModel, resp *workflowReadResponse, ctx context.Context, diags *diag.Diagnostics) {
	wf := resp.Workflows[0]

	state.ID = types.StringValue(wf.ID)
	state.Name = types.StringValue(wf.Name)
	if wf.Description != "" {
		state.Description = types.StringValue(wf.Description)
	} else {
		state.Description = types.StringNull()
	}

	// Build UUID→status metadata map from top-level statuses
	uuidToGlobalStatus := make(map[string]workflowReadGlobalStatus)
	for _, s := range resp.Statuses {
		uuidToGlobalStatus[s.StatusReference] = s
	}

	// Build name→userRef map from current state (for Read/Update; empty for Import)
	nameToUserRef := make(map[string]string)
	for _, s := range state.Statuses {
		nameToUserRef[s.Name.ValueString()] = s.StatusReference.ValueString()
	}

	// Collect UUIDs referenced by this workflow
	wfStatusUUIDs := make(map[string]bool)
	for _, s := range wf.Statuses {
		wfStatusUUIDs[s.StatusReference] = true
	}

	// Build UUID→userRef map for transition mapping
	uuidToUserRef := make(map[string]string)
	for uuid := range wfStatusUUIDs {
		gs, ok := uuidToGlobalStatus[uuid]
		if !ok {
			continue
		}
		if userRef, ok := nameToUserRef[gs.Name]; ok {
			uuidToUserRef[uuid] = userRef
		} else {
			// Import case: derive reference from status name
			uuidToUserRef[uuid] = strings.ToLower(strings.ReplaceAll(gs.Name, " ", "_"))
		}
	}

	// Build name→global status map for this workflow's statuses
	nameToGlobalStatus := make(map[string]workflowReadGlobalStatus)
	for _, gs := range resp.Statuses {
		if wfStatusUUIDs[gs.StatusReference] {
			nameToGlobalStatus[gs.Name] = gs
		}
	}

	// Map statuses preserving original state order (for Read/Update)
	if len(state.Statuses) > 0 {
		for i, existing := range state.Statuses {
			name := existing.Name.ValueString()
			if gs, ok := nameToGlobalStatus[name]; ok {
				state.Statuses[i].StatusCategory = types.StringValue(gs.StatusCategory)
				state.Statuses[i].StatusID = types.StringValue(gs.ID)
				delete(nameToGlobalStatus, name)
			}
		}
	} else {
		// Import case: build statuses from API response
		var newStatuses []WorkflowStatusModel
		for _, gs := range resp.Statuses {
			if !wfStatusUUIDs[gs.StatusReference] {
				continue
			}
			sm := WorkflowStatusModel{
				Name:            types.StringValue(gs.Name),
				StatusCategory:  types.StringValue(gs.StatusCategory),
				StatusID:        types.StringValue(gs.ID),
				StatusReference: types.StringValue(strings.ToLower(strings.ReplaceAll(gs.Name, " ", "_"))),
			}
			newStatuses = append(newStatuses, sm)
		}
		state.Statuses = newStatuses
	}

	// Build a helper to map a single API transition to a model
	mapTransition := func(t workflowReadTransition) WorkflowTransitionModel {
		tm := WorkflowTransitionModel{
			Name: types.StringValue(t.Name),
			Type: types.StringValue(strings.ToLower(t.Type)),
		}

		// Map toStatusReference UUID → user reference
		if userRef, ok := uuidToUserRef[t.ToStatusReference]; ok {
			tm.ToStatusReference = types.StringValue(userRef)
		} else {
			tm.ToStatusReference = types.StringValue(t.ToStatusReference)
		}

		// Map fromStatusReference from links
		if len(t.Links) > 0 && t.Links[0].FromStatusReference != "" {
			if userRef, ok := uuidToUserRef[t.Links[0].FromStatusReference]; ok {
				tm.FromStatusReference = types.StringValue(userRef)
			} else {
				tm.FromStatusReference = types.StringValue(t.Links[0].FromStatusReference)
			}
		} else {
			tm.FromStatusReference = types.StringNull()
		}

		// Map validators
		if len(t.Validators) > 0 {
			tm.Validators = make([]WorkflowRuleModel, len(t.Validators))
			for j, v := range t.Validators {
				tm.Validators[j] = mapRuleDefToModel(ctx, v)
			}
		}

		// Map condition
		if t.Conditions != nil && t.Conditions.Operation != "" {
			condModel := &WorkflowTransitionConditionModel{
				Operator: types.StringValue(t.Conditions.Operation),
			}
			if len(t.Conditions.Conditions) > 0 {
				condModel.Rules = make([]WorkflowRuleModel, len(t.Conditions.Conditions))
				for j, c := range t.Conditions.Conditions {
					condModel.Rules[j] = mapRuleDefToModel(ctx, c)
				}
			}
			tm.Condition = condModel
		}

		// Map actions → post-functions
		if len(t.Actions) > 0 {
			tm.PostFunctions = make([]WorkflowRuleModel, len(t.Actions))
			for j, a := range t.Actions {
				tm.PostFunctions[j] = mapRuleDefToModel(ctx, a)
			}
		}

		return tm
	}

	// Map transitions preserving original state order
	apiTransByName := make(map[string]workflowReadTransition)
	for _, t := range wf.Transitions {
		apiTransByName[t.Name] = t
	}

	if len(state.Transitions) > 0 {
		// Preserve config order: match by name
		for i, existing := range state.Transitions {
			name := existing.Name.ValueString()
			if t, ok := apiTransByName[name]; ok {
				state.Transitions[i] = mapTransition(t)
				delete(apiTransByName, name)
			}
		}
	} else {
		// Import case: use API order
		state.Transitions = make([]WorkflowTransitionModel, len(wf.Transitions))
		for i, t := range wf.Transitions {
			state.Transitions[i] = mapTransition(t)
		}
	}
}

// buildTransitionDefs converts the plan transition models to API transition definitions.
func (r *WorkflowResource) buildTransitionDefs(ctx context.Context, transitions []WorkflowTransitionModel, refToUUID map[string]string) []workflowTransDef {
	result := make([]workflowTransDef, len(transitions))
	for i, t := range transitions {
		toRef := t.ToStatusReference.ValueString()
		toUUID := refToUUID[toRef]

		td := workflowTransDef{
			ID:                fmt.Sprintf("%d", i+1),
			Name:              t.Name.ValueString(),
			ToStatusReference: toUUID,
			Links:             []workflowTransLink{},
			Type:              strings.ToUpper(t.Type.ValueString()),
		}

		if !t.FromStatusReference.IsNull() && !t.FromStatusReference.IsUnknown() {
			fromRef := t.FromStatusReference.ValueString()
			fromUUID := refToUUID[fromRef]
			td.Links = []workflowTransLink{
				{FromStatusReference: fromUUID},
			}
		}

		// Validators
		if len(t.Validators) > 0 {
			td.Validators = make([]workflowRuleDef, len(t.Validators))
			for j, v := range t.Validators {
				td.Validators[j] = mapModelToRuleDef(ctx, v)
			}
		}

		// Condition
		if t.Condition != nil && !t.Condition.Operator.IsNull() {
			cd := &workflowCondDef{
				Operation: strings.ToUpper(t.Condition.Operator.ValueString()),
			}
			if len(t.Condition.Rules) > 0 {
				cd.Conditions = make([]workflowRuleDef, len(t.Condition.Rules))
				for j, cr := range t.Condition.Rules {
					cd.Conditions[j] = mapModelToRuleDef(ctx, cr)
				}
			}
			td.Conditions = cd
		}

		// Post-functions → actions in API
		if len(t.PostFunctions) > 0 {
			td.Actions = make([]workflowRuleDef, len(t.PostFunctions))
			for j, pf := range t.PostFunctions {
				td.Actions[j] = mapModelToRuleDef(ctx, pf)
			}
		}

		result[i] = td
	}
	return result
}

// generateUUID generates a random UUID v4 string.
func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// mapModelToRuleDef converts a WorkflowRuleModel to a workflowRuleDef.
func mapModelToRuleDef(ctx context.Context, m WorkflowRuleModel) workflowRuleDef {
	rd := workflowRuleDef{
		RuleKey: m.RuleKey.ValueString(),
	}
	if !m.Parameters.IsNull() && !m.Parameters.IsUnknown() {
		params := make(map[string]string)
		m.Parameters.ElementsAs(ctx, &params, false)
		rd.Parameters = params
	}
	return rd
}

// mapRuleDefToModel converts a workflowRuleDef to a WorkflowRuleModel.
func mapRuleDefToModel(ctx context.Context, rd workflowRuleDef) WorkflowRuleModel {
	rm := WorkflowRuleModel{
		RuleKey: types.StringValue(rd.RuleKey),
	}
	if len(rd.Parameters) > 0 {
		mapVal, _ := types.MapValueFrom(ctx, types.StringType, rd.Parameters)
		rm.Parameters = mapVal
	} else {
		rm.Parameters = types.MapNull(types.StringType)
	}
	return rm
}
