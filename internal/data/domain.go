package data

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &domainDataSource{}
)

// NewDomainDataSource is a helper function to simplify the provider implementation.
func NewDomainDataSource() datasource.DataSource {
	return &domainDataSource{}
}

// domainDataSource is the data source implementation.
type domainDataSource struct{}

// Metadata returns the data source type name.
func (d *domainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

// Schema defines the schema for the data source.
func (d *domainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

// Read refreshes the Terraform state with the latest data.
func (d *domainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	panic("not implemented")
}
