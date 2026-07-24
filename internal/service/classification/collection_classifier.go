// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package classification

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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
	_ resource.Resource                = &CollectionClassifierResource{}
	_ resource.ResourceWithImportState = &CollectionClassifierResource{}
)

// noColonRegex rejects ':' in names so the collection_name:classifier_name
// import ID stays unambiguous.
var noColonRegex = regexp.MustCompile(`^[^:]*$`)

func NewCollectionClassifierResource() resource.Resource {
	return &CollectionClassifierResource{}
}

type CollectionClassifierResource struct {
	client *client.Client
}

type CollectionClassifierResourceModel struct {
	ID             types.String `tfsdk:"id"`
	CollectionName types.String `tfsdk:"collection_name"`
	ClassifierName types.String `tfsdk:"classifier_name"`
}

func (r *CollectionClassifierResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_classifier_collection_classifier"
}

func (r *CollectionClassifierResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the membership of a single classifier in a classifier collection. " +
			"Use one instance per classifier (works well with `for_each`). " +
			"ALTR-managed collections cannot have classifiers added.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Membership identifier (collection_name:classifier_name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"collection_name": schema.StringAttribute{
				Description: "Name of the classifier collection. Must not contain ':' (it is the import ID separator).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(noColonRegex, "must not contain ':'"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"classifier_name": schema.StringAttribute{
				Description: "Name of the classifier to include in the collection. Must not contain ':' (it is the import ID separator).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(noColonRegex, "must not contain ':'"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *CollectionClassifierResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CollectionClassifierResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CollectionClassifierResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.AddClassifiersToCollection(plan.CollectionName.ValueString(), []string{plan.ClassifierName.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding classifier to collection",
			"Could not add classifier to collection, unexpected error: "+err.Error(),
		)

		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", plan.CollectionName.ValueString(), plan.ClassifierName.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CollectionClassifierResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CollectionClassifierResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// collection_names is a plain attribute on the classifier record (not a
	// paginated projection), so the single GET returns the complete list.
	classifier, err := r.client.GetClassifier(state.ClassifierName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading classifier",
			"Could not read classifier "+state.ClassifierName.ValueString()+": "+err.Error(),
		)

		return
	}

	if classifier != nil {
		for _, collectionName := range classifier.CollectionNames {
			if collectionName == state.CollectionName.ValueString() {
				state.ID = types.StringValue(fmt.Sprintf("%s:%s", state.CollectionName.ValueString(), state.ClassifierName.ValueString()))
				resp.Diagnostics.Append(resp.State.Set(ctx, state)...)

				return
			}
		}
	}

	// Classifier gone or no longer a member of the collection.
	resp.State.RemoveResource(ctx)
}

func (r *CollectionClassifierResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Both attributes require replacement, so Update is never reached through
	// a normal plan.
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Collection membership cannot be updated in place. Changes require replacement.",
	)
}

func (r *CollectionClassifierResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CollectionClassifierResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RemoveClassifiersFromCollection(state.CollectionName.ValueString(), []string{state.ClassifierName.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error removing classifier from collection",
			"Could not remove classifier from collection, unexpected error: "+err.Error(),
		)

		return
	}
}

func (r *CollectionClassifierResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Format: collection_name:classifier_name
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID in format: collection_name:classifier_name",
		)

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("collection_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("classifier_name"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
