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

var _ resource.Resource = &RepositoryVeeamVaultResource{}
var _ resource.ResourceWithImportState = &RepositoryVeeamVaultResource{}

type RepositoryVeeamVaultResource struct {
	client *client.Client
}

type RepositoryVeeamVaultResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	VaultID            types.String `tfsdk:"vault_id"`
	Folder             types.String `tfsdk:"folder"`
	EnableImmutability types.Bool   `tfsdk:"enable_immutability"`
}

func NewRepositoryVeeamVaultResource() resource.Resource {
	return &RepositoryVeeamVaultResource{}
}

func (r *RepositoryVeeamVaultResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_veeam_vault"
}

func (r *RepositoryVeeamVaultResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Veeam Data Cloud Vault Backup Repository in Veeam Backup & Replication.",
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
				Description: "Name of the Veeam Data Cloud Vault backup repository.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the backup repository.",
			},
			"vault_id": schema.StringAttribute{
				Required:    true,
				Description: "ID / GUID of the registered Veeam Data Cloud Vault account.",
			},
			"folder": schema.StringAttribute{
				Required:    true,
				Description: "Folder path inside Veeam Data Cloud Vault to map as repository.",
			},
			"enable_immutability": schema.BoolAttribute{
				Optional:    true,
				Description: "Set to true to enable Object Lock immutability on Veeam Data Cloud Vault container.",
			},
		},
	}
}

func (r *RepositoryVeeamVaultResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RepositoryVeeamVaultResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RepositoryVeeamVaultResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spec := client.CreateVeeamDataCloudVaultRepositorySpec{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Type:        "VeeamDataCloudVault",
		Account: client.VeeamDataCloudAccountSettings{
			Vault: client.VeeamDataCloudVaultInfo{
				ID: plan.VaultID.ValueString(),
			},
		},
		Container: client.VeeamDataCloudContainerSettings{
			Folder: plan.Folder.ValueString(),
			Immutability: client.ObjectStorageImmutabilitySettings{
				IsEnabled: plan.EnableImmutability.ValueBool(),
			},
		},
	}

	created, err := r.client.CreateRepository(ctx, spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Veeam Data Cloud Vault Repository", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RepositoryVeeamVaultResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RepositoryVeeamVaultResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repo, err := r.client.GetRepositoryByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Veeam Data Cloud Vault Repository", err.Error())
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

func (r *RepositoryVeeamVaultResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Repository modification is not directly supported; replace resource instead.",
	)
}

func (r *RepositoryVeeamVaultResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RepositoryVeeamVaultResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRepository(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Veeam Data Cloud Vault Repository", err.Error())
		return
	}
}

func (r *RepositoryVeeamVaultResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
