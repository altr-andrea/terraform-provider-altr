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

var _ datasource.DataSource = &ClassifierCollectionDataSource{}

func NewClassifierCollectionDataSource() datasource.DataSource {
	return &ClassifierCollectionDataSource{}
}

type ClassifierCollectionDataSource struct {
	client *client.Client
}

type ClassifierCollectionDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	ClassifierCount types.Int64  `tfsdk:"classifier_count"`
	ClassifierNames types.List   `tfsdk:"classifier_names"`
}

func (d *ClassifierCollectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_classifier_collection"
}

func (d *ClassifierCollectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving information about a classifier collection, including ALTR-managed collections.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Collection identifier (same as name).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the collection.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of the collection.",
				Computed:    true,
			},
			"classifier_count": schema.Int64Attribute{
				Description: "Number of classifiers in the collection.",
				Computed:    true,
			},
			"classifier_names": schema.ListAttribute{
				Description: "Names of the classifiers in the collection.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *ClassifierCollectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ClassifierCollectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ClassifierCollectionDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	collection, err := d.client.GetClassifierCollection(config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading classifier collection",
			"Could not read classifier collection "+config.Name.ValueString()+": "+err.Error(),
		)

		return
	}

	if collection == nil {
		resp.Diagnostics.AddError(
			"Classifier collection not found",
			"Classifier collection with name '"+config.Name.ValueString()+"' does not exist.",
		)

		return
	}

	classifiers, err := d.client.ListClassifierCollectionClassifiers(collection.Name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading collection classifiers",
			"Could not read classifiers in collection "+collection.Name+": "+err.Error(),
		)

		return
	}

	classifierNames := make([]string, 0, len(classifiers))
	for _, classifier := range classifiers {
		classifierNames = append(classifierNames, classifier.Name)
	}

	config.ID = types.StringValue(collection.Name)
	config.Name = types.StringValue(collection.Name)
	config.Description = types.StringValue(collection.Description)
	config.ClassifierCount = types.Int64Value(int64(collection.ClassifierCount))

	names, diags := types.ListValueFrom(ctx, types.StringType, classifierNames)
	resp.Diagnostics.Append(diags...)
	config.ClassifierNames = names

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
