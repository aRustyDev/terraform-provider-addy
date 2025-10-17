package data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/aRustyDev/terraform-provider-addy/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &apiTokenDetailsDataSource{}
	_ datasource.DataSourceWithConfigure = &apiTokenDetailsDataSource{}
)

func NewApiTokenDetailsDataSource() datasource.DataSource {
	return &apiTokenDetailsDataSource{}
}

type apiTokenDetailsDataSource struct {
	client  *http.Client
	apiKey  string
}

type apiTokenDetailsModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	CreatedAt types.String `tfsdk:"created_at"`
	ExpiresAt types.String `tfsdk:"expires_at"`
}

type apiTokenDetailsResponse struct {
	Name      string  `json:"name"`
	CreatedAt string  `json:"created_at"`
	ExpiresAt *string `json:"expires_at"`
}

func (d *apiTokenDetailsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_token_details"
}

func (d *apiTokenDetailsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches details about the current API token.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier attribute.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the API token.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The creation timestamp of the API token.",
				Computed:            true,
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "The expiration timestamp of the API token. Null if the token doesn't expire.",
				Computed:            true,
			},
		},
	}
}

func (d *apiTokenDetailsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*DataSourceData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *DataSourceData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = providerData.Client
	d.apiKey = providerData.ApiKey
}

func (d *apiTokenDetailsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state apiTokenDetailsModel

	tflog.Debug(ctx, "Reading API token details")

	body, err := utils.Curl(ctx, d.client, "api-token-details", "GET", d.apiKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read API Token Details",
			err.Error(),
		)
		return
	}

	var tokenDetails apiTokenDetailsResponse
	err = json.Unmarshal(body, &tokenDetails)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Parse API Token Details",
			err.Error(),
		)
		return
	}

	state.ID = types.StringValue("api-token-details")
	state.Name = types.StringValue(tokenDetails.Name)
	state.CreatedAt = types.StringValue(tokenDetails.CreatedAt)
	
	if tokenDetails.ExpiresAt != nil {
		state.ExpiresAt = types.StringValue(*tokenDetails.ExpiresAt)
	} else {
		state.ExpiresAt = types.StringNull()
	}

	tflog.Debug(ctx, "API token details read successfully", map[string]interface{}{
		"name":       tokenDetails.Name,
		"created_at": tokenDetails.CreatedAt,
		"expires_at": tokenDetails.ExpiresAt,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}