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

var _ resource.Resource = &RepositoryNfsResource{}
var _ resource.ResourceWithImportState = &RepositoryNfsResource{}

type RepositoryNfsResource struct {
	client *client.Client
}

type RepositoryNfsResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	SharePath   types.String `tfsdk:"share_path"`
}

func NewRepositoryNfsResource() resource.Resource {
	return &RepositoryNfsResource{}
}

func (r *RepositoryNfsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_nfs"
}

func (r *RepositoryNfsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an NFS Share Backup Repository in Veeam Backup & Replication.",
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
			"share_path": schema.StringAttribute{
				Required:    true,
				Description: "NFS share path (e.g. nas.lab.local:/export/backups).",
			},
		},
	}
}

func (r *RepositoryNfsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RepositoryNfsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RepositoryNfsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spec := client.CreateNfsRepositorySpec{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Type:        "Nfs",
		Share: client.NfsShareSettings{
			SharePath: plan.SharePath.ValueString(),
			GatewayServer: client.RepositoryShareGatewayModel{
				AutoSelectEnabled: true,
			},
		},
	}

	created, err := r.client.CreateRepository(ctx, spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating NFS Backup Repository", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RepositoryNfsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RepositoryNfsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repo, err := r.client.GetRepositoryByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading NFS Backup Repository", err.Error())
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

func (r *RepositoryNfsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Repository modification is not directly supported; replace resource instead.",
	)
}

func (r *RepositoryNfsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RepositoryNfsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRepository(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting NFS Backup Repository", err.Error())
		return
	}
}

func (r *RepositoryNfsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
