package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/michael/terraform-provider-veeam/pkg/client"
)

var _ datasource.DataSource = &ScaleOutRepositoriesDataSource{}

type ScaleOutRepositoriesDataSource struct {
	client *client.Client
}

type ScaleOutRepositoryModel struct {
	ID                  types.String   `tfsdk:"id"`
	Name                types.String   `tfsdk:"name"`
	Description         types.String   `tfsdk:"description"`
	PerformanceExtents  []types.String `tfsdk:"performance_extent_ids"`
}

type ScaleOutRepositoriesDataSourceModel struct {
	Repositories []ScaleOutRepositoryModel `tfsdk:"repositories"`
}

func NewScaleOutRepositoriesDataSource() datasource.DataSource {
	return &ScaleOutRepositoriesDataSource{}
}

func (d *ScaleOutRepositoriesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scale_out_repositories"
}

func (d *ScaleOutRepositoriesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Scale-Out Backup Repositories (SOBR) configured in Veeam Backup & Replication.",
		Attributes: map[string]schema.Attribute{
			"repositories": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of Scale-Out Backup Repositories.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique identifier (GUID) of the scale-out backup repository.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the scale-out backup repository.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Description of the scale-out backup repository.",
						},
						"performance_extent_ids": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "List of performance extent repository GUIDs.",
						},
					},
				},
			},
		},
	}
}

func (d *ScaleOutRepositoriesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *ScaleOutRepositoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ScaleOutRepositoriesDataSourceModel

	sobrs, err := d.client.GetScaleOutRepositories(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to Read Scale-Out Backup Repositories", err.Error())
		return
	}

	state.Repositories = make([]ScaleOutRepositoryModel, 0, len(sobrs))
	for _, sobr := range sobrs {
		extentIDs := make([]types.String, 0, len(sobr.PerformanceTier.PerformanceExtents))
		for _, ext := range sobr.PerformanceTier.PerformanceExtents {
			extentIDs = append(extentIDs, types.StringValue(ext.ID))
		}

		state.Repositories = append(state.Repositories, ScaleOutRepositoryModel{
			ID:                 types.StringValue(sobr.ID),
			Name:               types.StringValue(sobr.Name),
			Description:        types.StringValue(sobr.Description),
			PerformanceExtents: extentIDs,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
