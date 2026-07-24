// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"encoding/json"
	"testing"
)

func TestUpdateClassifierInputMarshalJSON(t *testing.T) {
	t.Parallel()

	strPtr := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name  string
		input UpdateClassifierInput
		want  string
	}{
		{
			// Omitted fields must be absent so the API preserves them.
			name:  "empty input omits all fields",
			input: UpdateClassifierInput{},
			want:  `{}`,
		},
		{
			name:  "set description",
			input: UpdateClassifierInput{Description: strPtr("updated")},
			want:  `{"description":"updated"}`,
		},
		{
			// Explicit empty string clears the description server-side.
			name:  "clear description",
			input: UpdateClassifierInput{Description: strPtr("")},
			want:  `{"description":""}`,
		},
		{
			name:  "set pattern",
			input: UpdateClassifierInput{Pattern: strPtr(`\d{3}`)},
			want:  `{"pattern":"\\d{3}"}`,
		},
		{
			// Explicit null clears the pattern server-side.
			name:  "remove pattern",
			input: UpdateClassifierInput{RemovePattern: true},
			want:  `{"pattern":null}`,
		},
		{
			name:  "remove wins over pattern value",
			input: UpdateClassifierInput{Pattern: strPtr("x"), RemovePattern: true},
			want:  `{"pattern":null}`,
		},
		{
			name:  "set minimum threshold",
			input: UpdateClassifierInput{MinimumThreshold: intPtr(80)},
			want:  `{"minimum_threshold":80}`,
		},
		{
			name:  "set compound ruleset",
			input: UpdateClassifierInput{CompoundRuleset: json.RawMessage(`{"operator":"OR"}`)},
			want:  `{"compound_ruleset":{"operator":"OR"}}`,
		},
		{
			// Explicit null clears the compound ruleset server-side.
			name:  "remove compound ruleset",
			input: UpdateClassifierInput{RemoveCompoundRuleset: true},
			want:  `{"compound_ruleset":null}`,
		},
		{
			// Swapping modes clears one field and sets the other in one PATCH.
			name: "swap ruleset for pattern",
			input: UpdateClassifierInput{
				Pattern:               strPtr("x"),
				RemoveCompoundRuleset: true,
			},
			want: `{"compound_ruleset":null,"pattern":"x"}`,
		},
		{
			name: "swap pattern for ruleset",
			input: UpdateClassifierInput{
				RemovePattern:   true,
				CompoundRuleset: json.RawMessage(`{"operator":"AND"}`),
			},
			want: `{"compound_ruleset":{"operator":"AND"},"pattern":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("unexpected marshal error: %v", err)
			}

			if string(got) != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
