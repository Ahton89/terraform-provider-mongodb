package provider

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

type userDataSource struct {
	client *mongo.Client
}

type userDataSourceModel struct {
	Users []userDataSourceUserModel `tfsdk:"users"`
}

type userDataSourceUserModel struct {
	Username tftypes.String            `tfsdk:"username"`
	Roles    []userDataSourceRoleModel `tfsdk:"roles"`
}

type userDataSourceRoleModel struct {
	Role     tftypes.String `tfsdk:"role"`
	Database tftypes.String `tfsdk:"database"`
}

func (u *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (u *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"username": schema.StringAttribute{
							Computed: true,
						},
						"roles": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"role": schema.StringAttribute{
										Computed: true,
									},
									"database": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (u *userDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state userDataSourceModel
	var r responseUsers

	err := u.client.Database("admin").RunCommand(ctx, bson.D{
		{Key: "usersInfo", Value: 1},
	}).Decode(&r)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read users",
			err.Error(),
		)
		return
	}

	if len(r.Users) > 0 {
		for _, user := range r.Users {

			// Skip default users
			if slices.Contains(defaultUsers, user.User) {
				continue
			}

			var roles []userDataSourceRoleModel

			for _, role := range user.Roles {
				roles = append(roles, userDataSourceRoleModel{
					Role:     tftypes.StringValue(role.Role),
					Database: tftypes.StringValue(role.Db),
				})
			}

			state.Users = append(state.Users, userDataSourceUserModel{
				Username: tftypes.StringValue(user.User),
				Roles:    roles,
			})
		}

	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (u *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	u.client = client
}
