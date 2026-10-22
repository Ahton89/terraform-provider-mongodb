package provider

import (
	"context"
	"fmt"

	"terraform-provider-mongodb/internal/mongoclient/interfaces"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &dataSourceDefaultRWConcern{}
	_ datasource.DataSourceWithConfigure = &dataSourceDefaultRWConcern{}
)

func DataSourceDefaultRWConcern() datasource.DataSource {
	return &dataSourceDefaultRWConcern{}
}

type dataSourceDefaultRWConcern struct {
	client interfaces.Client
}

func (d *dataSourceDefaultRWConcern) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_default_rw_concern"
}

func (d *dataSourceDefaultRWConcern) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves the cluster-wide default read and write concern using the `getDefaultRWConcern` command.",
		Attributes: map[string]schema.Attribute{
			"default_read_concern": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The default read concern.",
				Attributes: map[string]schema.Attribute{
					"level": schema.StringAttribute{
						Computed:    true,
						Description: "The read concern level.",
					},
				},
			},
			"default_read_concern_source": schema.StringAttribute{
				Computed:    true,
				Description: "Where the default read concern comes from: `implicit` (server default) or `global` (set with setDefaultRWConcern).",
			},
			"default_write_concern": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The default write concern. `null` when it is implicit, MongoDB does not report the implicit value.",
				Attributes: map[string]schema.Attribute{
					"w": schema.StringAttribute{
						Computed:    true,
						Description: "The number of members that must acknowledge the write, `majority` or a custom write concern name.",
					},
					"journal": schema.BoolAttribute{
						Computed:    true,
						Description: "MongoDB `j` option. Whether the write must be written to the on-disk journal before acknowledgment.",
					},
					"wtimeout": schema.Int64Attribute{
						Computed:    true,
						Description: "Time limit in milliseconds for the write concern.",
					},
				},
			},
			"default_write_concern_source": schema.StringAttribute{
				Computed:    true,
				Description: "Where the default write concern comes from: `implicit` (server default) or `global` (set with setDefaultRWConcern).",
			},
		},
	}
}

func (d *dataSourceDefaultRWConcern) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	state, err := d.client.DataSource().DefaultRWConcern().Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read default read/write concern", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *dataSourceDefaultRWConcern) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
