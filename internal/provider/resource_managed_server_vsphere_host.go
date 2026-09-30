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

var _ resource.Resource = &ManagedServerVSphereResource{}
var _ resource.ResourceWithImportState = &ManagedServerVSphereResource{}

type ManagedServerVSphereResource struct {
	client *client.Client
}

type ManagedServerVSphereResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Description           types.String `tfsdk:"description"`
	CredentialsID         types.String `tfsdk:"credentials_id"`
	Port                  types.Int64  `tfsdk:"port"`
	CertificateThumbprint types.String `tfsdk:"certificate_thumbprint"`
}

func NewManagedServerVSphereResource() resource.Resource {
	return &ManagedServerVSphereResource{}
}

func (r *ManagedServerVSphereResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_server_vsphere"
}

func (r *ManagedServerVSphereResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a VMware vSphere vCenter server or standalone ESXi host added to Veeam Backup Infrastructure.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier (GUID) of the managed vSphere server in Veeam.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "DNS name or IP address of the vCenter server or ESXi host.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the managed vSphere server.",
			},
			"credentials_id": schema.StringAttribute{
				Required:    true,
				Description: "ID (GUID) of stored Veeam credentials used to connect to vSphere/vCenter.",
			},
			"port": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Port used to connect to vSphere web API (default: 443).",
			},
			"certificate_thumbprint": schema.StringAttribute{
				Optional:    true,
				Description: "TLS certificate thumbprint used to verify the vCenter server identity.",
			},
		},
	}
}

func (r *ManagedServerVSphereResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ManagedServerVSphereResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ManagedServerVSphereResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	portVal := int(plan.Port.ValueInt64())
	if portVal <= 0 {
		portVal = 443
	}

	spec := client.CreateViHostManagedServerSpec{
		Name:                  plan.Name.ValueString(),
		Description:           plan.Description.ValueString(),
		Type:                  "ViHost",
		CredentialsID:         plan.CredentialsID.ValueString(),
		Port:                  portVal,
		CertificateThumbprint: plan.CertificateThumbprint.ValueString(),
	}

	created, err := r.client.CreateManagedServer(ctx, spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Registering vSphere Managed Server", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.Port = types.Int64Value(int64(portVal))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ManagedServerVSphereResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ManagedServerVSphereResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.GetManagedServerByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading vSphere Managed Server", err.Error())
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

func (r *ManagedServerVSphereResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Managed server modification is not directly supported; replace resource instead.",
	)
}

func (r *ManagedServerVSphereResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ManagedServerVSphereResourceModel
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

func (r *ManagedServerVSphereResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
