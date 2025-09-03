package provider

import (
	"context"
	"fmt"
	"reflect"
	"regexp"

	"terraform-provider-mongodb/internal/mongoclient/interfaces"
	"terraform-provider-mongodb/internal/mongoclient/types"
	"terraform-provider-mongodb/internal/provider/modifier"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ resource.Resource                = &resourceReplicaSet{}
	_ resource.ResourceWithConfigure   = &resourceReplicaSet{}
	_ resource.ResourceWithImportState = &resourceReplicaSet{}
)

func ResourceReplicaSet() resource.Resource {
	return &resourceReplicaSet{}
}

type resourceReplicaSet struct {
	client interfaces.Client
}

func (r *resourceReplicaSet) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_replicaset"
}

func (r *resourceReplicaSet) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "> **IMPORTANT: Updating members in a replica set can currently only add/change\n" +
			"> members one at a time. This functionality will be improved, but to avoid errors - add/change\n" +
			"> members in an existing replica set one at a time. This does not apply to the first creation of\n" +
			"> a replica set, when first created you can specify an arbitrary number of members**",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the replica set to create.",
				PlanModifiers: []planmodifier.String{
					modifier.ImmutableString(),
				},
			},
			"version": schema.Int64Attribute{
				Description: "The version of the replica set. Automatically incremented each time the configuration is changed.",
				Optional:    true,
			},
			"members": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Required:    true,
							Description: "The id of the replica set member.",
						},
						"host": schema.StringAttribute{
							Required:    true,
							Description: "The host of the replica set member.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^.+:\d+$`),
									"Host must be a valid mongodb host string, e.g localhost:27017",
								),
							},
						},
						"arbiter_only": schema.BoolAttribute{
							Optional:    true,
							Description: "Whether the replica set member is an arbiter only.",
						},
						"build_indexes": schema.BoolAttribute{
							Optional:    true,
							Description: "Whether the replica set member should build indexes.",
						},
						"hidden": schema.BoolAttribute{
							Optional:    true,
							Description: "Whether the replica set member is hidden.",
						},
						"priority": schema.Float64Attribute{
							Optional:    true,
							Description: "The priority of the replica set member.",
						},
						"secondary_delay_secs": schema.Int64Attribute{
							Optional:    true,
							Description: "The delay of the replica set member.",
						},
						"votes": schema.Int64Attribute{
							Optional:    true,
							Description: "The number of votes of the replica set member.",
						},
					},
				},
			},
			"protocol_version": schema.Int64Attribute{
				Description: "The protocol version of the replica set.",
				Optional:    true,
			},
			"write_concern_majority_journal_default": schema.BoolAttribute{
				Description: "Whether to use majority write concern with journaling by default.",
				Optional:    true,
			},
			"settings": schema.SingleNestedAttribute{
				Description: "The replica set settings.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"chaining_allowed": schema.BoolAttribute{
						Description: "Whether to allow chaining of secondary replication",
						Optional:    true,
					},
					"heartbeat_interval_millis": schema.Int64Attribute{
						Description: "Frequency of heartbeats between members",
						Optional:    true,
					},
					"heartbeat_timeout_secs": schema.Int64Attribute{
						Description: "Timeout for heartbeat responses",
						Optional:    true,
					},
					"election_timeout_millis": schema.Int64Attribute{
						Description: "Timeout for elections",
						Optional:    true,
					},
					"catch_up_timeout_millis": schema.Int64Attribute{
						Description: "Timeout for catch-up operations (-1 for infinite)",
						Optional:    true,
					},
					"catch_up_takeover_delay_millis": schema.Int64Attribute{
						Description: "Delay before catch-up takeover",
						Optional:    true,
					},
					"get_last_error_defaults": schema.SingleNestedAttribute{
						Description: "Default error handling settings",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"w": schema.Int64Attribute{
								Description: "Write concern value",
								Optional:    true,
							},
							"wtimeout": schema.Int64Attribute{
								Description: "Write concern timeout",
								Optional:    true,
							},
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

func (r *resourceReplicaSet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := types.ReplicaSet{}

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

	if err := r.client.Resource().ReplicaSet().Create(apiCtx, plan); err != nil {
		resp.Diagnostics.AddError("Failed to create replica set", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceReplicaSet) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := types.ReplicaSet{}

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

	exist, err := r.client.Resource().ReplicaSet().Exists(apiCtx, state)
	if err != nil {
		resp.Diagnostics.AddError("Failed to check replica set existence", err.Error())
		return
	}

	if !exist {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceReplicaSet) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := types.ReplicaSet{}
	state := types.ReplicaSet{}

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

	if err := r.client.Resource().ReplicaSet().Update(apiCtx, plan); err != nil {
		resp.Diagnostics.AddError("Failed to update replica set", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceReplicaSet) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"Delete Not Supported",
		"The delete method is not implemented for this resource, because it requires manual actions from the administrator.",
	)
}

func (r *resourceReplicaSet) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	state, err := r.client.Resource().ReplicaSet().ImportState(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import replica set", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceReplicaSet) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *resourceReplicaSet) onlyTimeoutsChanged(plan, state types.ReplicaSet) bool {
	cpPlan := plan
	cpState := state

	cpPlan.Timeouts = timeouts.Value{}
	cpState.Timeouts = timeouts.Value{}

	return reflect.DeepEqual(cpPlan, cpState)
}
