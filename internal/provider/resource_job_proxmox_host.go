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

var _ resource.Resource = &JobProxmoxResource{}
var _ resource.ResourceWithImportState = &JobProxmoxResource{}

type JobProxmoxResource struct {
	client *client.Client
}

type ProxmoxIncludeObjectModel struct {
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	ObjectID types.String `tfsdk:"object_id"`
}

type JobProxmoxResourceModel struct {
	ID                types.String                 `tfsdk:"id"`
	Name              types.String                 `tfsdk:"name"`
	Description       types.String                 `tfsdk:"description"`
	RepositoryID      types.String                 `tfsdk:"repository_id"`
	Includes          []ProxmoxIncludeObjectModel `tfsdk:"includes"`
	RetentionQuantity types.Int64                  `tfsdk:"retention_quantity"`
	RetentionType     types.String                 `tfsdk:"retention_type"`
}

func NewJobProxmoxResource() resource.Resource {
	return &JobProxmoxResource{}
}

func (r *JobProxmoxResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job_proxmox"
}

func (r *JobProxmoxResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Proxmox VE Backup Job in Veeam Backup & Replication.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier (GUID) of the Proxmox backup job in Veeam.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the Proxmox backup job.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the Proxmox backup job.",
			},
			"repository_id": schema.StringAttribute{
				Required:    true,
				Description: "GUID of the backup repository where backups will be stored.",
			},
			"retention_quantity": schema.Int64Attribute{
				Optional:    true,
				Description: "Number of restore points to retain (default: 7).",
			},
			"retention_type": schema.StringAttribute{
				Optional:    true,
				Description: "Retention policy unit (RestorePoints or Days, default: RestorePoints).",
			},
			"includes": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of Proxmox inventory objects (VMs, Nodes, Clusters) to back up.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the Proxmox VM, Node, or Cluster object.",
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "Proxmox inventory object type (VirtualMachine, ProxmoxNode, ProxmoxCluster).",
						},
						"object_id": schema.StringAttribute{
							Optional:    true,
							Description: "Unique object ID in Proxmox inventory.",
						},
					},
				},
			},
		},
	}
}

func (r *JobProxmoxResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *JobProxmoxResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan JobProxmoxResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	retentionQty := int(plan.RetentionQuantity.ValueInt64())
	if retentionQty <= 0 {
		retentionQty = 7
	}
	retentionType := plan.RetentionType.ValueString()
	if retentionType == "" {
		retentionType = "RestorePoints"
	}

	includes := make([]client.ProxmoxIncludeObject, 0, len(plan.Includes))
	for _, inc := range plan.Includes {
		includes = append(includes, client.ProxmoxIncludeObject{
			Name:     inc.Name.ValueString(),
			Type:     inc.Type.ValueString(),
			ObjectID: inc.ObjectID.ValueString(),
		})
	}

	spec := client.CreateProxmoxJobSpec{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Type:        "ProxmoxBackupJob",
		Proxmox: client.ProxmoxScopeModel{
			Includes: includes,
		},
		Storage: client.BackupJobStorageModel{
			BackupRepositoryID: plan.RepositoryID.ValueString(),
			RetentionPolicy: client.BackupJobRetentionPolicySettingsModel{
				Type:     retentionType,
				Quantity: retentionQty,
			},
		},
		Schedule: client.BackupScheduleModel{
			RunAutomatically: false,
		},
	}

	created, err := r.client.CreateProxmoxJob(ctx, spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Proxmox Backup Job", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.RetentionQuantity = types.Int64Value(int64(retentionQty))
	plan.RetentionType = types.StringValue(retentionType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *JobProxmoxResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state JobProxmoxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	job, err := r.client.GetJobByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Proxmox Backup Job", err.Error())
		return
	}

	if job == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(job.Name)
	if job.Description != "" {
		state.Description = types.StringValue(job.Description)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *JobProxmoxResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Job modification is not directly supported; replace resource instead.",
	)
}

func (r *JobProxmoxResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state JobProxmoxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteJob(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Proxmox Backup Job", err.Error())
		return
	}
}

func (r *JobProxmoxResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
