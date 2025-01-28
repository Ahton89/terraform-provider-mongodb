package provider

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithConfigure   = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *mongo.Client
}

type userResourceModel struct {
	Username tftypes.String          `tfsdk:"username"`
	Password tftypes.String          `tfsdk:"password"`
	Roles    []userResourceRoleModel `tfsdk:"roles"`
}

func (u *userResourceModel) create(ctx context.Context, client *mongo.Client) error {
	username := u.Username.ValueString()

	// Check if the user is a default user
	if slices.Contains(defaultUsers, username) {
		return fmt.Errorf("user %s is a default user and cannot be created", username)
	}

	// Check if the user already exists
	var r responseUsers

	err := client.Database("admin").RunCommand(ctx, bson.D{
		{Key: "usersInfo", Value: 1},
	}).Decode(&r)
	if err != nil {
		return fmt.Errorf("failed to read users: %s", err)
	}

	if r.userExist(username) {
		return fmt.Errorf("user %s already exists", username)
	}

	// Create the user
	var roles bson.A
	for _, r := range u.Roles {
		roles = append(roles, bson.D{
			{Key: "role", Value: r.Role.ValueString()},
			{Key: "db", Value: r.Database.ValueString()},
		})
	}

	var command = bson.D{
		{Key: "createUser", Value: u.Username.ValueString()},
		{Key: "pwd", Value: u.Password.ValueString()},
		{Key: "roles", Value: roles},
	}

	return client.Database("admin").RunCommand(ctx, command).Err()
}

func (u *userResourceModel) delete(ctx context.Context, client *mongo.Client) error {
	var command = bson.D{
		{Key: "dropUser", Value: u.Username.ValueString()},
	}

	return client.Database("admin").RunCommand(ctx, command).Err()
}

func (u *userResourceModel) update(ctx context.Context, client *mongo.Client) error {
	var roles bson.A
	for _, r := range u.Roles {
		roles = append(roles, bson.D{
			{Key: "role", Value: r.Role.ValueString()},
			{Key: "db", Value: r.Database.ValueString()},
		})
	}

	var command = bson.D{
		{Key: "updateUser", Value: u.Username.ValueString()},
		{Key: "pwd", Value: u.Password.ValueString()},
		{Key: "roles", Value: roles},
	}

	return client.Database("admin").RunCommand(ctx, command).Err()
}

type userResourceRoleModel struct {
	Database tftypes.String `tfsdk:"database"`
	Role     tftypes.String `tfsdk:"role"`
}

func (u *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (u *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The name of the user to create.",
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The password of the user to create.",
			},
			"roles": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"database": schema.StringAttribute{
							Required:    true,
							Description: "The database to assign the role to.",
						},
						"role": schema.StringAttribute{
							Required:    true,
							Description: "The role to assign to the user.",
						},
					},
				},
			},
		},
	}
}

func (u *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := plan.create(ctx, u.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create user",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (u *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel
	var r responseUsers

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	if !r.userExist(state.Username.ValueString()) {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (u *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := plan.update(ctx, u.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update user",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (u *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := state.delete(ctx, u.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete user",
			err.Error(),
		)
		return
	}
}

func (u *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state userResourceModel
	var r responseUsers

	username := req.ID

	if slices.Contains(defaultUsers, username) {
		resp.Diagnostics.AddError(
			"Failed to import user",
			fmt.Sprintf("user %s is a default user and cannot be imported", username),
		)
		return
	}

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

	if !r.userExist(username) {
		resp.Diagnostics.AddError(
			"Failed to import user",
			fmt.Sprintf("user %s does not exist", username),
		)
		return
	}

	state.Username = tftypes.StringValue(username)

	user := r.getUser(username)

	for _, role := range user.Roles {
		state.Roles = append(state.Roles, userResourceRoleModel{
			Database: tftypes.StringValue(role.Db),
			Role:     tftypes.StringValue(role.Role),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (u *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
