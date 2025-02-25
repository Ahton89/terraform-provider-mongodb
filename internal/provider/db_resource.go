package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"terraform-provider-mongodb/internal/mongoclient/interfaces"
	"terraform-provider-mongodb/internal/mongoclient/types"
)

var (
	_ resource.Resource                = &resourceDatabase{}
	_ resource.ResourceWithConfigure   = &resourceDatabase{}
	_ resource.ResourceWithImportState = &resourceDatabase{}
)

func ResourceDatabase() resource.Resource {
	return &resourceDatabase{}
}

type resourceDatabase struct {
	client interfaces.Client
}

func (r *resourceDatabase) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *resourceDatabase) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Database name to create.",
			},
		},
	}
}

func (r *resourceDatabase) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := types.Database{}

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Provider().Resource().Database().Create(ctx, plan)
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

func (r *resourceDatabase) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := types.Database{}

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	exist, err := r.client.Provider().Resource().Database().Exists(ctx, state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to check database existence",
			err.Error(),
		)
		return
	}

	if !exist {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceDatabase) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning(
		"Update not supported",
		"The update method is not implemented for this resource, and any changes will require resource recreation.",
	)
}

func (r *resourceDatabase) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := types.Database{}

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Provider().Resource().Database().Delete(ctx, state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete database",
			err.Error(),
		)
		return
	}
}

func (r *resourceDatabase) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	database := req.ID

	state, err := r.client.Provider().Resource().Database().ImportState(ctx, database)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to import database state",
			err.Error(),
		)
		return
	}

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceDatabase) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(interfaces.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected interfaces.Client, got: %T. Please report this issue to the SRE team.", req.ProviderData),
		)

		return
	}

	r.client = client
}
