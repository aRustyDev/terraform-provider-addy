package data

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &appVersionDataSource{}
)

// NewappVersionDataSource is a helper function to simplify the provider implementation.
func NewAppVersionDataSource() datasource.DataSource {
	return &appVersionDataSource{}
}

// appVersionDataSource is the data source implementation.
type appVersionDataSource struct{}

// Metadata returns the data source type name.
func (d *appVersionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_appVersion"
}

// Schema defines the schema for the data source.
func (d *appVersionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

// Read refreshes the Terraform state with the latest data.
func (d *appVersionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	panic("not implemented")
}
