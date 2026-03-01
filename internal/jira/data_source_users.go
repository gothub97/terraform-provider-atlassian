package jira

import (
	"context"
	"fmt"
	"net/url"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UsersDataSource{}

// UsersDataSource implements the atlassian_jira_users data source.
type UsersDataSource struct {
	client *atlassian.Client
}

// UsersDataSourceModel describes the data source data model.
type UsersDataSourceModel struct {
	Query types.String      `tfsdk:"query"`
	Users []UserEntryModel  `tfsdk:"users"`
}

// UserEntryModel describes a single user in the users list.
type UserEntryModel struct {
	AccountID    types.String `tfsdk:"account_id"`
	DisplayName  types.String `tfsdk:"display_name"`
	EmailAddress types.String `tfsdk:"email_address"`
	Active       types.Bool   `tfsdk:"active"`
}

// userSearchResult represents a single user from GET /rest/api/3/user/search.
type userSearchResult struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	Active       bool   `json:"active"`
}

// NewUsersDataSource returns a new data source factory function.
func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

func (d *UsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_users"
}

func (d *UsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to search for Jira users by display name or email address.",
		Attributes: map[string]schema.Attribute{
			"query": schema.StringAttribute{
				Description: "Search string matching display name or email address.",
				Required:    true,
			},
			"users": schema.ListNestedAttribute{
				Description: "The list of matching users.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"account_id": schema.StringAttribute{
							Description: "The account ID of the user.",
							Computed:    true,
						},
						"display_name": schema.StringAttribute{
							Description: "The display name of the user.",
							Computed:    true,
						},
						"email_address": schema.StringAttribute{
							Description: "The email address of the user.",
							Computed:    true,
						},
						"active": schema.BoolAttribute{
							Description: "Whether the user account is active.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *UsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config UsersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("/rest/api/3/user/search?query=%s&maxResults=50", url.QueryEscape(config.Query.ValueString()))

	var apiResp []userSearchResult
	if err := d.client.Get(ctx, path, &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Users",
			"An error occurred while calling the Jira API to search users.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state := UsersDataSourceModel{
		Query: config.Query,
		Users: make([]UserEntryModel, len(apiResp)),
	}

	for i, u := range apiResp {
		state.Users[i] = UserEntryModel{
			AccountID:    types.StringValue(u.AccountID),
			DisplayName:  types.StringValue(u.DisplayName),
			EmailAddress: types.StringValue(u.EmailAddress),
			Active:       types.BoolValue(u.Active),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
