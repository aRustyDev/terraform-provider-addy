package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &domainResource{}
)

// NewdomainResource is a helper function to simplify the provider implementation.
func NewDomainResource() resource.Resource {
	return &domainResource{}
}

// domainResource is the resource implementation.
type domainResource struct{}

// Metadata returns the resource type name.
func (r *domainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

// Schema defines the schema for the resource.
func (r *domainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

// Create creates the resource and sets the initial Terraform state.
func (r *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	panic("not implemented")
}

// Read refreshes the Terraform state with the latest data.
func (r *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	panic("not implemented")
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *domainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	panic("not implemented")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	panic("not implemented")
}
