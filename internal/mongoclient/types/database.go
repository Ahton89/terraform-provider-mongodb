package types

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	DefaultDatabases = []string{"admin", "config", "local"} // Default databases to exclude from listing
	DefaultDatabase  = "admin"                              // Default database for making queries
)

type Databases struct {
	Databases []Database `tfsdk:"databases"`
}

type Database struct {
	Name     string         `tfsdk:"name"`
	Timeouts timeouts.Value `tfsdk:"timeouts" bson:"-"`
}

func (d *Database) ClearTimeouts() {
	rsTimeoutsAttrTypes := map[string]attr.Type{
		"create": types.StringType,
		"read":   types.StringType,
		"delete": types.StringType,
	}

	d.Timeouts = timeouts.Value{
		Object: types.ObjectNull(rsTimeoutsAttrTypes),
	}
}

func (d *Database) GetTimeouts() timeouts.Value  { return d.Timeouts }
func (d *Database) SetTimeouts(v timeouts.Value) { d.Timeouts = v }
