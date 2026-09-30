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

var _ resource.Resource = &ManagedServerProxmoxResource{}
var _ resource.ResourceWithImportState = &ManagedServerProxmoxResource{}

type ManagedServerProxmoxResource struct {
	client *client.Client
}

type ManagedServerProxmoxResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	ServerType     types.String `tfsdk:"server_type"`
	CredentialsID  types.String `tfsdk:"credentials_id"`
	SSHFingerprint types.String `tfsdk:"ssh_fingerprint"`
	SSHPort        types.Int64  `tfsdk:"ssh_port"`
}

func NewManagedServerProxmoxResource() resource.Resource {
	return &ManagedServerProxmoxResource{}
}

func (r *ManagedServerProxmoxResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_server_proxmox"
}

func (r *ManagedServerProxmoxResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Proxmox VE node or cluster added to Veeam Backup Infrastructure.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier (GUID) of the managed Proxmox VE server in Veeam.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "DNS name or IP address of the Proxmox VE node or cluster.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the managed Proxmox VE server.",
			},
			"server_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Type of Proxmox server: 'ProxmoxNode' or 'ProxmoxCluster'. Default: 'ProxmoxNode'.",
			},
			"credentials_id": schema.StringAttribute{
				Required:    true,
				Description: "ID (GUID) of stored Veeam SSH / root credentials entry.",
			},
			"ssh_fingerprint": schema.StringAttribute{
				Optional:    true,
				Description: "SSH host key fingerprint for server identity validation.",
			},
			"ssh_port": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "SSH port used to connect to Proxmox VE (default: 22).",
			},
		},
	}
}

func (r *ManagedServerProxmoxResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ManagedServerProxmoxResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ManagedServerProxmoxResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sType := plan.ServerType.ValueString()
	if sType == "" {
		sType = "ProxmoxNode"
	}

	portVal := int(plan.SSHPort.ValueInt64())
	if portVal <= 0 {
		portVal = 22
	}

	spec := client.CreateProxmoxManagedServerSpec{
		Name:                   plan.Name.ValueString(),
		Description:            plan.Description.ValueString(),
		Type:                   sType,
		CredentialsStorageType: "Stored",
		CredentialsID:          plan.CredentialsID.ValueString(),
		SSHFingerprint:         plan.SSHFingerprint.ValueString(),
		SSHSettings: &client.LinuxHostSSHSettingsModel{
			SSHPort: portVal,
		},
	}

	created, err := r.client.CreateManagedServer(ctx, spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Registering Proxmox VE Managed Server", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.ServerType = types.StringValue(sType)
	plan.SSHPort = types.Int64Value(int64(portVal))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ManagedServerProxmoxResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ManagedServerProxmoxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.GetManagedServerByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Proxmox VE Managed Server", err.Error())
		return
	}

	if server == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(server.Name)
	if server.Description != "" {
		state.Description = types.StringValue(server.Description)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ManagedServerProxmoxResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Managed server modification is not directly supported; replace resource instead.",
	)
}

func (r *ManagedServerProxmoxResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ManagedServerProxmoxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteManagedServer(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Managed Server", err.Error())
		return
	}
}

func (r *ManagedServerProxmoxResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
