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

var _ resource.Resource = &RepositoryWindowsResource{}
var _ resource.ResourceWithImportState = &RepositoryWindowsResource{}

type RepositoryWindowsResource struct {
	client *client.Client
}

type RepositoryWindowsResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	HostID      types.String `tfsdk:"host_id"`
	Path        types.String `tfsdk:"path"`
}

func NewRepositoryWindowsResource() resource.Resource {
	return &RepositoryWindowsResource{}
}

func (r *RepositoryWindowsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_windows"
}

func (r *RepositoryWindowsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Microsoft Windows Local Backup Repository in Veeam Backup & Replication.",
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
				Description: "GUID of the managed Windows server (veeam_managed_server_windows.id).",
			},
			"path": schema.StringAttribute{
				Required:    true,
				Description: "Windows directory path where backup files are stored (e.g. D:\\Backups).",
			},
		},
	}
}

func (r *RepositoryWindowsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RepositoryWindowsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RepositoryWindowsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spec := client.CreateWindowsLocalRepositorySpec{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Type:        "WinLocal",
		HostID:      plan.HostID.ValueString(),
		Repository: client.WindowsLocalRepositorySettings{
			Path: plan.Path.ValueString(),
		},
	}

	created, err := r.client.CreateRepository(ctx, spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Windows Backup Repository", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RepositoryWindowsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RepositoryWindowsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repo, err := r.client.GetRepositoryByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Windows Backup Repository", err.Error())
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

func (r *RepositoryWindowsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Repository modification is not directly supported; replace resource instead.",
	)
}

func (r *RepositoryWindowsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RepositoryWindowsResourceModel
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

func (r *RepositoryWindowsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
