// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package classification

import (
	"context"
	"fmt"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ClassifierDataSource{}

func NewClassifierDataSource() datasource.DataSource {
	return &ClassifierDataSource{}
}

type ClassifierDataSource struct {
	client *client.Client
}

type ClassifierDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Pattern          types.String `tfsdk:"pattern"`
	MinimumThreshold types.Int64  `tfsdk:"minimum_threshold"`
	SampleSize       types.Int64  `tfsdk:"sample_size"`
	SampleType       types.String `tfsdk:"sample_type"`
	CompoundRuleset  types.String `tfsdk:"compound_ruleset"`
	CollectionNames  types.List   `tfsdk:"collection_names"`
	CollectionCount  types.Int64  `tfsdk:"collection_count"`
}

func (d *ClassifierDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_classifier"
}

func (d *ClassifierDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving information about a data classifier, including ALTR-managed classifiers.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Classifier identifier (same as name).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the classifier.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable explanation of what the classifier detects.",
				Computed:    true,
			},
			"pattern": schema.StringAttribute{
				Description: "Regex pattern matched against sampled column values.",
				Computed:    true,
			},
			"minimum_threshold": schema.Int64Attribute{
				Description: "Percent (1-100) of sampled values that must match for a column to be considered a match.",
				Computed:    true,
			},
			"sample_size": schema.Int64Attribute{
				Description: "Number of values sampled per column when evaluating the classifier.",
				Computed:    true,
			},
			"sample_type": schema.StringAttribute{
				Description: "How values are sampled (e.g. ROWS). Empty for compound ruleset classifiers.",
				Computed:    true,
			},
			"compound_ruleset": schema.StringAttribute{
				Description: "JSON object describing the compound ruleset, if the classifier uses one.",
				Computed:    true,
			},
			"collection_names": schema.ListAttribute{
				Description: "Names of the collections this classifier belongs to.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"collection_count": schema.Int64Attribute{
				Description: "Number of collections this classifier belongs to.",
				Computed:    true,
			},
		},
	}
}

func (d *ClassifierDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = c
}

func (d *ClassifierDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ClassifierDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	classifier, err := d.client.GetClassifier(config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading classifier",
			"Could not read classifier "+config.Name.ValueString()+": "+err.Error(),
		)

		return
	}

	if classifier == nil {
		resp.Diagnostics.AddError(
			"Classifier not found",
			"Classifier with name '"+config.Name.ValueString()+"' does not exist.",
		)

		return
	}

	config.ID = types.StringValue(classifier.Name)
	config.Name = types.StringValue(classifier.Name)
	config.Description = types.StringValue(classifier.Description)
	config.MinimumThreshold = types.Int64Value(int64(classifier.MinimumThreshold))
	config.SampleSize = types.Int64Value(int64(classifier.SampleSize))
	config.SampleType = types.StringValue(classifier.SampleType)
	config.CollectionCount = types.Int64Value(int64(classifier.CollectionCount))

	if classifier.Pattern != "" {
		config.Pattern = types.StringValue(classifier.Pattern)
	} else {
		config.Pattern = types.StringNull()
	}

	if len(classifier.CompoundRuleset) > 0 {
		config.CompoundRuleset = types.StringValue(string(classifier.CompoundRuleset))
	} else {
		config.CompoundRuleset = types.StringNull()
	}

	collectionNames, diags := types.ListValueFrom(ctx, types.StringType, classifier.CollectionNames)
	resp.Diagnostics.Append(diags...)
	config.CollectionNames = collectionNames

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
