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

var _ resource.Resource = &RepositoryS3CompatibleResource{}
var _ resource.ResourceWithImportState = &RepositoryS3CompatibleResource{}

type RepositoryS3CompatibleResource struct {
	client *client.Client
}

type RepositoryS3CompatibleResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	ServiceEndpoint types.String `tfsdk:"service_endpoint"`
	RegionID        types.String `tfsdk:"region_id"`
	BucketName      types.String `tfsdk:"bucket_name"`
	FolderName      types.String `tfsdk:"folder_name"`
	CredentialsID   types.String `tfsdk:"credentials_id"`
}

func NewRepositoryS3CompatibleResource() resource.Resource {
	return &RepositoryS3CompatibleResource{}
}

func (r *RepositoryS3CompatibleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_s3_compatible"
}

func (r *RepositoryS3CompatibleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an S3 Compatible Object Storage Backup Repository in Veeam Backup & Replication.",
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
				Description: "Name of the object storage backup repository.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the backup repository.",
			},
			"service_endpoint": schema.StringAttribute{
				Required:    true,
				Description: "Endpoint URL of S3 compatible storage (e.g. s3.lab.local:9000).",
			},
			"region_id": schema.StringAttribute{
				Optional:    true,
				Description: "Region ID of S3 compatible storage (default: us-east-1).",
			},
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "S3 Bucket name.",
			},
			"folder_name": schema.StringAttribute{
				Required:    true,
				Description: "Folder path inside the S3 bucket to map as the backup repository.",
			},
			"credentials_id": schema.StringAttribute{
				Required:    true,
				Description: "GUID of Veeam S3 Access/Secret key credential.",
			},
		},
	}
}

func (r *RepositoryS3CompatibleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RepositoryS3CompatibleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RepositoryS3CompatibleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region := plan.RegionID.ValueString()
	if region == "" {
		region = "us-east-1"
	}

	spec := client.CreateS3CompatibleRepositorySpec{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Type:        "S3Compatible",
		Account: client.S3CompatibleAccountSettings{
			ServicePoint:  plan.ServiceEndpoint.ValueString(),
			RegionID:      region,
			CredentialsID: plan.CredentialsID.ValueString(),
		},
		Bucket: client.S3CompatibleBucketSettings{
			BucketName: plan.BucketName.ValueString(),
			FolderName: plan.FolderName.ValueString(),
		},
	}

	created, err := r.client.CreateRepository(ctx, spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating S3 Compatible Backup Repository", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	if plan.RegionID.IsNull() || plan.RegionID.IsUnknown() {
		plan.RegionID = types.StringValue(region)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RepositoryS3CompatibleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RepositoryS3CompatibleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repo, err := r.client.GetRepositoryByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading S3 Compatible Backup Repository", err.Error())
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

func (r *RepositoryS3CompatibleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Repository modification is not directly supported; replace resource instead.",
	)
}

func (r *RepositoryS3CompatibleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RepositoryS3CompatibleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRepository(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting S3 Compatible Backup Repository", err.Error())
		return
	}
}

func (r *RepositoryS3CompatibleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
