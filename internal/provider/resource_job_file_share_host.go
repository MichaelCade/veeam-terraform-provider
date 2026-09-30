package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/michael/terraform-provider-veeam/pkg/client"
)

var _ resource.Resource = &JobFileShareResource{}
var _ resource.ResourceWithImportState = &JobFileShareResource{}

type JobFileShareResource struct {
	client *client.Client
}

type FileShareItemModel struct {
	Type         types.String `tfsdk:"type"`
	FileServerID types.String `tfsdk:"file_server_id"`
	Path         types.String `tfsdk:"path"`
}

type JobFileShareResourceModel struct {
	ID                types.String         `tfsdk:"id"`
	Name              types.String         `tfsdk:"name"`
	Description       types.String         `tfsdk:"description"`
	RepositoryID      types.String         `tfsdk:"repository_id"`
	RetentionQuantity types.Int64          `tfsdk:"retention_quantity"`
	RetentionType     types.String         `tfsdk:"retention_type"`
	Objects           []FileShareItemModel `tfsdk:"objects"`
	Schedule          *ScheduleModel       `tfsdk:"schedule"`
}

func NewJobFileShareResource() resource.Resource {
	return &JobFileShareResource{}
}

func (r *JobFileShareResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job_file_share"
}

func (r *JobFileShareResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a NAS / File Share Backup Job in Veeam Backup & Replication.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier (GUID) of the file share backup job in Veeam.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the file share backup job.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the file share backup job.",
			},
			"repository_id": schema.StringAttribute{
				Required:    true,
				Description: "GUID of the target backup repository.",
			},
			"retention_quantity": schema.Int64Attribute{
				Optional:    true,
				Description: "Retention period quantity (number of Days or Months to keep versions, default: 28).",
			},
			"retention_type": schema.StringAttribute{
				Optional:    true,
				Description: "Retention period type (Days or Months, default: Days).",
			},
			"objects": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of file share objects or server sources to back up.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Optional:    true,
							Description: "Type of file backup object (e.g. FileServerShare, FileServerRoot, Directory, File). Defaults to FileServerShare.",
						},
						"file_server_id": schema.StringAttribute{
							Optional:    true,
							Description: "GUID of the unstructured data server (SMB or NFS share source). Auto-resolved from path if omitted.",
						},
						"path": schema.StringAttribute{
							Optional:    true,
							Description: "File share path (UNC path or NFS share).",
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"schedule": schema.SingleNestedBlock{
				Description: "Job scheduling configuration.",
				Attributes: map[string]schema.Attribute{
					"run_automatically": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "If true, automatic scheduling is enabled for this job.",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
				},
				Blocks: map[string]schema.Block{
					"daily": schema.SingleNestedBlock{
						Description: "Daily schedule configuration.",
						Attributes: map[string]schema.Attribute{
							"is_enabled": schema.BoolAttribute{
								Optional:    true,
								Computed:    true,
								Description: "If true, daily schedule is enabled.",
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.UseStateForUnknown(),
								},
							},
							"local_time": schema.StringAttribute{
								Optional:    true,
								Computed:    true,
								Description: "Time of day to execute backup (HH:MM). Defaults to 22:00.",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"daily_kind": schema.StringAttribute{
								Optional:    true,
								Computed:    true,
								Description: "Kind of daily schedule: Everyday, WeekDays, SelectedDays.",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"days": schema.ListAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "Days of week for SelectedDays daily kind.",
							},
						},
					},
				},
			},
		},
	}
}

func (r *JobFileShareResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildScheduleSpec(planSchedule *ScheduleModel) client.BackupScheduleModel {
	var spec client.BackupScheduleModel
	if planSchedule == nil {
		spec.RunAutomatically = false
		return spec
	}

	runAuto := true
	if !planSchedule.RunAutomatically.IsNull() && !planSchedule.RunAutomatically.IsUnknown() {
		runAuto = planSchedule.RunAutomatically.ValueBool()
	}
	spec.RunAutomatically = runAuto
	planSchedule.RunAutomatically = types.BoolValue(runAuto)

	if planSchedule.Daily != nil {
		dailyEnabled := true
		if !planSchedule.Daily.IsEnabled.IsNull() && !planSchedule.Daily.IsEnabled.IsUnknown() {
			dailyEnabled = planSchedule.Daily.IsEnabled.ValueBool()
		}

		localTime := planSchedule.Daily.LocalTime.ValueString()
		if localTime == "" {
			localTime = "22:00"
		}

		dailyKind := planSchedule.Daily.DailyKind.ValueString()
		if dailyKind == "" {
			dailyKind = "Everyday"
		}

		var days []string
		for _, d := range planSchedule.Daily.Days {
			if d.ValueString() != "" {
				days = append(days, d.ValueString())
			}
		}

		spec.Daily = &client.ScheduleDailyModel{
			IsEnabled: dailyEnabled,
			LocalTime: localTime,
			DailyKind: dailyKind,
			Days:      days,
		}

		planSchedule.Daily.IsEnabled = types.BoolValue(dailyEnabled)
		planSchedule.Daily.LocalTime = types.StringValue(localTime)
		planSchedule.Daily.DailyKind = types.StringValue(dailyKind)
	}

	return spec
}

func (r *JobFileShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan JobFileShareResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	retentionQty := int(plan.RetentionQuantity.ValueInt64())
	if retentionQty <= 0 {
		retentionQty = 28
	}
	retentionType := plan.RetentionType.ValueString()
	if retentionType == "" {
		retentionType = "Days"
	}

	objects := make([]client.FileBackupItemModel, 0, len(plan.Objects))
	for _, obj := range plan.Objects {
		objType := obj.Type.ValueString()
		if objType == "" {
			objType = "FileServerShare"
		}

		fileServerID := obj.FileServerID.ValueString()
		filePath := obj.Path.ValueString()

		// 1. If fileServerID is empty but path is given, auto-lookup fileServerID from inventory
		if fileServerID == "" && filePath != "" {
			if servers, err := r.client.GetUnstructuredDataServers(ctx); err == nil {
				for _, s := range servers {
					if s.Path == filePath || s.Name == filePath {
						fileServerID = s.ID
						break
					}
				}
			}
		}

		// 2. If path is empty but fileServerID is given, auto-lookup path from inventory
		if filePath == "" && fileServerID != "" {
			if s, err := r.client.GetUnstructuredDataServerByID(ctx, fileServerID); err == nil && s != nil {
				filePath = s.Path
				if filePath == "" {
					filePath = s.Name
				}
			}
		}

		if fileServerID == "" {
			resp.Diagnostics.AddError(
				"Missing File Server ID",
				fmt.Sprintf("Could not resolve fileServerId for object path '%s'. Please specify 'file_server_id' or ensure the file share is registered via veeam_unstructured_data_smb_share or veeam_unstructured_data_nfs_share.", filePath),
			)
			return
		}

		if filePath == "" {
			resp.Diagnostics.AddError(
				"Missing Path",
				fmt.Sprintf("Path is required for file server ID '%s'. Please specify 'path' attribute in the object configuration.", fileServerID),
			)
			return
		}

		objects = append(objects, client.FileBackupItemModel{
			Type:         objType,
			FileServerID: fileServerID,
			Path:         filePath,
		})
	}

	scheduleSpec := buildScheduleSpec(plan.Schedule)

	spec := client.CreateFileShareJobSpec{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Type:        "FileBackup",
		Objects:     objects,
		BackupRepository: client.FileBackupPrimaryRepositoryModel{
			BackupRepositoryID: plan.RepositoryID.ValueString(),
			RetentionPolicy: &client.UnstructuredDataRetentionPolicySettingsModel{
				Type:     retentionType,
				Quantity: retentionQty,
			},
		},
		Schedule: scheduleSpec,
	}

	created, err := r.client.CreateFileShareJob(ctx, spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating File Share Backup Job", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.RetentionQuantity = types.Int64Value(int64(retentionQty))
	plan.RetentionType = types.StringValue(retentionType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *JobFileShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state JobFileShareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	job, err := r.client.GetJobByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading File Share Backup Job", err.Error())
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

	if job.Schedule != nil {
		if state.Schedule == nil {
			state.Schedule = &ScheduleModel{}
		}
		state.Schedule.RunAutomatically = types.BoolValue(job.Schedule.RunAutomatically)

		if job.Schedule.Daily != nil {
			if state.Schedule.Daily == nil {
				state.Schedule.Daily = &DailyScheduleModel{}
			}
			state.Schedule.Daily.IsEnabled = types.BoolValue(job.Schedule.Daily.IsEnabled)
			if job.Schedule.Daily.LocalTime != "" {
				state.Schedule.Daily.LocalTime = types.StringValue(job.Schedule.Daily.LocalTime)
			}
			if job.Schedule.Daily.DailyKind != "" {
				state.Schedule.Daily.DailyKind = types.StringValue(job.Schedule.Daily.DailyKind)
			}
			if len(job.Schedule.Daily.Days) > 0 {
				var days []types.String
				for _, d := range job.Schedule.Daily.Days {
					days = append(days, types.StringValue(d))
				}
				state.Schedule.Daily.Days = days
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *JobFileShareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan JobFileShareResourceModel
	var state JobFileShareResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	retentionQty := int(plan.RetentionQuantity.ValueInt64())
	if retentionQty <= 0 {
		retentionQty = 28
	}
	retentionType := plan.RetentionType.ValueString()
	if retentionType == "" {
		retentionType = "Days"
	}

	objects := make([]client.FileBackupItemModel, 0, len(plan.Objects))
	for _, obj := range plan.Objects {
		objType := obj.Type.ValueString()
		if objType == "" {
			objType = "FileServerShare"
		}

		fileServerID := obj.FileServerID.ValueString()
		filePath := obj.Path.ValueString()

		// 1. If fileServerID is empty but path is given, auto-lookup fileServerID from inventory
		if fileServerID == "" && filePath != "" {
			if servers, err := r.client.GetUnstructuredDataServers(ctx); err == nil {
				for _, s := range servers {
					if s.Path == filePath || s.Name == filePath {
						fileServerID = s.ID
						break
					}
				}
			}
		}

		// 2. If path is empty but fileServerID is given, auto-lookup path from inventory
		if filePath == "" && fileServerID != "" {
			if s, err := r.client.GetUnstructuredDataServerByID(ctx, fileServerID); err == nil && s != nil {
				filePath = s.Path
				if filePath == "" {
					filePath = s.Name
				}
			}
		}

		if fileServerID == "" {
			resp.Diagnostics.AddError(
				"Missing File Server ID",
				fmt.Sprintf("Could not resolve fileServerId for object path '%s'. Please specify 'file_server_id' or ensure the file share is registered via veeam_unstructured_data_smb_share or veeam_unstructured_data_nfs_share.", filePath),
			)
			return
		}

		if filePath == "" {
			resp.Diagnostics.AddError(
				"Missing Path",
				fmt.Sprintf("Path is required for file server ID '%s'. Please specify 'path' attribute in the object configuration.", fileServerID),
			)
			return
		}

		objects = append(objects, client.FileBackupItemModel{
			Type:         objType,
			FileServerID: fileServerID,
			Path:         filePath,
		})
	}

	scheduleSpec := buildScheduleSpec(plan.Schedule)

	spec := client.CreateFileShareJobSpec{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Type:        "FileBackup",
		Objects:     objects,
		BackupRepository: client.FileBackupPrimaryRepositoryModel{
			BackupRepositoryID: plan.RepositoryID.ValueString(),
			RetentionPolicy: &client.UnstructuredDataRetentionPolicySettingsModel{
				Type:     retentionType,
				Quantity: retentionQty,
			},
		},
		Schedule: scheduleSpec,
	}

	updated, err := r.client.UpdateFileShareJob(ctx, state.ID.ValueString(), spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating File Share Backup Job", err.Error())
		return
	}

	plan.ID = types.StringValue(updated.ID)
	plan.RetentionQuantity = types.Int64Value(int64(retentionQty))
	plan.RetentionType = types.StringValue(retentionType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *JobFileShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state JobFileShareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteJob(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting File Share Backup Job", err.Error())
		return
	}
}

func (r *JobFileShareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

