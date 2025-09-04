package provider

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"terraform-provider-mongodb/internal/mongoclient"
	mongoclientTypes "terraform-provider-mongodb/internal/mongoclient/types"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_              provider.Provider = &mongoDBProvider{}
	defaultTimeout                   = 15 * time.Minute
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &mongoDBProvider{
			Version: version,
		}
	}
}

type mongoDBProvider struct {
	Version string
}

type mongodb struct {
	ConnectionString types.String `tfsdk:"connection_string"`
	RetryAttempts    types.Int32  `tfsdk:"retry_attempts"`
	RetryDelaySec    types.Int32  `tfsdk:"retry_delay_sec"`
	DefaultTimeout   types.Int32  `tfsdk:"default_timeout"`
}

func (m *mongoDBProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mongodb"
	resp.Version = m.Version
}

func (m *mongoDBProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: fmt.Sprintf(`
> **IMPORTANT:** This provider supports only MongoDB v%s  
> **DEFAULT TIMEOUT:** %0.f minutes for all resource operations (create, read, update, delete)  
> You can override the default timeout by setting the "timeouts" block in each resource.`,
			mongoclientTypes.MongoDBRequiredVersion,
			defaultTimeout.Minutes(),
		),
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
			"retry_attempts": schema.Int32Attribute{
				Optional:    true,
				Description: "The number of retry attempts for operations that fail due to transient errors.",
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
			},
			"retry_delay_sec": schema.Int32Attribute{
				Optional:    true,
				Description: "The delay in seconds between retry attempts.",
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
			},
			"default_timeout": schema.Int32Attribute{
				Optional: true,
				Description: "The default timeout in minutes for all resource operations (create, read, update, delete)." +
					" Current is 15 minutes.",
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
			},
		},
	}
}

func (m *mongoDBProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config mongodb

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.DefaultTimeout.ValueInt32() > 0 {
		defaultTimeout = time.Duration(config.DefaultTimeout.ValueInt32()) * time.Minute
	}

	client := mongoclient.New(
		config.ConnectionString.ValueString(),
		uint(config.RetryAttempts.ValueInt32()),
		uint(config.RetryDelaySec.ValueInt32()),
	)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (m *mongoDBProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		DataSourceDatabases,
		DataSourceUsers,
		DataSourceReplicaSet,
	}
}

func (m *mongoDBProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		ResourceDatabase,
		ResourceUser,
		ResourceReplicaSet,
	}
}
