package provider

import (
	"context"
	"fmt"

	"terraform-provider-mongodb/internal/mongoclient/interfaces"
	"terraform-provider-mongodb/internal/mongoclient/types"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

func (r *resourceDatabase) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Database name to create.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceDatabase) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := types.Database{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, defaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiCtx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	err := r.client.Resource().Database().Create(apiCtx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create database", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDatabase) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := types.Database{}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := state.Timeouts.Read(ctx, defaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiCtx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	exist, err := r.client.Resource().Database().Exists(apiCtx, state)
	if err != nil {
		resp.Diagnostics.AddError("Failed to check database existence", err.Error())
		return
	}

	if !exist {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceDatabase) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning(
		"Update not supported",
		"The update method is not implemented for this resource, and any changes will require resource recreation.",
	)
}

func (r *resourceDatabase) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := types.Database{}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, defaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiCtx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	err := r.client.Resource().Database().Delete(apiCtx, state)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete database", err.Error())
		return
	}
}

func (r *resourceDatabase) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	database := req.ID

	state, err := r.client.Resource().Database().ImportState(ctx, database)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import database", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceDatabase) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(interfaces.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected interfaces.Client, got: %T.", req.ProviderData),
		)

		return
	}

	r.client = client
}
