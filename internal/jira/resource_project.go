package jira

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var (
	_ resource.Resource                = &ProjectResource{}
	_ resource.ResourceWithImportState = &ProjectResource{}

	projectKeyRegex = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,9}$`)

	// defaultTemplateKeys maps project_type_key to its default template.
	defaultTemplateKeys = map[string]string{
		"software":     "com.pyxis.greenhopper.jira:gh-simplified-agility-kanban",
		"business":     "com.atlassian.jira-core-project-templates:jira-core-simplified-task-tracking",
		"service_desk": "com.atlassian.servicedesk:simplified-it-service-management",
	}

	validProjectTypeKeys = map[string]bool{
		"software":     true,
		"business":     true,
		"service_desk": true,
	}

	validAssigneeTypes = map[string]bool{
		"PROJECT_LEAD": true,
		"UNASSIGNED":   true,
	}
)

// ProjectResource implements the atlassian_jira_project resource.
type ProjectResource struct {
	client *atlassian.Client
}

// ProjectResourceModel describes the resource data model.
type ProjectResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Key                types.String `tfsdk:"key"`
	Name               types.String `tfsdk:"name"`
	ProjectTypeKey     types.String `tfsdk:"project_type_key"`
	ProjectTemplateKey types.String `tfsdk:"project_template_key"`
	LeadAccountID      types.String `tfsdk:"lead_account_id"`
	Description        types.String `tfsdk:"description"`
	AssigneeType       types.String `tfsdk:"assignee_type"`
	Self               types.String `tfsdk:"self"`
}

// projectCreateRequest is the JSON body for POST /rest/api/3/project.
type projectCreateRequest struct {
	Key                string `json:"key"`
	Name               string `json:"name"`
	ProjectTypeKey     string `json:"projectTypeKey"`
	ProjectTemplateKey string `json:"projectTemplateKey,omitempty"`
	LeadAccountID      string `json:"leadAccountId"`
	Description        string `json:"description,omitempty"`
	AssigneeType       string `json:"assigneeType,omitempty"`
}

// projectUpdateRequest is the JSON body for PUT /rest/api/3/project/{id}.
type projectUpdateRequest struct {
	Name          string `json:"name,omitempty"`
	LeadAccountID string `json:"leadAccountId,omitempty"`
	Description   string `json:"description,omitempty"`
	AssigneeType  string `json:"assigneeType,omitempty"`
}

// projectAPIResponse represents the JSON response from the project API.
type projectAPIResponse struct {
	ID             string `json:"id"`
	Key            string `json:"key"`
	Name           string `json:"name"`
	ProjectTypeKey string `json:"projectTypeKey"`
	Description    string `json:"description"`
	AssigneeType   string `json:"assigneeType"`
	Self           string `json:"self"`
	Lead           struct {
		AccountID string `json:"accountId"`
	} `json:"lead"`
}

// projectCreateResponse represents the JSON response from POST /rest/api/3/project.
type projectCreateResponse struct {
	ID   int    `json:"id"`
	Key  string `json:"key"`
	Self string `json:"self"`
}

// NewProjectResource returns a new resource factory function.
func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

func (r *ProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_project"
}

func (r *ProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The numeric ID of the project.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "The project key. Must match ^[A-Z][A-Z0-9]{1,9}$.",
				Required:    true,
				Validators: []validator.String{
					projectKeyValidator{},
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the project.",
				Required:    true,
			},
			"project_type_key": schema.StringAttribute{
				Description: "The type of the project. Must be one of: software, business, service_desk.",
				Required:    true,
				Validators: []validator.String{
					projectTypeKeyValidator{},
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_template_key": schema.StringAttribute{
				Description: "The template to use for the project. If not set, a default template is chosen based on project_type_key.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"lead_account_id": schema.StringAttribute{
				Description: "The account ID of the project lead.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the project.",
				Optional:    true,
			},
			"assignee_type": schema.StringAttribute{
				Description: "The default assignee type. Must be one of: PROJECT_LEAD, UNASSIGNED.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					assigneeTypeValidator{},
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"self": schema.StringAttribute{
				Description: "The URL of the project.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ProjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine project template key
	templateKey := plan.ProjectTemplateKey.ValueString()
	if plan.ProjectTemplateKey.IsNull() || plan.ProjectTemplateKey.IsUnknown() {
		templateKey = DeriveTemplateKey(plan.ProjectTypeKey.ValueString())
	}

	createReq := projectCreateRequest{
		Key:                plan.Key.ValueString(),
		Name:               plan.Name.ValueString(),
		ProjectTypeKey:     plan.ProjectTypeKey.ValueString(),
		ProjectTemplateKey: templateKey,
		LeadAccountID:      plan.LeadAccountID.ValueString(),
		Description:        plan.Description.ValueString(),
		AssigneeType:       plan.AssigneeType.ValueString(),
	}

	var createResp projectCreateResponse
	if err := r.client.Post(ctx, "/rest/api/3/project", createReq, &createResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Project",
			"An error occurred while calling the Jira API to create the project.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Read back the project to get all computed fields
	var apiResp projectAPIResponse
	if err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/project/%s", createResp.Key), &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Created Project",
			"The project was created but an error occurred while reading it back.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	mapProjectAPIToState(&plan, &apiResp, templateKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp projectAPIResponse
	err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/project/%s", state.ID.ValueString()), &apiResp)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Project",
			"An error occurred while calling the Jira API to read the project.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Preserve the project_template_key from state since the API does not return it.
	existingTemplateKey := state.ProjectTemplateKey.ValueString()
	mapProjectAPIToState(&state, &apiResp, existingTemplateKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := projectUpdateRequest{
		Name:          plan.Name.ValueString(),
		LeadAccountID: plan.LeadAccountID.ValueString(),
		Description:   plan.Description.ValueString(),
		AssigneeType:  plan.AssigneeType.ValueString(),
	}

	var updateResp projectAPIResponse
	if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/project/%s", state.ID.ValueString()), updateReq, &updateResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Project",
			"An error occurred while calling the Jira API to update the project.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Read back to get the latest state
	var apiResp projectAPIResponse
	if err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/project/%s", state.ID.ValueString()), &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Updated Project",
			"The project was updated but an error occurred while reading it back.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	existingTemplateKey := state.ProjectTemplateKey.ValueString()
	mapProjectAPIToState(&plan, &apiResp, existingTemplateKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/rest/api/3/project/%s?enableUndo=true", state.ID.ValueString()), nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			// Already deleted, nothing to do.
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Project",
			"An error occurred while calling the Jira API to delete the project.\n\n"+
				"Error: "+err.Error(),
		)
	}
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by key or numeric ID - read the project to get all fields
	var apiResp projectAPIResponse
	err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/project/%s", req.ID), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Import Project",
			"An error occurred while calling the Jira API to read the project for import.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Derive a template key based on the project type since the API does not return it
	templateKey := DeriveTemplateKey(apiResp.ProjectTypeKey)

	var state ProjectResourceModel
	mapProjectAPIToState(&state, &apiResp, templateKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// DeriveTemplateKey returns the default project template key for a given project type.
func DeriveTemplateKey(projectTypeKey string) string {
	if tmpl, ok := defaultTemplateKeys[projectTypeKey]; ok {
		return tmpl
	}
	return ""
}

// mapProjectAPIToState maps a project API response to the Terraform state model.
func mapProjectAPIToState(state *ProjectResourceModel, apiResp *projectAPIResponse, templateKey string) {
	state.ID = types.StringValue(apiResp.ID)
	state.Key = types.StringValue(apiResp.Key)
	state.Name = types.StringValue(apiResp.Name)
	state.ProjectTypeKey = types.StringValue(apiResp.ProjectTypeKey)
	state.LeadAccountID = types.StringValue(apiResp.Lead.AccountID)
	if apiResp.Description != "" {
		state.Description = types.StringValue(apiResp.Description)
	} else {
		state.Description = types.StringNull()
	}
	state.AssigneeType = types.StringValue(apiResp.AssigneeType)
	state.Self = types.StringValue(apiResp.Self)

	if templateKey != "" {
		state.ProjectTemplateKey = types.StringValue(templateKey)
	} else {
		state.ProjectTemplateKey = types.StringNull()
	}
}

// --- Validators ---

// projectKeyValidator validates the project key format.
type projectKeyValidator struct{}

func (v projectKeyValidator) Description(_ context.Context) string {
	return "project key must match ^[A-Z][A-Z0-9]{1,9}$"
}

func (v projectKeyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v projectKeyValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if !projectKeyRegex.MatchString(value) {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Invalid Project Key",
			fmt.Sprintf("The project key %q must start with an uppercase letter and contain only uppercase letters and digits (2-10 characters total). Pattern: ^[A-Z][A-Z0-9]{1,9}$", value),
		))
	}
}

// projectTypeKeyValidator validates the project type key.
type projectTypeKeyValidator struct{}

func (v projectTypeKeyValidator) Description(_ context.Context) string {
	return "project type key must be one of: software, business, service_desk"
}

func (v projectTypeKeyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v projectTypeKeyValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if !validProjectTypeKeys[value] {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Invalid Project Type Key",
			fmt.Sprintf("The project type key %q must be one of: software, business, service_desk.", value),
		))
	}
}

// assigneeTypeValidator validates the assignee type.
type assigneeTypeValidator struct{}

func (v assigneeTypeValidator) Description(_ context.Context) string {
	return "assignee type must be one of: PROJECT_LEAD, UNASSIGNED"
}

func (v assigneeTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v assigneeTypeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if !validAssigneeTypes[value] {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Invalid Assignee Type",
			fmt.Sprintf("The assignee type %q must be one of: PROJECT_LEAD, UNASSIGNED.", value),
		))
	}
}
