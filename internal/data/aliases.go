package data

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &aliasesDataSource{}
)

// NewaliasesDataSource is a helper function to simplify the provider implementation.
func NewAliasesDataSource() datasource.DataSource {
	return &aliasesDataSource{}
}

// aliasesDataSource is the data source implementation.
type aliasesDataSource struct{}

// Metadata returns the data source type name.
func (d *aliasesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aliases"
}

// Schema defines the schema for the data source.
func (d *aliasesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

// Read refreshes the Terraform state with the latest data.
func (d *aliasesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	panic("not implemented")
}
