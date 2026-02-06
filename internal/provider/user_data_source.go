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
	_ datasource.DataSource              = &dataSourceUsers{}
	_ datasource.DataSourceWithConfigure = &dataSourceUsers{}
)

func DataSourceUsers() datasource.DataSource {
	return &dataSourceUsers{}
}

type dataSourceUsers struct {
	client interfaces.Client
}

func (d *dataSourceUsers) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *dataSourceUsers) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a list of all MongoDB users with their roles and permissions, excluding system users.",
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"username": schema.StringAttribute{
							Computed:    true,
							Description: "The username of the user.",
						},
						"password": schema.StringAttribute{
							Computed:    true,
							Sensitive:   true,
							Description: "The password of the user.",
						},
						"roles": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"role": schema.StringAttribute{
										Computed:    true,
										Description: "The role of the user.",
									},
									"database": schema.StringAttribute{
										Computed:    true,
										Description: "The database of the user.",
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
				},
			},
		},
	}
}

func (d *dataSourceUsers) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	state, err := d.client.DataSource().User().Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read users", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *dataSourceUsers) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
