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

var _ resource.Resource = &JobHyperVResource{}
var _ resource.ResourceWithImportState = &JobHyperVResource{}

type JobHyperVResource struct {
	client *client.Client
}

type HyperVIncludeObjectModel struct {
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	HostName types.String `tfsdk:"host_name"`
	ObjectID types.String `tfsdk:"object_id"`
}

type JobHyperVResourceModel struct {
	ID                types.String               `tfsdk:"id"`
	Name              types.String               `tfsdk:"name"`
	Description       types.String               `tfsdk:"description"`
	RepositoryID      types.String               `tfsdk:"repository_id"`
	Includes          []HyperVIncludeObjectModel `tfsdk:"includes"`
	RetentionQuantity types.Int64                `tfsdk:"retention_quantity"`
	RetentionType     types.String               `tfsdk:"retention_type"`
}

func NewJobHyperVResource() resource.Resource {
	return &JobHyperVResource{}
}

func (r *JobHyperVResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job_hyperv"
}

func (r *JobHyperVResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Microsoft Hyper-V Backup Job in Veeam Backup & Replication.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier (GUID) of the Hyper-V backup job in Veeam.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the Hyper-V backup job.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the Hyper-V backup job.",
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
				Description: "List of Hyper-V inventory objects (VMs, Hosts, Clusters) to back up.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the Hyper-V VM or host object.",
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "Hyper-V inventory object type (e.g. VirtualMachine, Host, Cluster).",
						},
						"host_name": schema.StringAttribute{
							Optional:    true,
							Description: "Hyper-V Host or Cluster name.",
						},
						"object_id": schema.StringAttribute{
							Optional:    true,
							Description: "Unique object ID in Hyper-V inventory.",
						},
					},
				},
			},
		},
	}
}

func (r *JobHyperVResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *JobHyperVResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan JobHyperVResourceModel
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

	includes := make([]client.HyperVIncludeObject, 0, len(plan.Includes))
	for _, inc := range plan.Includes {
		includes = append(includes, client.HyperVIncludeObject{
			Platform: "HyperV",
			Name:     inc.Name.ValueString(),
			Type:     inc.Type.ValueString(),
			HostName: inc.HostName.ValueString(),
			ObjectID: inc.ObjectID.ValueString(),
		})
	}

	spec := client.CreateHyperVJobSpec{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Type:        "HyperVBackup",
		VirtualMachines: client.HyperVVirtualMachinesSpec{
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

	created, err := r.client.CreateHyperVJob(ctx, spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Hyper-V Backup Job", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.RetentionQuantity = types.Int64Value(int64(retentionQty))
	plan.RetentionType = types.StringValue(retentionType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *JobHyperVResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state JobHyperVResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	job, err := r.client.GetJobByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Hyper-V Backup Job", err.Error())
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

func (r *JobHyperVResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan JobHyperVResourceModel
	var state JobHyperVResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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

	includes := make([]client.HyperVIncludeObject, 0, len(plan.Includes))
	for _, inc := range plan.Includes {
		includes = append(includes, client.HyperVIncludeObject{
			Platform: "HyperV",
			Name:     inc.Name.ValueString(),
			Type:     inc.Type.ValueString(),
			HostName: inc.HostName.ValueString(),
			ObjectID: inc.ObjectID.ValueString(),
		})
	}

	spec := client.CreateHyperVJobSpec{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Type:        "HyperVBackup",
		VirtualMachines: client.HyperVVirtualMachinesSpec{
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

	updated, err := r.client.UpdateHyperVJob(ctx, state.ID.ValueString(), spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Hyper-V Backup Job", err.Error())
		return
	}

	plan.ID = types.StringValue(updated.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *JobHyperVResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state JobHyperVResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteJob(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Hyper-V Backup Job", err.Error())
		return
	}
}

func (r *JobHyperVResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
