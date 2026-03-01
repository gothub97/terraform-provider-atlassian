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
	_ resource.Resource                = &SecuritySchemeResource{}
	_ resource.ResourceWithImportState = &SecuritySchemeResource{}
)

// SecuritySchemeResource implements the atlassian_jira_security_scheme resource.
type SecuritySchemeResource struct {
	client *atlassian.Client
}

// SecuritySchemeResourceModel describes the resource data model.
type SecuritySchemeResourceModel struct {
	ID          types.String               `tfsdk:"id"`
	Name        types.String               `tfsdk:"name"`
	Description types.String               `tfsdk:"description"`
	DefaultLevel types.String              `tfsdk:"default_level_id"`
	Levels      []SecuritySchemeLevelModel `tfsdk:"level"`
}

// SecuritySchemeLevelModel describes a security level within a scheme.
type SecuritySchemeLevelModel struct {
	ID          types.String                  `tfsdk:"id"`
	Name        types.String                  `tfsdk:"name"`
	Description types.String                  `tfsdk:"description"`
	IsDefault   types.Bool                    `tfsdk:"is_default"`
	Members     []SecuritySchemeMemberModel   `tfsdk:"member"`
}

// SecuritySchemeMemberModel describes a member of a security level.
type SecuritySchemeMemberModel struct {
	ID        types.String `tfsdk:"id"`
	Type      types.String `tfsdk:"type"`
	Parameter types.String `tfsdk:"parameter"`
}

// --- API types ---

type securitySchemeCreateRequest struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description,omitempty"`
	Levels      []securityLevelCreateReq  `json:"levels,omitempty"`
}

type securityLevelCreateReq struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description,omitempty"`
	IsDefault   bool                       `json:"isDefault,omitempty"`
	Members     []securityMemberCreateReq  `json:"members,omitempty"`
}

type securityMemberCreateReq struct {
	Type      string `json:"type"`
	Parameter string `json:"parameter,omitempty"`
}

type securitySchemeUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type securitySchemeAPIResponse struct {
	ID             int                      `json:"id"`
	Name           string                   `json:"name"`
	Description    string                   `json:"description"`
	DefaultLevelID int                      `json:"defaultSecurityLevelId"`
	Levels         []securityLevelResponse  `json:"levels"`
}

type securityLevelResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// securityLevelMemberPage is a single item from the paginated level/member endpoint.
type securityLevelMemberPage struct {
	ID                   string                       `json:"id"`
	IssueSecurityLevelID string                       `json:"issueSecurityLevelId"`
	IssueSecuritySchemeID string                      `json:"issueSecuritySchemeId"`
	Holder               securityMemberHolderResponse `json:"holder"`
}

type securityMemberHolderResponse struct {
	Type      string `json:"type"`
	Parameter string `json:"parameter"`
	Value     string `json:"value"`
}

// NewSecuritySchemeResource returns a new resource factory function.
func NewSecuritySchemeResource() resource.Resource {
	return &SecuritySchemeResource{}
}

func (r *SecuritySchemeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_security_scheme"
}

func (r *SecuritySchemeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Jira issue security scheme.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the security scheme.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the security scheme.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the security scheme.",
				Optional:    true,
			},
			"default_level_id": schema.StringAttribute{
				Description: "The ID of the default security level.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"level": schema.ListNestedBlock{
				Description: "Security levels in the scheme.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the security level.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the security level.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the security level.",
							Optional:    true,
							Computed:    true,
						},
						"is_default": schema.BoolAttribute{
							Description: "Whether this is the default security level.",
							Optional:    true,
							Computed:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"member": schema.ListNestedBlock{
							Description: "Members of the security level.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Description: "The ID of the member entry.",
										Computed:    true,
									},
									"type": schema.StringAttribute{
										Description: "The member type (reporter, group, user, projectrole, applicationRole).",
										Required:    true,
									},
									"parameter": schema.StringAttribute{
										Description: "The parameter value (group ID, user account ID, role ID, etc). Not needed for reporter type.",
										Optional:    true,
										Computed:    true,
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

func (r *SecuritySchemeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SecuritySchemeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SecuritySchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := securitySchemeCreateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	for _, level := range plan.Levels {
		levelReq := securityLevelCreateReq{
			Name:        level.Name.ValueString(),
			Description: level.Description.ValueString(),
			IsDefault:   level.IsDefault.ValueBool(),
		}
		for _, member := range level.Members {
			memberReq := securityMemberCreateReq{
				Type: member.Type.ValueString(),
			}
			if !member.Parameter.IsNull() && !member.Parameter.IsUnknown() {
				memberReq.Parameter = member.Parameter.ValueString()
			}
			levelReq.Members = append(levelReq.Members, memberReq)
		}
		apiReq.Levels = append(apiReq.Levels, levelReq)
	}

	var createResp struct {
		ID string `json:"id"`
	}
	if err := r.client.Post(ctx, "/rest/api/3/issuesecurityschemes", apiReq, &createResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Security Scheme",
			"An error occurred while calling the Jira API to create the security scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(createResp.ID)

	// Read back to populate level/member IDs
	diags := r.readScheme(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SecuritySchemeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SecuritySchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := r.readScheme(ctx, &state)
	if diags.HasError() {
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

func (r *SecuritySchemeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SecuritySchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state SecuritySchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID
	schemeID := state.ID.ValueString()

	// Update name/description
	updateReq := securitySchemeUpdateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}
	if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/issuesecurityschemes/%s", schemeID), updateReq, nil); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Security Scheme",
			"An error occurred while calling the Jira API to update the security scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Get existing levels from scheme endpoint
	var schemeResp securitySchemeAPIResponse
	if err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/issuesecurityschemes/%s", schemeID), &schemeResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Security Scheme",
			"An error occurred while reading the security scheme.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Delete existing levels (async)
	for _, level := range schemeResp.Levels {
		taskPath, err := r.client.DeleteWithRedirect(ctx, fmt.Sprintf("/rest/api/3/issuesecurityschemes/%s/level/%s", schemeID, level.ID))
		if err != nil {
			if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
				continue
			}
			resp.Diagnostics.AddError(
				"Unable to Delete Security Level",
				"An error occurred while deleting a security level.\n\n"+
					"Error: "+err.Error(),
			)
			return
		}
		if taskPath != "" {
			if err := r.client.WaitForTask(ctx, taskPath, 0); err != nil {
				resp.Diagnostics.AddError(
					"Security Level Deletion Failed",
					"The async deletion of a security level failed.\n\n"+
						"Error: "+err.Error(),
				)
				return
			}
		}
	}

	// Add new levels from plan
	if len(plan.Levels) > 0 {
		addLevelsReq := struct {
			Levels []securityLevelCreateReq `json:"levels"`
		}{}
		for _, level := range plan.Levels {
			levelReq := securityLevelCreateReq{
				Name:        level.Name.ValueString(),
				Description: level.Description.ValueString(),
				IsDefault:   level.IsDefault.ValueBool(),
			}
			for _, member := range level.Members {
				memberReq := securityMemberCreateReq{
					Type: member.Type.ValueString(),
				}
				if !member.Parameter.IsNull() && !member.Parameter.IsUnknown() {
					memberReq.Parameter = member.Parameter.ValueString()
				}
				levelReq.Members = append(levelReq.Members, memberReq)
			}
			addLevelsReq.Levels = append(addLevelsReq.Levels, levelReq)
		}

		if err := r.client.Put(ctx, fmt.Sprintf("/rest/api/3/issuesecurityschemes/%s/level", schemeID), addLevelsReq, nil); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Add Security Levels",
				"An error occurred while adding security levels.\n\n"+
					"Error: "+err.Error(),
			)
			return
		}
	}

	// Read back to populate level/member IDs
	diags := r.readScheme(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SecuritySchemeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SecuritySchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/rest/api/3/issuesecurityschemes/%s", state.ID.ValueString()), nil)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Security Scheme",
			"An error occurred while calling the Jira API to delete the security scheme.\n\n"+
				"Error: "+err.Error(),
		)
	}
}

func (r *SecuritySchemeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readScheme reads the security scheme and its levels/members from the API.
func (r *SecuritySchemeResource) readScheme(ctx context.Context, state *SecuritySchemeResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	schemeID := state.ID.ValueString()

	// Read scheme metadata
	var apiResp securitySchemeAPIResponse
	err := r.client.Get(ctx, fmt.Sprintf("/rest/api/3/issuesecurityschemes/%s", schemeID), &apiResp)
	if err != nil {
		if apiErr, ok := err.(*atlassian.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			diags.AddError(notFoundSentinel, "Resource not found")
			return diags
		}
		diags.AddError(
			"Unable to Read Security Scheme",
			"An error occurred while calling the Jira API.\n\n"+
				"Error: "+err.Error(),
		)
		return diags
	}

	state.Name = types.StringValue(apiResp.Name)
	if apiResp.Description != "" {
		state.Description = types.StringValue(apiResp.Description)
	} else {
		state.Description = types.StringNull()
	}
	if apiResp.DefaultLevelID > 0 {
		state.DefaultLevel = types.StringValue(strconv.Itoa(apiResp.DefaultLevelID))
	} else {
		state.DefaultLevel = types.StringNull()
	}

	defaultLevelIDStr := strconv.Itoa(apiResp.DefaultLevelID)

	// Read members from paginated endpoint
	members, err := atlassian.Paginate[securityLevelMemberPage](ctx, r.client,
		fmt.Sprintf("/rest/api/3/issuesecurityschemes/level/member?schemeId=%s", schemeID))
	if err != nil {
		diags.AddError(
			"Unable to Read Security Level Members",
			"An error occurred while reading security level members.\n\n"+
				"Error: "+err.Error(),
		)
		return diags
	}

	// Group members by level ID
	membersByLevel := make(map[string][]securityLevelMemberPage)
	for _, m := range members {
		membersByLevel[m.IssueSecurityLevelID] = append(membersByLevel[m.IssueSecurityLevelID], m)
	}

	// Build levels from scheme response
	state.Levels = make([]SecuritySchemeLevelModel, 0, len(apiResp.Levels))
	for _, level := range apiResp.Levels {
		levelModel := SecuritySchemeLevelModel{
			ID:        types.StringValue(level.ID),
			Name:      types.StringValue(level.Name),
			IsDefault: types.BoolValue(level.ID == defaultLevelIDStr),
		}
		if level.Description != "" {
			levelModel.Description = types.StringValue(level.Description)
		} else {
			levelModel.Description = types.StringNull()
		}

		levelMembers := membersByLevel[level.ID]
		levelModel.Members = make([]SecuritySchemeMemberModel, 0, len(levelMembers))
		for _, member := range levelMembers {
			memberModel := SecuritySchemeMemberModel{
				ID:   types.StringValue(member.ID),
				Type: types.StringValue(member.Holder.Type),
			}
			param := member.Holder.Parameter
			if param == "" {
				param = member.Holder.Value
			}
			if param != "" {
				memberModel.Parameter = types.StringValue(param)
			} else {
				memberModel.Parameter = types.StringNull()
			}
			levelModel.Members = append(levelModel.Members, memberModel)
		}

		state.Levels = append(state.Levels, levelModel)
	}

	return diags
}
