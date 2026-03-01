package jira

import (
	"context"
	"fmt"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &MyselfDataSource{}

// MyselfDataSource implements the atlassian_jira_myself data source.
type MyselfDataSource struct {
	client *atlassian.Client
}

// MyselfDataSourceModel describes the data source data model.
type MyselfDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	AccountID    types.String `tfsdk:"account_id"`
	AccountType  types.String `tfsdk:"account_type"`
	DisplayName  types.String `tfsdk:"display_name"`
	EmailAddress types.String `tfsdk:"email_address"`
	Active       types.Bool   `tfsdk:"active"`
	TimeZone     types.String `tfsdk:"time_zone"`
	Locale       types.String `tfsdk:"locale"`
	Self         types.String `tfsdk:"self"`
}

// myselfAPIResponse represents the JSON response from GET /rest/api/3/myself.
type myselfAPIResponse struct {
	AccountID    string `json:"accountId"`
	AccountType  string `json:"accountType"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	Active       bool   `json:"active"`
	TimeZone     string `json:"timeZone"`
	Locale       string `json:"locale"`
	Self         string `json:"self"`
}

// NewMyselfDataSource returns a new data source factory function.
func NewMyselfDataSource() datasource.DataSource {
	return &MyselfDataSource{}
}

func (d *MyselfDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jira_myself"
}

func (d *MyselfDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about the currently authenticated user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource, set to the account_id.",
				Computed:    true,
			},
			"account_id": schema.StringAttribute{
				Description: "The account ID of the authenticated user.",
				Computed:    true,
			},
			"account_type": schema.StringAttribute{
				Description: "The account type of the user (e.g., atlassian, app, customer).",
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
			"time_zone": schema.StringAttribute{
				Description: "The time zone of the user.",
				Computed:    true,
			},
			"locale": schema.StringAttribute{
				Description: "The locale of the user.",
				Computed:    true,
			},
			"self": schema.StringAttribute{
				Description: "The URL of the user's profile.",
				Computed:    true,
			},
		},
	}
}

func (d *MyselfDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MyselfDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var apiResp myselfAPIResponse
	if err := d.client.Get(ctx, "/rest/api/3/myself", &apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Myself",
			"An error occurred while calling the Jira API to read the current user.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	state := MyselfDataSourceModel{
		ID:           types.StringValue(apiResp.AccountID),
		AccountID:    types.StringValue(apiResp.AccountID),
		AccountType:  types.StringValue(apiResp.AccountType),
		DisplayName:  types.StringValue(apiResp.DisplayName),
		EmailAddress: types.StringValue(apiResp.EmailAddress),
		Active:       types.BoolValue(apiResp.Active),
		TimeZone:     types.StringValue(apiResp.TimeZone),
		Locale:       types.StringValue(apiResp.Locale),
		Self:         types.StringValue(apiResp.Self),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
