package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/michael/terraform-provider-veeam/pkg/client"
)

var _ resource.Resource = &RepositoryScaleOutResource{}
var _ resource.ResourceWithImportState = &RepositoryScaleOutResource{}

type RepositoryScaleOutResource struct {
	client *client.Client
}

type RepositoryScaleOutResourceModel struct {
	ID                     types.String   `tfsdk:"id"`
	Name                   types.String   `tfsdk:"name"`
	Description            types.String   `tfsdk:"description"`
	PerformanceExtentIDs   []types.String `tfsdk:"performance_extent_ids"`
}

func NewRepositoryScaleOutResource() resource.Resource {
	return &RepositoryScaleOutResource{}
}

func (r *RepositoryScaleOutResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sobr_repository"
}

func (r *RepositoryScaleOutResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Scale-Out Backup Repository (SOBR) in Veeam Backup & Replication.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier (GUID) of the Scale-Out Backup Repository in Veeam.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the Scale-Out Backup Repository.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the Scale-Out Backup Repository.",
			},
			"performance_extent_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of backup repository GUIDs to combine into the Performance Tier.",
			},
		},
	}
}

func (r *RepositoryScaleOutResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *RepositoryScaleOutResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RepositoryScaleOutResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	extents := make([]client.ScaleOutExtentSpec, 0, len(plan.PerformanceExtentIDs))
	for _, extID := range plan.PerformanceExtentIDs {
		extents = append(extents, client.ScaleOutExtentSpec{
			ID: extID.ValueString(),
		})
	}

	spec := client.CreateScaleOutRepositorySpec{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		PerformanceTier: client.PerformanceTierSpec{
			PerformanceExtents: extents,
		},
	}

	created, err := r.client.CreateScaleOutRepository(ctx, spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Scale-Out Backup Repository", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RepositoryScaleOutResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RepositoryScaleOutResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sobr, err := r.client.GetScaleOutRepositoryByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Scale-Out Backup Repository", err.Error())
		return
	}

	if sobr == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(sobr.Name)
	if sobr.Description != "" {
		state.Description = types.StringValue(sobr.Description)
	}

	extentIDs := make([]types.String, 0, len(sobr.PerformanceTier.PerformanceExtents))
	for _, ext := range sobr.PerformanceTier.PerformanceExtents {
		extentIDs = append(extentIDs, types.StringValue(ext.ID))
	}
	state.PerformanceExtentIDs = extentIDs

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RepositoryScaleOutResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Scale-Out Repository modification is not directly supported; replace resource instead.",
	)
}

func (r *RepositoryScaleOutResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RepositoryScaleOutResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteScaleOutRepository(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Scale-Out Backup Repository", err.Error())
		return
	}
}

func (r *RepositoryScaleOutResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
