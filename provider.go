package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
)

// Ensure the implementation satisfies the provider.Provider interface.
var _ provider.Provider = &OutpostProvider{}

// OutpostProvider defines the provider implementation.
type OutpostProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// New returns a new instance of the provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OutpostProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name.
func (p *OutpostProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "outpost"
	resp.Version = p.version
}

// Schema returns the provider schema.
func (p *OutpostProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Outpost provider provides utility functions for Terraform configurations, including YAML encoding with null value omission for Helm values files.",
	}
}

// Configure prepares the provider for data sources and resources.
func (p *OutpostProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// No configuration needed for this provider
}

// DataSources returns a slice of data source implementations.
func (p *OutpostProvider) DataSources(ctx context.Context) []func() provider.DataSource {
	return []func() provider.DataSource{}
}

// Resources returns a slice of resource implementations.
func (p *OutpostProvider) Resources(ctx context.Context) []func() provider.Resource {
	return []func() provider.Resource{}
}

// Functions returns a slice of function implementations.
func (p *OutpostProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewHelmValuesEncodeFunction,
	}
}

