package provider

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-mongodb/internal/mongoclient"
	"terraform-provider-mongodb/internal/mongoclient/interfaces"
)

var (
	_ provider.Provider = &MongoDBProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MongoDBProvider{
			Version: version,
		}
	}
}

type MongoDBProvider struct {
	MongoDB interfaces.Client
	Version string
}

type mongodb struct {
	ConnectionString types.String `tfsdk:"connection_string"`
}

func (m *MongoDBProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mongodb"
	resp.Version = m.Version
}

func (m *MongoDBProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"connection_string": schema.StringAttribute{
				Required:    true,
				Description: "The connection string to the MongoDB.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						// This regex is a simplified version of the official MongoDB connection string regex.
						// It does not support connection string for MongoDb Atlas.
						regexp.MustCompile(`^mongodb://([^:@/]+):([^@/]+)@?([^/?]+)(?:/([^?]*))?(?:\?(.*))?$`),
						"Connection string must be a valid MongoDB connection string.",
					),
				},
			},
		},
	}
}

func (m *MongoDBProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config mongodb

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := mongoclient.New(ctx, config.ConnectionString.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create MongoDB client",
			err.Error(),
		)
		return
	}

	err = client.Ping(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to check MongoDB connection",
			err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client

	m.MongoDB = client
}

func (m *MongoDBProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		DataSourceDatabases,
		DataSourceUsers,
		DataSourceReplicaSet,
	}
}

func (m *MongoDBProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		ResourceDatabase,
		ResourceUser,
		ResourceReplicaSet,
	}
}
