// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package classification

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                     = &ClassifierResource{}
	_ resource.ResourceWithImportState      = &ClassifierResource{}
	_ resource.ResourceWithConfigValidators = &ClassifierResource{}
)

func NewClassifierResource() resource.Resource {
	return &ClassifierResource{}
}

type ClassifierResource struct {
	client *client.Client
}

type ClassifierResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Pattern          types.String `tfsdk:"pattern"`
	MinimumThreshold types.Int64  `tfsdk:"minimum_threshold"`
	SampleSize       types.Int64  `tfsdk:"sample_size"`
	SampleType       types.String `tfsdk:"sample_type"`
	CompoundRuleset  types.String `tfsdk:"compound_ruleset"`
}

func (r *ClassifierResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_classifier"
}

func (r *ClassifierResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a custom data classifier used for automated data classification. " +
			"A classifier detects sensitive data with a regex `pattern` sampled against column values, " +
			"a `compound_ruleset` of conditions (row data, metadata, column location, content type, data length, column size, Amazon Comprehend), " +
			"or both combined (regex plus a metadata filter). At least one of `pattern` or `compound_ruleset` must be set.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Classifier identifier (same as name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Unique name for the classifier. Names starting with 'ALTR' are reserved. Changing the name forces replacement.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable explanation of what the classifier detects. Removing the attribute keeps the last value; set it to an empty string to clear it.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(400),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pattern": schema.StringAttribute{
				Description: "RE2 regex pattern matched against sampled column values. At least one of `pattern` or `compound_ruleset` must be set.",
				Optional:    true,
				Validators: []validator.String{
					// Min length 1: an explicit empty pattern would bypass the
					// at-least-one-of check (empty string is non-null), be
					// dropped by omitempty on create, and read back as null.
					stringvalidator.LengthBetween(1, 750),
				},
			},
			"minimum_threshold": schema.Int64Attribute{
				Description: "Percent (1-100) of sampled values that must match for a column to be considered a match. Defaults to 70. Removing the attribute keeps the last applied value.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 100),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"sample_size": schema.Int64Attribute{
				Description: "Number of values sampled per column. Sampling is configured on the classification job, not the classifier.",
				Computed:    true,
			},
			"sample_type": schema.StringAttribute{
				Description: "How values are sampled (e.g. ROWS). Sampling is configured on the classification job, not the classifier.",
				Computed:    true,
			},
			"compound_ruleset": schema.StringAttribute{
				Description: "JSON object describing a compound ruleset: an `operator` (AND/OR) over `conditions` " +
					"targeting ROW_DATA, METADATA, COLUMN_LOCATION, CONTENT_TYPE, DATA_LENGTH, COLUMN_SIZE, or AMAZON_COMPREHEND. " +
					"Combine with `pattern` to narrow regex matches with metadata or location filters. " +
					"To negate a condition or group, set its `negated` boolean — there is no NOT operator; the API rejects one. " +
					"The API normalizes the ruleset on write (e.g. adds default `negated` flags), so the value is tracked as written: " +
					"out-of-band edits to the ruleset are NOT detected as drift, and after import the state holds the normalized form, " +
					"so the first plan shows an update unless the configuration matches it.",
				Optional: true,
				Validators: []validator.String{
					jsonObjectValidator{},
				},
			},
		},
	}
}

func (r *ClassifierResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("pattern"),
			path.MatchRoot("compound_ruleset"),
		),
	}
}

func (r *ClassifierResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = c
}

func (r *ClassifierResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ClassifierResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	input := client.CreateClassifierInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Pattern:     plan.Pattern.ValueString(),
	}

	if !plan.MinimumThreshold.IsNull() && !plan.MinimumThreshold.IsUnknown() {
		v := int(plan.MinimumThreshold.ValueInt64())
		input.MinimumThreshold = &v
	}

	if !plan.CompoundRuleset.IsNull() {
		input.CompoundRuleset = json.RawMessage(plan.CompoundRuleset.ValueString())
	}

	classifier, err := r.client.CreateClassifier(input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating classifier",
			"Could not create classifier, unexpected error: "+err.Error(),
		)

		return
	}

	r.mapClassifierToModel(classifier, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ClassifierResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ClassifierResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	classifier, err := r.client.GetClassifier(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading classifier",
			"Could not read classifier "+state.Name.ValueString()+": "+err.Error(),
		)

		return
	}

	if classifier == nil {
		resp.State.RemoveResource(ctx)

		return
	}

	r.mapClassifierToModel(classifier, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ClassifierResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  ClassifierResourceModel
		state ClassifierResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	input := client.UpdateClassifierInput{}

	if !plan.Description.Equal(state.Description) {
		input.Description = plan.Description.ValueStringPointer()
	}

	if !plan.Pattern.Equal(state.Pattern) {
		if plan.Pattern.IsNull() {
			input.RemovePattern = true
		} else {
			input.Pattern = plan.Pattern.ValueStringPointer()
		}
	}

	if !plan.MinimumThreshold.Equal(state.MinimumThreshold) && !plan.MinimumThreshold.IsNull() && !plan.MinimumThreshold.IsUnknown() {
		v := int(plan.MinimumThreshold.ValueInt64())
		input.MinimumThreshold = &v
	}

	if !plan.CompoundRuleset.Equal(state.CompoundRuleset) {
		if plan.CompoundRuleset.IsNull() {
			input.RemoveCompoundRuleset = true
		} else {
			input.CompoundRuleset = json.RawMessage(plan.CompoundRuleset.ValueString())
		}
	}

	classifier, err := r.client.UpdateClassifier(state.Name.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating classifier",
			"Could not update classifier, unexpected error: "+err.Error(),
		)

		return
	}

	r.mapClassifierToModel(classifier, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ClassifierResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ClassifierResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteClassifier(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting classifier",
			"Could not delete classifier, unexpected error: "+err.Error(),
		)

		return
	}
}

func (r *ClassifierResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID is the classifier name.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

// mapClassifierToModel maps API fields onto the model. compound_ruleset is
// normalized server-side (e.g. default negated flags are added), so a value
// already held by the model is preserved as written; the API value is only
// taken when the model has none (import).
func (r *ClassifierResource) mapClassifierToModel(classifier *client.Classifier, model *ClassifierResourceModel) {
	model.ID = types.StringValue(classifier.Name)
	model.Name = types.StringValue(classifier.Name)
	model.Description = types.StringValue(classifier.Description)
	model.MinimumThreshold = types.Int64Value(int64(classifier.MinimumThreshold))
	model.SampleSize = types.Int64Value(int64(classifier.SampleSize))
	model.SampleType = types.StringValue(classifier.SampleType)

	if classifier.Pattern != "" {
		model.Pattern = types.StringValue(classifier.Pattern)
	} else {
		model.Pattern = types.StringNull()
	}

	if len(classifier.CompoundRuleset) == 0 {
		model.CompoundRuleset = types.StringNull()
	} else if model.CompoundRuleset.IsNull() || model.CompoundRuleset.IsUnknown() {
		model.CompoundRuleset = types.StringValue(string(classifier.CompoundRuleset))
	}
}
