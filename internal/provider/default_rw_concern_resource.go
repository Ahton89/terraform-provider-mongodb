package provider

import (
	"context"
	"fmt"
	"reflect"

	"terraform-provider-mongodb/internal/mongoclient/interfaces"
	"terraform-provider-mongodb/internal/mongoclient/types"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                     = &resourceDefaultRWConcern{}
	_ resource.ResourceWithConfigure        = &resourceDefaultRWConcern{}
	_ resource.ResourceWithImportState      = &resourceDefaultRWConcern{}
	_ resource.ResourceWithConfigValidators = &resourceDefaultRWConcern{}
	_ resource.ResourceWithModifyPlan       = &resourceDefaultRWConcern{}
)

func ResourceDefaultRWConcern() resource.Resource {
	return &resourceDefaultRWConcern{}
}

type resourceDefaultRWConcern struct {
	client interfaces.Client
}

func (r *resourceDefaultRWConcern) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_default_rw_concern"
}

func (r *resourceDefaultRWConcern) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the cluster-wide default read and write concern using the `setDefaultRWConcern` command. " +
			"Only one instance of this resource should exist per replica set or sharded cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always `default`, the cluster-wide defaults are a singleton.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"default_read_concern": schema.SingleNestedAttribute{
				Optional: true,
				Description: "The cluster-wide default read concern. " +
					"Removing it from the configuration unsets the default read concern.",
				Attributes: map[string]schema.Attribute{
					"level": schema.StringAttribute{
						Required:    true,
						Description: "The read concern level. One of `local`, `available`, `majority`.",
						Validators: []validator.String{
							stringvalidator.OneOf(types.ReadConcernLevels...),
						},
					},
				},
			},
			"default_write_concern": schema.SingleNestedAttribute{
				Optional: true,
				Description: "The cluster-wide default write concern. " +
					"Once set, MongoDB does not allow to unset it, so it cannot be removed from the configuration.",
				Attributes: map[string]schema.Attribute{
					"w": schema.StringAttribute{
						Required: true,
						Description: "The number of members that must acknowledge the write, `majority` or a custom write concern name. " +
							"`0` is not allowed.",
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
							stringvalidator.NoneOf("0"),
						},
					},
					"journal": schema.BoolAttribute{
						Optional:    true,
						Description: "MongoDB `j` option. Whether the write must be written to the on-disk journal before acknowledgment.",
					},
					"wtimeout": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
						Description: "Time limit in milliseconds for the write concern. Defaults to `0` (no timeout).",
						Validators: []validator.Int64{
							int64validator.AtLeast(0),
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
			}),
		},
	}
}

func (r *resourceDefaultRWConcern) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("default_read_concern"),
			path.MatchRoot("default_write_concern"),
		),
	}
}

func (r *resourceDefaultRWConcern) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var stateWriteConcern, planWriteConcern fwtypes.Object

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("default_write_concern"), &stateWriteConcern)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("default_write_concern"), &planWriteConcern)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !stateWriteConcern.IsNull() && planWriteConcern.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("default_write_concern"),
			"Default write concern cannot be unset",
			"Starting in MongoDB 5.0, once the cluster-wide default write concern is set it cannot be unset. "+
				"Keep the default_write_concern block in the configuration and change its values instead.",
		)
	}
}

func (r *resourceDefaultRWConcern) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := types.DefaultRWConcern{}

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

	if err := r.client.Resource().DefaultRWConcern().Create(apiCtx, plan); err != nil {
		resp.Diagnostics.AddError("Failed to set default read/write concern", err.Error())
		return
	}

	plan.Id = fwtypes.StringValue(types.DefaultRWConcernID)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDefaultRWConcern) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := types.DefaultRWConcern{}

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

	actual, err := r.client.Resource().DefaultRWConcern().Read(apiCtx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read default read/write concern", err.Error())
		return
	}

	// Only the concerns present in the state are managed by this resource.
	if state.DefaultReadConcern != nil {
		state.DefaultReadConcern = actual.DefaultReadConcern
	}

	if state.DefaultWriteConcern != nil {
		state.DefaultWriteConcern = actual.DefaultWriteConcern
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceDefaultRWConcern) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := types.DefaultRWConcern{}
	state := types.DefaultRWConcern{}

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

	if err := r.client.Resource().DefaultRWConcern().Update(apiCtx, plan, state); err != nil {
		resp.Diagnostics.AddError("Failed to update default read/write concern", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDefaultRWConcern) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError(
		"Delete Not Supported",
		"The cluster-wide default read/write concern cannot be deleted, because starting in MongoDB 5.0 the default write concern cannot be unset. "+
			"To stop managing it with Terraform, remove the resource from the state with a `removed` block (Terraform 1.7+) "+
			"or by running `terraform state rm mongodb_default_rw_concern.<name>`. The values on the server remain unchanged.",
	)
}

func (r *resourceDefaultRWConcern) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != types.DefaultRWConcernID {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("The cluster-wide default read/write concern is a singleton, use %q as import ID.", types.DefaultRWConcernID),
		)
		return
	}

	state, err := r.client.Resource().DefaultRWConcern().ImportState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import default read/write concern", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDefaultRWConcern) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceDefaultRWConcern) onlyTimeoutsChanged(plan, state types.DefaultRWConcern) bool {
	cpPlan := plan
	cpState := state

	cpPlan.Timeouts = timeouts.Value{}
	cpState.Timeouts = timeouts.Value{}

	return reflect.DeepEqual(cpPlan, cpState)
}
