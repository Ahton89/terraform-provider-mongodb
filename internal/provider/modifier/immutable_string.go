package modifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
)

type immutableString struct{}

func (m immutableString) Description(_ context.Context) string {
	return "Prevents updates to the attribute"
}

func (m immutableString) MarkdownDescription(_ context.Context) string {
	return "Prevents updates to the attribute"
}

func (m immutableString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue != tftypes.StringNull() && req.StateValue != tftypes.StringValue("") && req.StateValue != req.PlanValue {
		resp.Diagnostics.AddError(
			"Immutable Attribute Error",
			fmt.Sprintf("Attribute cannot be changed. Attempted to change from '%s' to '%s'.", req.StateValue, req.PlanValue),
		)
	}
}

func ImmutableString() planmodifier.String {
	return immutableString{}
}
