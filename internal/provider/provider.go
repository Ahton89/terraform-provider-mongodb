package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	_ provider.Provider = &MongodbProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MongodbProvider{
			Version: version,
		}
	}
}

type MongodbProvider struct {
	Client  *mongo.Client
	Version string
}

type mongodbProviderModel struct {
	ConnectionString types.String `tfsdk:"connection_string"`
}

func (m *MongodbProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mongodb"
	resp.Version = m.Version
}

func (m *MongodbProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
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

func (m *MongodbProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

	var config mongodbProviderModel

	diags := req.Config.Get(ctx, &config)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ConnectionString.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("connection_string"),
			"Unknown Connection String",
			"The provider requires a connection string to be set.",
		)
		return
	}

	// Create a new MongoDB client
	client, err := mongo.Connect(options.Client().ApplyURI(config.ConnectionString.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create MongoDB client",
			err.Error(),
		)
		return
	}

	// Check the connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect to MongoDB",
			err.Error(),
		)
		return
	}

	// Set the client on the response
	resp.DataSourceData = client
	resp.ResourceData = client

	// Set the client on the provider
	m.Client = client

	// Check the MongoDB version
	supported := supportedVersion(ctx, client)
	if !supported {
		resp.Diagnostics.AddError(
			"Unsupported MongoDB version",
			fmt.Sprintf("MongoDB version must be at least %s", minVersion),
		)
		return
	}
}

func (m *MongodbProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDbDataSource,
		NewUserDataSource,
	}
}

func (m *MongodbProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDbResource,
		NewUserResource,
	}
}
