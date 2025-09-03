package types

import "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"

const (
	MongoDBRequiredVersion = "6"
)

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
	ChainingAllowed            bool                  `tfsdk:"chaining_allowed" bson:"chainingAllowed,omitempty"`
	HeartbeatIntervalMillis    int64                 `tfsdk:"heartbeat_interval_millis" bson:"heartbeatIntervalMillis,omitempty"`
	HeartbeatTimeoutSecs       int64                 `tfsdk:"heartbeat_timeout_secs" bson:"heartbeatTimeoutSecs,omitempty"`
	ElectionTimeoutMillis      int64                 `tfsdk:"election_timeout_millis" bson:"electionTimeoutMillis,omitempty"`
	CatchUpTimeoutMillis       int64                 `tfsdk:"catch_up_timeout_millis" bson:"catchUpTimeoutMillis,omitempty"`
	CatchUpTakeoverDelayMillis int64                 `tfsdk:"catch_up_takeover_delay_millis" bson:"catchUpTakeoverDelayMillis,omitempty"`
	GetLastErrorDefaults       *GetLastErrorDefaults `tfsdk:"get_last_error_defaults" bson:"getLastErrorDefaults,omitempty"`
}

type GetLastErrorDefaults struct {
	W        int64 `tfsdk:"w" bson:"w,omitempty"`
	WTimeout int64 `tfsdk:"wtimeout" bson:"wtimeout,omitempty"`
}

func (r *ReplicaSet) SetVersion(newVersion *int64) {
	r.Version = newVersion
}

func (r *ReplicaSet) ClearVersion() {
	r.Version = nil
}

func (r *ReplicaSet) RemoveDefaults() {
	r.ClearVersion()

	if r.Settings != nil {
		if r.Settings.ChainingAllowed == true &&
			r.Settings.HeartbeatIntervalMillis == 2000 &&
			r.Settings.HeartbeatTimeoutSecs == 10 &&
			r.Settings.ElectionTimeoutMillis == 10000 &&
			r.Settings.CatchUpTimeoutMillis == -1 &&
			r.Settings.CatchUpTakeoverDelayMillis == 30000 &&
			r.Settings.GetLastErrorDefaults != nil &&
			r.Settings.GetLastErrorDefaults.W == 1 &&
			r.Settings.GetLastErrorDefaults.WTimeout == 0 {
			r.Settings = nil
		}
	}

	if r.ProtocolVersion != nil && *r.ProtocolVersion == 1 {
		r.ProtocolVersion = nil
	}

	if r.WriteConcernMajorityJournalDefault != nil && *r.WriteConcernMajorityJournalDefault == true {
		r.WriteConcernMajorityJournalDefault = nil
	}

	for i := range r.Members {
		if r.Members[i].ArbiterOnly != nil && *r.Members[i].ArbiterOnly == false {
			r.Members[i].ArbiterOnly = nil
		}
		if r.Members[i].BuildIndexes != nil && *r.Members[i].BuildIndexes == true {
			r.Members[i].BuildIndexes = nil
		}
		if r.Members[i].Hidden != nil && *r.Members[i].Hidden == false {
			r.Members[i].Hidden = nil
		}
		if r.Members[i].Priority != nil && *r.Members[i].Priority == 1 {
			r.Members[i].Priority = nil
		}
		if r.Members[i].SecondaryDelaySecs != nil && *r.Members[i].SecondaryDelaySecs == 0 {
			r.Members[i].SecondaryDelaySecs = nil
		}
		if r.Members[i].Votes != nil && *r.Members[i].Votes == 1 {
			r.Members[i].Votes = nil
		}
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
