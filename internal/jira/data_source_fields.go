package jira

import (
	"context"
	"fmt"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &FieldsDataSource{}

// FieldsDataSource implements the atlassian_jira_fields data source.
type FieldsDataSource struct {
	client *atlassian.Client
}

// FieldsDataSourceModel describes the data source data model.
type FieldsDataSourceModel struct {
	Type   types.String     `tfsdk:"type"`
	Fields []FieldDataModel `tfsdk:"fields"`
}

// FieldDataModel describes a single field in the fields list.
type FieldDataModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Custom      types.Bool     `tfsdk:"custom"`
	SchemaType  types.String   `tfsdk:"schema_type"`
	ClauseNames []types.String `tfsdk:"clause_names"`
}

// fieldListItem represents a single field from the GET /rest/api/3/field response.
type fieldListItem struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Custom      bool     `json:"custom"`
	ClauseNames []string `json:"clauseNames"`
	Schema      *struct {
		Type string `json:"type"`
	} `json:"schema"`
}

// NewFieldsDataSource returns a new data source factory function.
func NewFieldsDataSource() datasource.DataSource {
	return &FieldsDataSource{}
}

func (d *FieldsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_fields"
}

func (d *FieldsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve all Jira fields, optionally filtered by type.",
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Description: "Optional filter by field type. Must be \"system\" or \"custom\". If not set, all fields are returned.",
				Optional:    true,
			},
			"fields": schema.ListNestedAttribute{
				Description: "The list of fields.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the field.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the field.",
							Computed:    true,
						},
						"custom": schema.BoolAttribute{
							Description: "Whether the field is a custom field.",
							Computed:    true,
						},
						"schema_type": schema.StringAttribute{
							Description: "The schema type of the field.",
							Computed:    true,
						},
						"clause_names": schema.ListAttribute{
							Description: "The clause names for the field in JQL.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *FieldsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*atlassian.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *atlassian.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *FieldsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config FieldsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp []fieldListItem
	if err := d.client.Get(ctx, "/rest/api/3/field", &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Fields",
			"An error occurred while calling the Jira API to list fields.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Apply client-side filter based on the type attribute
	filterType := ""
	if !config.Type.IsNull() && !config.Type.IsUnknown() {
		filterType = config.Type.ValueString()
	}

	var filtered []fieldListItem
	for _, f := range apiResp {
		switch filterType {
		case "custom":
			if !f.Custom {
				continue
			}
		case "system":
			if f.Custom {
				continue
			}
		}
		filtered = append(filtered, f)
	}

	state := FieldsDataSourceModel{
		Type:   config.Type,
		Fields: make([]FieldDataModel, len(filtered)),
	}

	for i, f := range filtered {
		clauseNames := make([]types.String, len(f.ClauseNames))
		for j, cn := range f.ClauseNames {
			clauseNames[j] = types.StringValue(cn)
		}

		schemaType := types.StringNull()
		if f.Schema != nil {
			schemaType = types.StringValue(f.Schema.Type)
		}

		state.Fields[i] = FieldDataModel{
			ID:          types.StringValue(f.ID),
			Name:        types.StringValue(f.Name),
			Custom:      types.BoolValue(f.Custom),
			SchemaType:  schemaType,
			ClauseNames: clauseNames,
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
