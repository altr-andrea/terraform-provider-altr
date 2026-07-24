// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package classification

import (
	"context"
	"fmt"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &ClassifierCollectionResource{}
	_ resource.ResourceWithImportState = &ClassifierCollectionResource{}
)

func NewClassifierCollectionResource() resource.Resource {
	return &ClassifierCollectionResource{}
}

type ClassifierCollectionResource struct {
	client *client.Client
}

type ClassifierCollectionResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (r *ClassifierCollectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_classifier_collection"
}

func (r *ClassifierCollectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a classifier collection. Collections group classifiers together for use in " +
			"classification jobs and agent tasks. Manage membership with the `altr_classifier_collection_classifier` resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Collection identifier (same as name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Unique name for the collection. Names starting with 'ALTR' are reserved. Changing the name forces replacement.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of the collection. Must be non-empty when set: the update API ignores empty descriptions, so a description can be changed but not cleared.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 400),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ClassifierCollectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ClassifierCollectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ClassifierCollectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	input := client.CreateClassifierCollectionInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	collection, err := r.client.CreateClassifierCollection(input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating classifier collection",
			"Could not create classifier collection, unexpected error: "+err.Error(),
		)

		return
	}

	r.mapCollectionToModel(collection, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ClassifierCollectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ClassifierCollectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	collection, err := r.client.GetClassifierCollection(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading classifier collection",
			"Could not read classifier collection "+state.Name.ValueString()+": "+err.Error(),
		)

		return
	}

	if collection == nil {
		resp.State.RemoveResource(ctx)

		return
	}

	r.mapCollectionToModel(collection, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ClassifierCollectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  ClassifierCollectionResourceModel
		state ClassifierCollectionResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// The planned description is always known and non-empty here: the
	// validator rejects "", UseStateForUnknown resolves an omitted attribute
	// to the prior state (no diff, so no Update), and description is the only
	// in-place-updatable attribute. The update API ignores empty descriptions,
	// so this invariant is what makes the PATCH meaningful.
	collection, err := r.client.UpdateClassifierCollection(state.Name.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating classifier collection",
			"Could not update classifier collection, unexpected error: "+err.Error(),
		)

		return
	}

	r.mapCollectionToModel(collection, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ClassifierCollectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ClassifierCollectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteClassifierCollection(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting classifier collection",
			"Could not delete classifier collection, unexpected error: "+err.Error(),
		)

		return
	}
}

func (r *ClassifierCollectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID is the collection name.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

func (r *ClassifierCollectionResource) mapCollectionToModel(collection *client.ClassifierCollection, model *ClassifierCollectionResourceModel) {
	model.ID = types.StringValue(collection.Name)
	model.Name = types.StringValue(collection.Name)
	model.Description = types.StringValue(collection.Description)
}
