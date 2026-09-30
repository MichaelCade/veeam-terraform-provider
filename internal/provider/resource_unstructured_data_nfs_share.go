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

var _ resource.Resource = &UnstructuredDataNfsShareResource{}
var _ resource.ResourceWithImportState = &UnstructuredDataNfsShareResource{}

type UnstructuredDataNfsShareResource struct {
	client *client.Client
}

type UnstructuredDataNfsShareResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Path              types.String `tfsdk:"path"`
	CacheRepositoryID types.String `tfsdk:"cache_repository_id"`
}

func NewUnstructuredDataNfsShareResource() resource.Resource {
	return &UnstructuredDataNfsShareResource{}
}

func (r *UnstructuredDataNfsShareResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unstructured_data_nfs_share"
}

func (r *UnstructuredDataNfsShareResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Registers an NFS share export as an unstructured data backup source in Veeam Inventory.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier (GUID) of the NFS share server in Veeam Inventory.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"path": schema.StringAttribute{
				Required:    true,
				Description: "Path to the NFS export in server:/folder format (e.g. 192.168.169.3:/data/veeam_nfs).",
			},
			"cache_repository_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Backup repository ID used as cache repository for processing file changes. Defaults to Veeam Default Backup Repository if omitted.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *UnstructuredDataNfsShareResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UnstructuredDataNfsShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UnstructuredDataNfsShareResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cacheRepoID := plan.CacheRepositoryID.ValueString()
	if cacheRepoID == "" {
		defaultID, err := r.client.GetDefaultCacheRepositoryID(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Error Resolving Cache Repository", err.Error())
			return
		}
		cacheRepoID = defaultID
	}

	spec := client.CreateNfsShareServerSpec{
		Type: "NFSShare",
		Path: plan.Path.ValueString(),
		Processing: client.UnstructuredDataProcessingModel{
			CacheRepositoryID: cacheRepoID,
			BackupProxies: client.BackupProxiesSettingsModel{
				AutoSelectEnabled: true,
			},
		},
	}

	created, err := r.client.CreateUnstructuredDataServer(ctx, spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating NFS Share Source", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.CacheRepositoryID = types.StringValue(cacheRepoID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UnstructuredDataNfsShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UnstructuredDataNfsShareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	srv, err := r.client.GetUnstructuredDataServerByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading NFS Share Source", err.Error())
		return
	}

	if srv == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if srv.Path != "" {
		state.Path = types.StringValue(srv.Path)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *UnstructuredDataNfsShareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"NFS Share source modification is not directly supported; replace resource instead.",
	)
}

func (r *UnstructuredDataNfsShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UnstructuredDataNfsShareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUnstructuredDataServer(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting NFS Share Source", err.Error())
		return
	}
}

func (r *UnstructuredDataNfsShareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
