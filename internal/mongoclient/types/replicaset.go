package types

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	MongoDBSupportedVersions = "6, 7, 8"
)

var SupportedMajorVersions = []string{"6", "7", "8"}

type ReplicaSetConfig struct {
	Config ReplicaSet `bson:"config"`
}
type ReplicaSet struct {
	Name                               string         `tfsdk:"name" bson:"_id"`
	Version                            *int64         `tfsdk:"version" bson:"version,omitempty"`
	Members                            []Member       `tfsdk:"members" bson:"members"`
	ProtocolVersion                    *int64         `tfsdk:"protocol_version" bson:"protocolVersion,omitempty"`
	WriteConcernMajorityJournalDefault *bool          `tfsdk:"write_concern_majority_journal_default" bson:"writeConcernMajorityJournalDefault,omitempty"`
	Settings                           *Settings      `tfsdk:"settings" bson:"settings,omitempty"`
	Timeouts                           timeouts.Value `tfsdk:"timeouts" bson:"-"`
}

type Member struct {
	Id                 int64    `tfsdk:"id" bson:"_id"`
	Host               string   `tfsdk:"host" bson:"host"`
	ArbiterOnly        *bool    `tfsdk:"arbiter_only" bson:"arbiterOnly,omitempty"`
	BuildIndexes       *bool    `tfsdk:"build_indexes" bson:"buildIndexes,omitempty"`
	Hidden             *bool    `tfsdk:"hidden" bson:"hidden,omitempty"`
	Priority           *float64 `tfsdk:"priority" bson:"priority,omitempty"`
	SecondaryDelaySecs *int64   `tfsdk:"secondary_delay_secs" bson:"secondaryDelaySecs,omitempty"`
	Votes              *int64   `tfsdk:"votes" bson:"votes,omitempty"`
}

type Settings struct {
	ChainingAllowed            *bool                 `tfsdk:"chaining_allowed" bson:"chainingAllowed,omitempty"`
	HeartbeatIntervalMillis    *int64                `tfsdk:"heartbeat_interval_millis" bson:"heartbeatIntervalMillis,omitempty"`
	HeartbeatTimeoutSecs       *int64                `tfsdk:"heartbeat_timeout_secs" bson:"heartbeatTimeoutSecs,omitempty"`
	ElectionTimeoutMillis      *int64                `tfsdk:"election_timeout_millis" bson:"electionTimeoutMillis,omitempty"`
	CatchUpTimeoutMillis       *int64                `tfsdk:"catch_up_timeout_millis" bson:"catchUpTimeoutMillis,omitempty"`
	CatchUpTakeoverDelayMillis *int64                `tfsdk:"catch_up_takeover_delay_millis" bson:"catchUpTakeoverDelayMillis,omitempty"`
	GetLastErrorDefaults       *GetLastErrorDefaults `tfsdk:"get_last_error_defaults" bson:"getLastErrorDefaults,omitempty"`
}

type GetLastErrorDefaults struct {
	W        *int64 `tfsdk:"w" bson:"w,omitempty"`
	WTimeout *int64 `tfsdk:"wtimeout" bson:"wtimeout,omitempty"`
}

func (r *ReplicaSet) SetVersion(newVersion *int64) {
	r.Version = newVersion
}

func (r *ReplicaSet) ClearVersion() {
	r.Version = nil
}

func (r *ReplicaSet) ClearTimeouts() {
	rsTimeoutsAttrTypes := map[string]attr.Type{
		"create": types.StringType,
		"read":   types.StringType,
		"update": types.StringType,
	}

	r.Timeouts = timeouts.Value{
		Object: types.ObjectNull(rsTimeoutsAttrTypes),
	}
}

type ReplicaSetStatus struct {
	OK      int    `bson:"ok"`
	Set     string `bson:"set"`
	Members []struct {
		Name     string `bson:"name"`
		StateStr string `bson:"stateStr"`
		Health   int    `bson:"health"`
	} `bson:"members"`
}
