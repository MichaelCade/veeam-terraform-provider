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

var _ resource.Resource = &CredentialResource{}
var _ resource.ResourceWithImportState = &CredentialResource{}

type CredentialResource struct {
	client *client.Client
}

type CredentialResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	Description        types.String `tfsdk:"description"`
	Type               types.String `tfsdk:"type"`
	AuthenticationType types.String `tfsdk:"authentication_type"`
}

func NewCredentialResource() resource.Resource {
	return &CredentialResource{}
}

func (r *CredentialResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

func (r *CredentialResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages stored credentials within Veeam Backup & Replication Credential Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier (GUID) of the credential in Veeam.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "Username or Access Key ID for the credential entry.",
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Password or Secret Access Key for the credential entry.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the credential entry.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Credential type (Standard, Linux, ManagedService).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"authentication_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Authentication type for Linux credentials (Password or SshKey). Defaults to Password for Linux type.",
			},
		},
	}
}

func (r *CredentialResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CredentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	authType := plan.AuthenticationType.ValueString()
	if plan.Type.ValueString() == "Linux" && authType == "" {
		authType = "Password"
	}

	created, err := r.client.CreateCredential(ctx, client.Credential{
		Username:           plan.Username.ValueString(),
		Password:           plan.Password.ValueString(),
		Description:        plan.Description.ValueString(),
		Type:               plan.Type.ValueString(),
		AuthenticationType: authType,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Veeam Credential", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	if created.AuthenticationType != "" {
		plan.AuthenticationType = types.StringValue(created.AuthenticationType)
	} else {
		plan.AuthenticationType = types.StringValue(authType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CredentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	creds, err := r.client.GetCredentials(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Veeam Credentials", err.Error())
		return
	}

	var found *client.Credential
	for _, c := range creds {
		if c.ID == state.ID.ValueString() {
			found = &c
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Username = types.StringValue(found.Username)
	state.Description = types.StringValue(found.Description)
	state.Type = types.StringValue(found.Type)
	if found.AuthenticationType != "" {
		state.AuthenticationType = types.StringValue(found.AuthenticationType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CredentialResourceModel
	var state CredentialResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	authType := plan.AuthenticationType.ValueString()
	if plan.Type.ValueString() == "Linux" && authType == "" {
		authType = "Password"
	}

	updated, err := r.client.UpdateCredential(ctx, state.ID.ValueString(), client.Credential{
		Username:           plan.Username.ValueString(),
		Password:           plan.Password.ValueString(),
		Description:        plan.Description.ValueString(),
		Type:               plan.Type.ValueString(),
		AuthenticationType: authType,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Veeam Credential", err.Error())
		return
	}

	plan.ID = types.StringValue(updated.ID)
	if updated.AuthenticationType != "" {
		plan.AuthenticationType = types.StringValue(updated.AuthenticationType)
	} else {
		plan.AuthenticationType = types.StringValue(authType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CredentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCredential(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Veeam Credential", err.Error())
		return
	}
}

func (r *CredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
