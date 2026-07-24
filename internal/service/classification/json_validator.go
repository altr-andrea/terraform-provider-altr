// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package classification

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// jsonObjectValidator validates that a string attribute holds a JSON object.
type jsonObjectValidator struct{}

func (v jsonObjectValidator) Description(_ context.Context) string {
	return "value must be a JSON object"
}

func (v jsonObjectValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v jsonObjectValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(req.ConfigValue.ValueString()), &obj); err != nil || obj == nil {
		message := "Attribute must contain a valid JSON object"
		if err != nil {
			message += ": " + err.Error()
		}

		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid JSON Object",
			message,
		)
	}
}
