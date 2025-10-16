package data

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &domainOptionsDataSource{}
)

// NewdomainOptionsDataSource is a helper function to simplify the provider implementation.
func NewDomainOptionsDataSource() datasource.DataSource {
	return &domainOptionsDataSource{}
}

// domainOptionsDataSource is the data source implementation.
type domainOptionsDataSource struct{}

// Metadata returns the data source type name.
func (d *domainOptionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domainOptions"
}

// Schema defines the schema for the data source.
func (d *domainOptionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

// Read refreshes the Terraform state with the latest data.
func (d *domainOptionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	panic("not implemented")
}
