package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"terraform-provider-mongodb/internal/mongoclient/interfaces"
)

var (
	_ datasource.DataSource              = &dataSourceReplicaSet{}
	_ datasource.DataSourceWithConfigure = &dataSourceReplicaSet{}
)

func DataSourceReplicaSet() datasource.DataSource {
	return &dataSourceReplicaSet{}
}

type dataSourceReplicaSet struct {
	client interfaces.Client
}

func (d *dataSourceReplicaSet) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_replicaset"
}

func (d *dataSourceReplicaSet) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Computed: true,
			},
			"members": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"host": schema.StringAttribute{
							Computed: true,
						},
						"arbiter_only": schema.BoolAttribute{
							Computed: true,
						},
						"build_indexes": schema.BoolAttribute{
							Computed: true,
						},
						"hidden": schema.BoolAttribute{
							Computed: true,
						},
						"priority": schema.Float64Attribute{
							Computed: true,
						},
						"secondary_delay_secs": schema.Int64Attribute{
							Computed: true,
						},
						"votes": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
			"protocol_version": schema.Int64Attribute{
				Computed: true,
			},
			"write_concern_majority_journal_default": schema.BoolAttribute{
				Computed: true,
			},
			"settings": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"chaining_allowed": schema.BoolAttribute{
						Computed: true,
					},
					"heartbeat_interval_millis": schema.Int64Attribute{
						Computed: true,
					},
					"heartbeat_timeout_secs": schema.Int64Attribute{
						Computed: true,
					},
					"election_timeout_millis": schema.Int64Attribute{
						Computed: true,
					},
					"catch_up_timeout_millis": schema.Int64Attribute{
						Computed: true,
					},
					"catch_up_takeover_delay_millis": schema.Int64Attribute{
						Computed: true,
					},
					"get_last_error_defaults": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"w": schema.Int64Attribute{
								Computed: true,
							},
							"wtimeout": schema.Int64Attribute{
								Computed: true,
							},
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceReplicaSet) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	state, err := d.client.Provider().DataSource().ReplicaSet().Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read replicaset",
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

func (d *dataSourceReplicaSet) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(interfaces.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected interfaces.Client, got: %T. Please report this issue to the SRE team.", req.ProviderData),
		)

		return
	}

	d.client = client
}
