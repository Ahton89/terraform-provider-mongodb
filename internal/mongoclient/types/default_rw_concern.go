package types

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	DefaultRWConcernID    = "default"
	RWConcernSourceGlobal = "global"
)

var ReadConcernLevels = []string{"local", "available", "majority"}

type DefaultRWConcern struct {
	Id                  types.String   `tfsdk:"id"`
	DefaultReadConcern  *ReadConcern   `tfsdk:"default_read_concern"`
	DefaultWriteConcern *WriteConcern  `tfsdk:"default_write_concern"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
}

type DefaultRWConcernInfo struct {
	DefaultReadConcern        *ReadConcern  `tfsdk:"default_read_concern"`
	DefaultReadConcernSource  string        `tfsdk:"default_read_concern_source"`
	DefaultWriteConcern       *WriteConcern `tfsdk:"default_write_concern"`
	DefaultWriteConcernSource string        `tfsdk:"default_write_concern_source"`
}

type ReadConcern struct {
	Level string `tfsdk:"level"`
}

type WriteConcern struct {
	W        string `tfsdk:"w"`
	Journal  *bool  `tfsdk:"journal"`
	WTimeout *int64 `tfsdk:"wtimeout"`
}

func (d *DefaultRWConcern) ClearTimeouts() {
	timeoutsAttrTypes := map[string]attr.Type{
		"create": types.StringType,
		"read":   types.StringType,
		"update": types.StringType,
	}

	d.Timeouts = timeouts.Value{
		Object: types.ObjectNull(timeoutsAttrTypes),
	}
}

// RWConcernDefaults is the getDefaultRWConcern command response.
type RWConcernDefaults struct {
	DefaultReadConcern        *ReadConcernDocument  `bson:"defaultReadConcern"`
	DefaultReadConcernSource  string                `bson:"defaultReadConcernSource"`
	DefaultWriteConcern       *WriteConcernDocument `bson:"defaultWriteConcern"`
	DefaultWriteConcernSource string                `bson:"defaultWriteConcernSource"`
}

type ReadConcernDocument struct {
	Level string `bson:"level"`
}

type WriteConcernDocument struct {
	W        any    `bson:"w"`
	J        *bool  `bson:"j"`
	WTimeout *int64 `bson:"wtimeout"`
}

// Global returns only the defaults set explicitly with setDefaultRWConcern, skipping the implicit server defaults.
func (d *RWConcernDefaults) Global() DefaultRWConcern {
	r := DefaultRWConcern{Id: types.StringValue(DefaultRWConcernID)}

	if d.DefaultReadConcernSource == RWConcernSourceGlobal {
		r.DefaultReadConcern = d.readConcern()
	}

	if d.DefaultWriteConcernSource == RWConcernSourceGlobal {
		r.DefaultWriteConcern = d.writeConcern()
	}

	return r
}

func (d *RWConcernDefaults) Info() DefaultRWConcernInfo {
	return DefaultRWConcernInfo{
		DefaultReadConcern:        d.readConcern(),
		DefaultReadConcernSource:  d.DefaultReadConcernSource,
		DefaultWriteConcern:       d.writeConcern(),
		DefaultWriteConcernSource: d.DefaultWriteConcernSource,
	}
}

func (d *RWConcernDefaults) readConcern() *ReadConcern {
	if d.DefaultReadConcern == nil || d.DefaultReadConcern.Level == "" {
		return nil
	}

	return &ReadConcern{Level: d.DefaultReadConcern.Level}
}

func (d *RWConcernDefaults) writeConcern() *WriteConcern {
	if d.DefaultWriteConcern == nil {
		return nil
	}

	return &WriteConcern{
		W:        formatWriteConcernW(d.DefaultWriteConcern.W),
		Journal:  d.DefaultWriteConcern.J,
		WTimeout: d.DefaultWriteConcern.WTimeout,
	}
}

func formatWriteConcernW(w any) string {
	switch v := w.(type) {
	case nil:
		return ""
	case string:
		return v
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return fmt.Sprint(v)
	}
}
