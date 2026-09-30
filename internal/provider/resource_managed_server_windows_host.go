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

var _ resource.Resource = &ManagedServerWindowsResource{}
var _ resource.ResourceWithImportState = &ManagedServerWindowsResource{}

type ManagedServerWindowsResource struct {
	client *client.Client
}

type ManagedServerWindowsResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	CredentialsID types.String `tfsdk:"credentials_id"`
}

func NewManagedServerWindowsResource() resource.Resource {
	return &ManagedServerWindowsResource{}
}

func (r *ManagedServerWindowsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_server_windows"
}

func (r *ManagedServerWindowsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Microsoft Windows host added to Veeam Backup Infrastructure.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier (GUID) of the managed Windows server in Veeam.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "DNS name or IP address of the Windows server.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the managed Windows server.",
			},
			"credentials_id": schema.StringAttribute{
				Required:    true,
				Description: "ID (GUID) of stored Veeam Windows credentials entry.",
			},
		},
	}
}

func (r *ManagedServerWindowsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ManagedServerWindowsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ManagedServerWindowsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spec := client.CreateWindowsManagedServerSpec{
		Name:                   plan.Name.ValueString(),
		Description:            plan.Description.ValueString(),
		Type:                   "WindowsHost",
		CredentialsStorageType: "Stored",
		CredentialsID:          plan.CredentialsID.ValueString(),
	}

	created, err := r.client.CreateManagedServer(ctx, spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Registering Windows Managed Server", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ManagedServerWindowsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ManagedServerWindowsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.GetManagedServerByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Windows Managed Server", err.Error())
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

func (r *ManagedServerWindowsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Managed server modification is not directly supported; replace resource instead.",
	)
}

func (r *ManagedServerWindowsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ManagedServerWindowsResourceModel
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

func (r *ManagedServerWindowsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
