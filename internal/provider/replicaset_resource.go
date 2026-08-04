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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &resourceReplicaSet{}
	_ resource.ResourceWithConfigure   = &resourceReplicaSet{}
	_ resource.ResourceWithImportState = &resourceReplicaSet{}
)

// MongoDB fills in every omitted member and settings field with its own
// default, so the provider has to declare the same defaults. Otherwise an
// attribute left out of the configuration is planned as null while the state
// read back from the server holds the server side value, which produces a
// permanent "1 -> null" diff.
var (
	getLastErrorDefaultsAttrTypes = map[string]attr.Type{
		"w":        fwtypes.Int64Type,
		"wtimeout": fwtypes.Int64Type,
	}

	settingsAttrTypes = map[string]attr.Type{
		"chaining_allowed":               fwtypes.BoolType,
		"heartbeat_interval_millis":      fwtypes.Int64Type,
		"heartbeat_timeout_secs":         fwtypes.Int64Type,
		"election_timeout_millis":        fwtypes.Int64Type,
		"catch_up_timeout_millis":        fwtypes.Int64Type,
		"catch_up_takeover_delay_millis": fwtypes.Int64Type,
		"get_last_error_defaults":        fwtypes.ObjectType{AttrTypes: getLastErrorDefaultsAttrTypes},
	}

	defaultGetLastErrorDefaults = fwtypes.ObjectValueMust(getLastErrorDefaultsAttrTypes, map[string]attr.Value{
		"w":        fwtypes.Int64Value(1),
		"wtimeout": fwtypes.Int64Value(0),
	})

	defaultSettings = fwtypes.ObjectValueMust(settingsAttrTypes, map[string]attr.Value{
		"chaining_allowed":               fwtypes.BoolValue(true),
		"heartbeat_interval_millis":      fwtypes.Int64Value(2000),
		"heartbeat_timeout_secs":         fwtypes.Int64Value(10),
		"election_timeout_millis":        fwtypes.Int64Value(10000),
		"catch_up_timeout_millis":        fwtypes.Int64Value(-1),
		"catch_up_takeover_delay_millis": fwtypes.Int64Value(30000),
		"get_last_error_defaults":        defaultGetLastErrorDefaults,
	})
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
							Computed:    true,
							Default:     booldefault.StaticBool(false),
							Description: "Whether the replica set member is an arbiter only. Defaults to `false`.",
						},
						"build_indexes": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(true),
							Description: "Whether the replica set member should build indexes. Defaults to `true`.",
						},
						"hidden": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
							Description: "Whether the replica set member is hidden. Defaults to `false`.",
						},
						"priority": schema.Float64Attribute{
							Optional: true,
							Computed: true,
							Default:  float64default.StaticFloat64(1),
							Description: "The priority of the replica set member. Defaults to `1`. " +
								"Members that are hidden, delayed or arbiters must be given a priority of `0` explicitly.",
						},
						"secondary_delay_secs": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Default:     int64default.StaticInt64(0),
							Description: "The delay of the replica set member. Defaults to `0`.",
						},
						"votes": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Default:     int64default.StaticInt64(1),
							Description: "The number of votes of the replica set member. Defaults to `1`.",
						},
					},
				},
			},
			"protocol_version": schema.Int64Attribute{
				Description: "The protocol version of the replica set. Defaults to `1`.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
			},
			"write_concern_majority_journal_default": schema.BoolAttribute{
				Description: "Whether to use majority write concern with journaling by default. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"settings": schema.SingleNestedAttribute{
				Description: "The replica set settings.",
				Optional:    true,
				Computed:    true,
				Default:     objectdefault.StaticValue(defaultSettings),
				Attributes: map[string]schema.Attribute{
					"chaining_allowed": schema.BoolAttribute{
						Description: "Whether to allow chaining of secondary replication. Defaults to `true`.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
					},
					"heartbeat_interval_millis": schema.Int64Attribute{
						Description: "Frequency of heartbeats between members. Defaults to `2000`.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(2000),
					},
					"heartbeat_timeout_secs": schema.Int64Attribute{
						Description: "Timeout for heartbeat responses. Defaults to `10`.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(10),
					},
					"election_timeout_millis": schema.Int64Attribute{
						Description: "Timeout for elections. Defaults to `10000`.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(10000),
					},
					"catch_up_timeout_millis": schema.Int64Attribute{
						Description: "Timeout for catch-up operations (-1 for infinite). Defaults to `-1`.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(-1),
					},
					"catch_up_takeover_delay_millis": schema.Int64Attribute{
						Description: "Delay before catch-up takeover. Defaults to `30000`.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(30000),
					},
					"get_last_error_defaults": schema.SingleNestedAttribute{
						Description: "Default error handling settings",
						Optional:    true,
						Computed:    true,
						Default:     objectdefault.StaticValue(defaultGetLastErrorDefaults),
						Attributes: map[string]schema.Attribute{
							"w": schema.Int64Attribute{
								Description: "Write concern value. Defaults to `1`.",
								Optional:    true,
								Computed:    true,
								Default:     int64default.StaticInt64(1),
							},
							"wtimeout": schema.Int64Attribute{
								Description: "Write concern timeout. Defaults to `0`.",
								Optional:    true,
								Computed:    true,
								Default:     int64default.StaticInt64(0),
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

	actual, exist, err := r.client.Resource().ReplicaSet().Exists(apiCtx, state)
	if err != nil {
		resp.Diagnostics.AddError("Failed to check replica set existence", err.Error())
		return
	}

	if !exist {
		resp.State.RemoveResource(ctx)
		return
	}

	actual.Timeouts = state.Timeouts
	resp.Diagnostics.Append(resp.State.Set(ctx, &actual)...)
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
	resp.Diagnostics.AddError(
		"Delete Not Supported",
		"The delete method is not implemented for this resource, because it requires manual actions from the administrator. After manually deleting you should clean up the state by running `terraform state rm mongodb_replicaset.<name>`.",
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
