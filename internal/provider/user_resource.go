package provider

import (
	"context"
	"fmt"
	"reflect"

	"terraform-provider-mongodb/internal/mongoclient/interfaces"
	"terraform-provider-mongodb/internal/mongoclient/types"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var (
	_ resource.Resource                = &resourceUser{}
	_ resource.ResourceWithConfigure   = &resourceUser{}
	_ resource.ResourceWithImportState = &resourceUser{}
)

func ResourceUser() resource.Resource {
	return &resourceUser{}
}

type resourceUser struct {
	client interfaces.Client
}

func (r *resourceUser) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *resourceUser) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
						"role": schema.StringAttribute{
							Required:    true,
							Description: "The role to assign to the user.",
						},
						"database": schema.StringAttribute{
							Required:    true,
							Description: "The database to assign the role to.",
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceUser) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := types.User{}

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

	err := r.client.Resource().User().Create(apiCtx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create user", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceUser) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := types.User{}

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

	exist, err := r.client.Resource().User().Exists(apiCtx, state)
	if err != nil {
		resp.Diagnostics.AddError("Failed to check user existence", err.Error())
		return
	}

	if !exist {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceUser) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := types.User{}
	state := types.User{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.onlyTimeoutsChanged(plan, state) {
		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, defaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiCtx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	err := r.client.Resource().User().Update(apiCtx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update user", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceUser) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := types.User{}

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

	err := r.client.Resource().User().Delete(apiCtx, state)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete user", err.Error())
		return
	}
}

func (r *resourceUser) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	username := req.ID

	state, err := r.client.Resource().User().ImportState(ctx, username)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import user state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceUser) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceUser) onlyTimeoutsChanged(plan, state types.User) bool {
	cpPlan := plan
	cpState := state

	cpPlan.Timeouts = timeouts.Value{}
	cpState.Timeouts = timeouts.Value{}

	return reflect.DeepEqual(cpPlan, cpState)
}
