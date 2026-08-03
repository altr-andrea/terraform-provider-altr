// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package alerting

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &DamAlertRuleResource{}
	_ resource.ResourceWithImportState = &DamAlertRuleResource{}
)

func NewDamAlertRuleResource() resource.Resource {
	return &DamAlertRuleResource{}
}

type DamAlertRuleResource struct {
	client *client.Client
}

type DamAlertRuleResourceModel struct {
	ID              types.String         `tfsdk:"id"`
	Name            types.String         `tfsdk:"name"`
	Description     types.String         `tfsdk:"description"`
	Severity        types.String         `tfsdk:"severity"`
	Enabled         types.Bool           `tfsdk:"enabled"`
	DataSourceScope types.String         `tfsdk:"data_source_scope"`
	RuleType        types.String         `tfsdk:"rule_type"`
	EmailRecipients types.Set            `tfsdk:"email_recipients"`
	FilterTree      jsontypes.Normalized `tfsdk:"filter_tree"`
}

func (r *DamAlertRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dam_alert_rule"
}

func (r *DamAlertRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DAM (Database Activity Monitoring) alert rule. Any change replaces the rule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the alert rule, assigned by the API.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the alert rule. Rule names are unique within an organization.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the alert rule.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"severity": schema.StringAttribute{
				Description: "Severity of alerts raised by this rule. One of `low`, `medium`, `high`, `critical`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("low", "medium", "high", "critical"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the alert rule is enabled. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"data_source_scope": schema.StringAttribute{
				Description: "Data sources the rule applies to. One of `oltp`, `snowflake`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("oltp", "snowflake"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rule_type": schema.StringAttribute{
				Description: "Type of the rule. One of `match`, `threshold`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("match", "threshold"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email_recipients": schema.SetAttribute{
				Description: "Email addresses that receive alerts raised by this rule.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"filter_tree": schema.StringAttribute{
				Description: "JSON-encoded filter tree that scopes the rule. Each node has `op`, `dimension`, " +
					"and one of `string_value`, `list_value`, `number_value`, or `children`. Which operators are " +
					"valid depends on the dimension; the API rejects bad combinations.",
				Optional:   true,
				CustomType: jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *DamAlertRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DamAlertRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DamAlertRuleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	input := client.CreateDamAlertRuleInput{
		Name:            plan.Name.ValueString(),
		Description:     plan.Description.ValueString(),
		Severity:        plan.Severity.ValueString(),
		Enabled:         plan.Enabled.ValueBool(),
		DataSourceScope: plan.DataSourceScope.ValueString(),
		RuleType:        plan.RuleType.ValueString(),
	}

	if !plan.EmailRecipients.IsNull() {
		resp.Diagnostics.Append(plan.EmailRecipients.ElementsAs(ctx, &input.EmailRecipients, false)...)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !plan.FilterTree.IsNull() {
		input.FilterTree = json.RawMessage(plan.FilterTree.ValueString())
	}

	// The API is idempotent by name, so a duplicate create would quietly take
	// over the existing rule. Two concurrent applies can still race past this.
	existing, err := r.client.GetDamAlertRuleByName(plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating DAM alert rule",
			"Could not check for existing DAM alert rule, unexpected error: "+err.Error(),
		)

		return
	}

	if existing != nil {
		resp.Diagnostics.AddError(
			"DAM Alert Rule Already Exists",
			fmt.Sprintf("A DAM alert rule named %q already exists. Import it: terraform import <address> %s",
				plan.Name.ValueString(), existing.RuleID),
		)

		return
	}

	rule, err := r.client.CreateDamAlertRule(input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating DAM alert rule",
			"Could not create DAM alert rule, unexpected error: "+err.Error(),
		)

		return
	}

	plan.ID = types.StringValue(rule.RuleID)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *DamAlertRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DamAlertRuleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.GetDamAlertRule(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading DAM alert rule",
			"Could not read DAM alert rule "+state.ID.ValueString()+": "+err.Error(),
		)

		return
	}

	if rule == nil {
		resp.State.RemoveResource(ctx)

		return
	}

	resp.Diagnostics.Append(mapDamAlertRuleToModel(ctx, rule, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *DamAlertRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No update endpoint. Every attribute is RequiresReplace, so this
	// shouldn't be reachable.
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"DAM alert rules cannot be updated in place. This is a bug in the provider; please report it.",
	)
}

func (r *DamAlertRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DamAlertRuleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDamAlertRule(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting DAM alert rule",
			"Could not delete DAM alert rule, unexpected error: "+err.Error(),
		)

		return
	}
}

func (r *DamAlertRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapDamAlertRuleToModel(ctx context.Context, rule *client.DamAlertRule, model *DamAlertRuleResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(rule.RuleID)
	model.Name = types.StringValue(rule.Name)
	model.Severity = types.StringValue(rule.Severity)
	model.Enabled = types.BoolValue(rule.Enabled)
	model.DataSourceScope = types.StringValue(rule.DataSourceScope)
	model.RuleType = types.StringValue(rule.RuleType)

	// Keep an explicitly empty value empty instead of flipping it to null,
	// which would force a replace.
	if rule.Description == "" {
		if model.Description.IsNull() || model.Description.ValueString() != "" {
			model.Description = types.StringNull()
		}
	} else {
		model.Description = types.StringValue(rule.Description)
	}

	if len(rule.EmailRecipients) == 0 {
		if model.EmailRecipients.IsNull() || len(model.EmailRecipients.Elements()) != 0 {
			model.EmailRecipients = types.SetNull(types.StringType)
		}
	} else {
		recipients, d := types.SetValueFrom(ctx, types.StringType, rule.EmailRecipients)
		diags.Append(d...)
		model.EmailRecipients = recipients
	}

	if len(rule.FilterTree) == 0 || string(rule.FilterTree) == "null" {
		model.FilterTree = jsontypes.NewNormalizedNull()
	} else {
		model.FilterTree = jsontypes.NewNormalizedValue(string(rule.FilterTree))
	}

	return diags
}
