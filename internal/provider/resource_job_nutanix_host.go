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

var _ resource.Resource = &JobNutanixResource{}
var _ resource.ResourceWithImportState = &JobNutanixResource{}

type JobNutanixResource struct {
	client *client.Client
}

type NutanixIncludeObjectModel struct {
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	ObjectID types.String `tfsdk:"object_id"`
}

type JobNutanixResourceModel struct {
	ID                types.String                  `tfsdk:"id"`
	Name              types.String                  `tfsdk:"name"`
	Description       types.String                  `tfsdk:"description"`
	RepositoryID      types.String                  `tfsdk:"repository_id"`
	Includes          []NutanixIncludeObjectModel `tfsdk:"includes"`
	RetentionQuantity types.Int64                   `tfsdk:"retention_quantity"`
	RetentionType     types.String                  `tfsdk:"retention_type"`
}

func NewJobNutanixResource() resource.Resource {
	return &JobNutanixResource{}
}

func (r *JobNutanixResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job_nutanix"
}

func (r *JobNutanixResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Nutanix AHV Backup Job in Veeam Backup & Replication.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier (GUID) of the Nutanix AHV backup job in Veeam.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the Nutanix AHV backup job.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the Nutanix AHV backup job.",
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
				Description: "List of Nutanix AHV inventory objects (VMs, Clusters) to back up.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the Nutanix AHV VM or Cluster object.",
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "Nutanix inventory object type (VirtualMachine, Cluster).",
						},
						"object_id": schema.StringAttribute{
							Optional:    true,
							Description: "Unique object ID in Nutanix inventory.",
						},
					},
				},
			},
		},
	}
}

func (r *JobNutanixResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *JobNutanixResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan JobNutanixResourceModel
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

	includes := make([]client.NutanixIncludeObject, 0, len(plan.Includes))
	for _, inc := range plan.Includes {
		includes = append(includes, client.NutanixIncludeObject{
			Name:     inc.Name.ValueString(),
			Type:     inc.Type.ValueString(),
			ObjectID: inc.ObjectID.ValueString(),
		})
	}

	spec := client.CreateNutanixJobSpec{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Type:        "NutanixAHVBackupJob",
		Nutanix: client.NutanixScopeModel{
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

	created, err := r.client.CreateNutanixJob(ctx, spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Nutanix AHV Backup Job", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.RetentionQuantity = types.Int64Value(int64(retentionQty))
	plan.RetentionType = types.StringValue(retentionType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *JobNutanixResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state JobNutanixResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	job, err := r.client.GetJobByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Nutanix AHV Backup Job", err.Error())
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

func (r *JobNutanixResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Job modification is not directly supported; replace resource instead.",
	)
}

func (r *JobNutanixResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state JobNutanixResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteJob(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Nutanix AHV Backup Job", err.Error())
		return
	}
}

func (r *JobNutanixResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
