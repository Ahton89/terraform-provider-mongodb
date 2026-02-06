package provider

import (
	"context"
	"fmt"

	"terraform-provider-mongodb/internal/mongoclient/interfaces"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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

func (d *dataSourceReplicaSet) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves the current configuration of a MongoDB replica set, including member details and settings.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Replica set name",
				Computed:    true,
			},
			"members": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "Member ID",
							Computed:    true,
						},
						"host": schema.StringAttribute{
							Description: "Host and port (e.g., 'hostname:27017')",
							Computed:    true,
						},
						"arbiter_only": schema.BoolAttribute{
							Description: "Whether the replica set member is an arbiter only",
							Computed:    true,
						},
						"build_indexes": schema.BoolAttribute{
							Description: "Whether the replica set member should build indexes",
							Computed:    true,
						},
						"hidden": schema.BoolAttribute{
							Description: "Whether the replica set member is hidden from clients",
							Computed:    true,
						},
						"priority": schema.Float64Attribute{
							Description: "Election priority (0-1000)",
							Computed:    true,
						},
						"secondary_delay_secs": schema.Int64Attribute{
							Description: "Replication delay in seconds",
							Computed:    true,
						},
						"votes": schema.Int64Attribute{
							Description: "Number of votes (0 or 1)",
							Computed:    true,
						},
					},
				},
			},
			"protocol_version": schema.Int64Attribute{
				Description: "Replica set protocol version",
				Computed:    true,
			},
			"write_concern_majority_journal_default": schema.BoolAttribute{
				Description: "Whether to use majority write concern with journaling by default",
				Computed:    true,
			},
			"settings": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"chaining_allowed": schema.BoolAttribute{
						Description: "Whether to allow chaining of secondary replication",
						Computed:    true,
					},
					"heartbeat_interval_millis": schema.Int64Attribute{
						Description: "Frequency of heartbeats between members",
						Computed:    true,
					},
					"heartbeat_timeout_secs": schema.Int64Attribute{
						Description: "Timeout for heartbeat responses",
						Computed:    true,
					},
					"election_timeout_millis": schema.Int64Attribute{
						Description: "Timeout for elections",
						Computed:    true,
					},
					"catch_up_timeout_millis": schema.Int64Attribute{
						Description: "Timeout for catch-up operations (-1 for infinite)",
						Computed:    true,
					},
					"catch_up_takeover_delay_millis": schema.Int64Attribute{
						Description: "Delay before catch-up takeover",
						Computed:    true,
					},
					"get_last_error_defaults": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"w": schema.Int64Attribute{
								Description: "Write concern value",
								Computed:    true,
							},
							"wtimeout": schema.Int64Attribute{
								Description: "Write concern timeout",
								Computed:    true,
							},
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: false,
				Read:   false,
				Update: false,
				Delete: false,
			}),
		},
	}
}

func (d *dataSourceReplicaSet) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	state, err := d.client.DataSource().ReplicaSet().Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read replicaset", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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
