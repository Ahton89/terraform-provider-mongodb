package provider

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	_ datasource.DataSource              = &dbDataSource{}
	_ datasource.DataSourceWithConfigure = &dbDataSource{}
)

func NewDbDataSource() datasource.DataSource {
	return &dbDataSource{}
}

type dbDataSource struct {
	client *mongo.Client
}

type dbDataSourceModel struct {
	Databases []dbDataSourceDatabase `tfsdk:"databases"`
}

type dbDataSourceDatabase struct {
	Name tftypes.String `tfsdk:"name"`
}

func (d *dbDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databases"
}

func (d *dbDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"databases": schema.ListAttribute{
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name": types.StringType,
					},
				},
				Description: "List of databases with names",
			},
		},
	}
}

func (d *dbDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dbDataSourceModel

	list, err := d.client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list databases",
			err.Error(),
		)
		return
	}

	for _, database := range list {
		if !slices.Contains(defaultDatabases, database) {
			state.Databases = append(state.Databases, dbDataSourceDatabase{
				Name: types.StringValue(database),
			})
		}
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *dbDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mongo.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mongo.Client, got: %T. Please report this issue to the SRE team.", req.ProviderData),
		)

		return
	}

	d.client = client
}
