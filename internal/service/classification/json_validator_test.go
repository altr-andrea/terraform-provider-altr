// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package classification

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestJSONObjectValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		value       types.String
		expectError bool
	}{
		{"valid object", types.StringValue(`{"column_name_pattern": "(ssn)"}`), false},
		{"valid nested object", types.StringValue(`{"operator":"OR","conditions":[{"target":"CONTENT_TYPE"}]}`), false},
		{"empty object", types.StringValue(`{}`), false},
		{"invalid json", types.StringValue(`{not json`), true},
		{"json array", types.StringValue(`[1, 2, 3]`), true},
		{"json string", types.StringValue(`"hello"`), true},
		{"json null", types.StringValue(`null`), true},
		{"empty string", types.StringValue(``), true},
		{"null value", types.StringNull(), false},
		{"unknown value", types.StringUnknown(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}

			jsonObjectValidator{}.ValidateString(context.Background(), req, resp)

			if tt.expectError != resp.Diagnostics.HasError() {
				t.Errorf("expected error: %t, got diagnostics: %v", tt.expectError, resp.Diagnostics)
			}
		})
	}
}
