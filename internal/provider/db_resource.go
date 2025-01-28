package provider

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	_ resource.Resource                = &dbResource{}
	_ resource.ResourceWithConfigure   = &dbResource{}
	_ resource.ResourceWithImportState = &dbResource{}
)

func NewDbResource() resource.Resource {
	return &dbResource{}
}

type dbResource struct {
	client *mongo.Client
}

type dbResourceModel struct {
	Name tftypes.String `tfsdk:"name"`
}

func (d *dbResourceModel) create(ctx context.Context, client *mongo.Client) error {
	database := d.Name.ValueString()

	if slices.Contains(defaultDatabases, database) {
		return fmt.Errorf("database %s is a default database and cannot be created", database)
	}

	exist, err := client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	if slices.Contains(exist, database) {
		return fmt.Errorf("database %s already exists", database)
	}

	db := client.Database(database)
	collection := db.Collection("created_by_terraform")
	document := bson.D{{"created_at", time.Now().Format(time.RFC850)}}

	_, err = collection.InsertOne(ctx, document)
	if err != nil {
		return err
	}

	return nil
}

func (d *dbResourceModel) delete(ctx context.Context, client *mongo.Client) error {
	database := d.Name.ValueString()

	if slices.Contains(defaultDatabases, database) {
		return fmt.Errorf("database %s is a default database and cannot be deleted", database)
	}

	exist, err := client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	if !slices.Contains(exist, database) {
		return fmt.Errorf("database %s does not exist", database)
	}
	
	return client.Database(database).Drop(ctx)
}

func (d *dbResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (d *dbResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The name of the database to create.",
			},
		},
	}
}

func (d *dbResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dbResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := plan.create(ctx, d.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create database",
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

func (d *dbResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dbResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	database := state.Name.ValueString()

	list, err := d.client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error checking database existence",
			err.Error(),
		)
		return
	}

	if !slices.Contains(list, database) {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *dbResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning(
		"Update Not Supported",
		"The update method is not implemented for this resource, and any changes will require resource recreation.",
	)
}

func (d *dbResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dbResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := state.delete(ctx, d.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete database",
			err.Error(),
		)
		return
	}
}

func (d *dbResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state dbResourceModel

	database := req.ID

	if slices.Contains(defaultDatabases, database) {
		resp.Diagnostics.AddError(
			"Failed to import database",
			fmt.Sprintf("database %s is a default database and cannot be imported", database),
		)
		return
	}

	list, err := d.client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error checking database existence",
			err.Error(),
		)
		return
	}

	if !slices.Contains(list, database) {
		resp.Diagnostics.AddError(
			"Failed to import database",
			fmt.Sprintf("database %s does not exist", database),
		)
		return
	}

	state.Name = tftypes.StringValue(database)

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *dbResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
