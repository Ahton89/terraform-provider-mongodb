package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	tfypes "github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-mongodb/internal/mongoclient/interfaces"
)

var (
	_ datasource.DataSource              = &dataSourceDatabases{}
	_ datasource.DataSourceWithConfigure = &dataSourceDatabases{}
)

func DataSourceDatabases() datasource.DataSource {
	return &dataSourceDatabases{}
}

type dataSourceDatabases struct {
	client interfaces.Client
}

func (d *dataSourceDatabases) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databases"
}

func (d *dataSourceDatabases) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"databases": schema.ListAttribute{
				Computed: true,
				ElementType: tfypes.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name": tfypes.StringType,
					},
				},
				Description: "List of databases with names",
			},
		},
	}
}

func (d *dataSourceDatabases) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	state, err := d.client.DataSource().Database().Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read databases",
			err.Error(),
		)

		return
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *dataSourceDatabases) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(interfaces.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected interfaces.Client, got: %T.", req.ProviderData),
		)

		return
	}

	d.client = client
}
