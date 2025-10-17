package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	addydata "github.com/aRustyDev/terraform-provider-addy/internal/data"
	addyresource "github.com/aRustyDev/terraform-provider-addy/internal/resource"
	addyutils "github.com/aRustyDev/terraform-provider-addy/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &addyProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &addyProvider{
			version: version,
		}
	}
}

// addyProvider is the provider implementation.
type addyProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// hashicupsProviderModel maps provider schema data to a Go type.
type addyProviderModel struct {
	ApiKey types.String `tfsdk:"api_key"`
}

// Metadata returns the provider type name.
func (p *addyProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "addy"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *addyProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

// Configure prepares a Addy API client for data sources and resources.
func (p *addyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Addy client")
	// Retrieve provider data from configuration
	var config addyProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.ApiKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Addy.io API Key",
			"The provider cannot create the Addy API client as there is an unknown configuration value for the Addy API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ADDY_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	api_key := os.Getenv("ADDY_API_KEY")

	if !config.ApiKey.IsNull() {
		api_key = config.ApiKey.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if api_key == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("key"),
			"Missing Addy API Key",
			"The provider cannot create the Addy API client as there is a missing or empty value for the Addy API key. "+
				"Set the key value in the configuration or use the ADDY_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new HashiCups client using the configuration values
	client, err := addyutils.NewClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create HashiCups API Client",
			"An unexpected error occurred when creating the HashiCups API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"HashiCups Client Error: "+err.Error(),
		)
		return
	}

	// Make the HashiCups client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = &addydata.DataSourceData{
		Client: client,
		ApiKey: api_key,
	}
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *addyProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		addydata.NewDomainDataSource,
		addydata.NewDomainOptionsDataSource,
		addydata.NewAliasDataSource,
		addydata.NewAliasesDataSource,
		addydata.NewAppVersionDataSource,
		addydata.NewApiTokenDetailsDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *addyProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		addyresource.NewDomainResource,
	}
}
