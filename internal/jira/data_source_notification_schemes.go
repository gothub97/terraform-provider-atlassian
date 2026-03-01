package jira

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &NotificationSchemesDataSource{}

// NotificationSchemesDataSource implements the atlassian_jira_notification_schemes data source.
type NotificationSchemesDataSource struct {
	client *atlassian.Client
}

// NotificationSchemesDataSourceModel describes the data source data model.
type NotificationSchemesDataSourceModel struct {
	Schemes []NotificationSchemeEntryModel `tfsdk:"schemes"`
}

// NotificationSchemeEntryModel describes a single notification scheme in the list.
type NotificationSchemeEntryModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

// notificationSchemePageItem represents a single scheme in the paginated response.
type notificationSchemePageItem struct {
	ID          json.Number `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
}

// NewNotificationSchemesDataSource returns a new data source factory function.
func NewNotificationSchemesDataSource() datasource.DataSource {
	return &NotificationSchemesDataSource{}
}

func (d *NotificationSchemesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_notification_schemes"
}

func (d *NotificationSchemesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve all Jira notification schemes.",
		Attributes: map[string]schema.Attribute{
			"schemes": schema.ListNestedAttribute{
				Description: "The list of notification schemes.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the notification scheme.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the notification scheme.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the notification scheme.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *NotificationSchemesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NotificationSchemesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	items, err := atlassian.Paginate[notificationSchemePageItem](ctx, d.client, "/rest/api/3/notificationscheme")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Notification Schemes",
			"An error occurred while calling the Jira API to list notification schemes.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state := NotificationSchemesDataSourceModel{
		Schemes: make([]NotificationSchemeEntryModel, len(items)),
	}

	for i, s := range items {
		state.Schemes[i] = NotificationSchemeEntryModel{
			ID:          types.StringValue(s.ID.String()),
			Name:        types.StringValue(s.Name),
			Description: types.StringValue(s.Description),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
