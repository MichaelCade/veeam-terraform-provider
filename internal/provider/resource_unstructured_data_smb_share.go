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

var _ resource.Resource = &UnstructuredDataSmbShareResource{}
var _ resource.ResourceWithImportState = &UnstructuredDataSmbShareResource{}

type UnstructuredDataSmbShareResource struct {
	client *client.Client
}

type UnstructuredDataSmbShareResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Path              types.String `tfsdk:"path"`
	CredentialsID     types.String `tfsdk:"credentials_id"`
	CacheRepositoryID types.String `tfsdk:"cache_repository_id"`
}

func NewUnstructuredDataSmbShareResource() resource.Resource {
	return &UnstructuredDataSmbShareResource{}
}

func (r *UnstructuredDataSmbShareResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unstructured_data_smb_share"
}

func (r *UnstructuredDataSmbShareResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Registers an SMB file share as an unstructured data backup source in Veeam Inventory.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier (GUID) of the SMB share server in Veeam Inventory.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"path": schema.StringAttribute{
				Required:    true,
				Description: "UNC path to the SMB shared folder (e.g. \\\\readynas716\\data\\veeam_smb).",
			},
			"credentials_id": schema.StringAttribute{
				Optional:    true,
				Description: "Credentials ID used to authenticate against the SMB share.",
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

func (r *UnstructuredDataSmbShareResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UnstructuredDataSmbShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UnstructuredDataSmbShareResourceModel
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

	spec := client.CreateSmbShareServerSpec{
		Type: "SMBShare",
		Path: plan.Path.ValueString(),
		Processing: client.UnstructuredDataProcessingModel{
			CacheRepositoryID: cacheRepoID,
			BackupProxies: client.BackupProxiesSettingsModel{
				AutoSelectEnabled: true,
			},
		},
	}

	if !plan.CredentialsID.IsNull() && plan.CredentialsID.ValueString() != "" {
		spec.AccessCredentialsRequired = true
		spec.AccessCredentialsID = plan.CredentialsID.ValueString()
	}

	created, err := r.client.CreateUnstructuredDataServer(ctx, spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating SMB Share Source", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.CacheRepositoryID = types.StringValue(cacheRepoID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UnstructuredDataSmbShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UnstructuredDataSmbShareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	srv, err := r.client.GetUnstructuredDataServerByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading SMB Share Source", err.Error())
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

func (r *UnstructuredDataSmbShareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"SMB Share source modification is not directly supported; replace resource instead.",
	)
}

func (r *UnstructuredDataSmbShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UnstructuredDataSmbShareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUnstructuredDataServer(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting SMB Share Source", err.Error())
		return
	}
}

func (r *UnstructuredDataSmbShareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
