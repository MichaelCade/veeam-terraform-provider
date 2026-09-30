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

var _ resource.Resource = &RepositoryLinuxResource{}
var _ resource.ResourceWithImportState = &RepositoryLinuxResource{}

type RepositoryLinuxResource struct {
	client *client.Client
}

type RepositoryLinuxResourceModel struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Description                types.String `tfsdk:"description"`
	HostID                     types.String `tfsdk:"host_id"`
	Path                       types.String `tfsdk:"path"`
	UseFastCloning             types.Bool   `tfsdk:"use_fast_cloning"`
	EnableGovernanceMode       types.Bool   `tfsdk:"enable_governance_mode"`
	GovernanceRetentionDays    types.Int64  `tfsdk:"governance_retention_days"`
}

func NewRepositoryLinuxResource() resource.Resource {
	return &RepositoryLinuxResource{}
}

func (r *RepositoryLinuxResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_linux"
}

func (r *RepositoryLinuxResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Linux-based Local or Hardened Backup Repository in Veeam Backup & Replication.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier (GUID) of the repository in Veeam.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the backup repository.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the backup repository.",
			},
			"host_id": schema.StringAttribute{
				Required:    true,
				Description: "GUID of the managed Linux server (veeam_managed_server_linux.id).",
			},
			"path": schema.StringAttribute{
				Required:    true,
				Description: "Linux directory path where backup files are stored (e.g. /mnt/backup).",
			},
			"use_fast_cloning": schema.BoolAttribute{
				Optional:    true,
				Description: "Set to true to enable fast cloning on XFS volumes.",
			},
			"enable_governance_mode": schema.BoolAttribute{
				Optional:    true,
				Description: "Set to true to enable governance mode for Linux Hardened Repository immutability.",
			},
			"governance_retention_days": schema.Int64Attribute{
				Optional:    true,
				Description: "Number of days to keep backups immutable when governance mode is enabled.",
			},
		},
	}
}

func (r *RepositoryLinuxResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RepositoryLinuxResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RepositoryLinuxResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repoType := "LinuxLocal"
	if plan.EnableGovernanceMode.ValueBool() {
		repoType = "LinuxHardened"
	}

	spec := client.CreateLinuxLocalRepositorySpec{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Type:        repoType,
		HostID:      plan.HostID.ValueString(),
		Repository: client.LinuxLocalRepositorySettings{
			Path:                        plan.Path.ValueString(),
			UseFastCloningOnXFSVolumes: plan.UseFastCloning.ValueBool(),
			EnableGovernanceMode:        plan.EnableGovernanceMode.ValueBool(),
			GovernanceModeRetentionDays: int(plan.GovernanceRetentionDays.ValueInt64()),
		},
	}

	created, err := r.client.CreateRepository(ctx, spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Linux Backup Repository", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RepositoryLinuxResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RepositoryLinuxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repo, err := r.client.GetRepositoryByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Linux Backup Repository", err.Error())
		return
	}

	if repo == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(repo.Name)
	if repo.Description != "" {
		state.Description = types.StringValue(repo.Description)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RepositoryLinuxResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Repository modification is not directly supported; replace resource instead.",
	)
}

func (r *RepositoryLinuxResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RepositoryLinuxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRepository(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Backup Repository", err.Error())
		return
	}
}

func (r *RepositoryLinuxResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
