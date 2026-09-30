package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/michael/terraform-provider-veeam/pkg/client"
)

var _ resource.Resource = &JobVMwareResource{}
var _ resource.ResourceWithImportState = &JobVMwareResource{}

type JobVMwareResource struct {
	client *client.Client
}

type JobVMwareResourceModel struct {
	ID                types.String          `tfsdk:"id"`
	Name              types.String          `tfsdk:"name"`
	Description       types.String          `tfsdk:"description"`
	RepositoryID      types.String          `tfsdk:"repository_id"`
	VMwareHostName    types.String          `tfsdk:"vmware_host_name"`
	VMwareObjectType  types.String          `tfsdk:"vmware_object_type"`
	VMwareObjectID    types.String          `tfsdk:"vmware_object_id"`
	VMwareURN         types.String          `tfsdk:"vmware_urn"`
	Includes          []IncludeModel        `tfsdk:"includes"`
	RetentionType     types.String          `tfsdk:"retention_type"`
	RetentionQuantity types.Int64           `tfsdk:"retention_quantity"`
	Schedule          *ScheduleModel        `tfsdk:"schedule"`
	GFSPolicy         *GFSPolicyModel       `tfsdk:"gfs_policy"`
	GuestProcessing   *GuestProcessingModel `tfsdk:"guest_processing"`
}

type IncludeModel struct {
	Name     types.String `tfsdk:"name"`
	HostName types.String `tfsdk:"host_name"`
	Type     types.String `tfsdk:"type"`
	ObjectID types.String `tfsdk:"object_id"`
	URN      types.String `tfsdk:"urn"`
}

type ScheduleModel struct {
	RunAutomatically types.Bool          `tfsdk:"run_automatically"`
	Daily            *DailyScheduleModel `tfsdk:"daily"`
}

type DailyScheduleModel struct {
	IsEnabled types.Bool     `tfsdk:"is_enabled"`
	LocalTime types.String   `tfsdk:"local_time"`
	DailyKind types.String   `tfsdk:"daily_kind"`
	Days      []types.String `tfsdk:"days"`
}

type GFSPolicyModel struct {
	IsEnabled types.Bool       `tfsdk:"is_enabled"`
	Weekly    *GFSWeeklyModel  `tfsdk:"weekly"`
	Monthly   *GFSMonthlyModel `tfsdk:"monthly"`
	Yearly    *GFSYearlyModel  `tfsdk:"yearly"`
}

type GFSWeeklyModel struct {
	IsEnabled    types.Bool  `tfsdk:"is_enabled"`
	KeepForWeeks types.Int64 `tfsdk:"keep_for_weeks"`
}

type GFSMonthlyModel struct {
	IsEnabled     types.Bool  `tfsdk:"is_enabled"`
	KeepForMonths types.Int64 `tfsdk:"keep_for_months"`
}

type GFSYearlyModel struct {
	IsEnabled    types.Bool  `tfsdk:"is_enabled"`
	KeepForYears types.Int64 `tfsdk:"keep_for_years"`
}

type GuestProcessingModel struct {
	AppAwareProcessingEnabled types.Bool   `tfsdk:"app_aware_processing_enabled"`
	GuestIndexingEnabled       types.Bool   `tfsdk:"guest_indexing_enabled"`
	GuestCredentialsID         types.String `tfsdk:"guest_credentials_id"`
}

func NewJobVMwareResource() resource.Resource {
	return &JobVMwareResource{}
}

func (r *JobVMwareResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job_vmware"
}

func (r *JobVMwareResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a VMware vSphere Backup Job in Veeam Backup & Replication.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier (GUID) of the backup job in Veeam.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the VMware backup job.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the backup job.",
			},
			"repository_id": schema.StringAttribute{
				Required:    true,
				Description: "Target backup repository ID (GUID).",
			},
			"vmware_host_name": schema.StringAttribute{
				Optional:    true,
				Description: "Default hostname or IP address of the VMware vSphere server containing objects to protect.",
			},
			"vmware_object_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Convenience fallback type for single-target job (e.g. vCenterServer, Host, VirtualMachine, Tag).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vmware_object_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Convenience fallback object ID for single-target job.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vmware_urn": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Convenience fallback URN for single-target job.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"includes": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of specific VMware objects to include in the backup job (individual VMs, Tags, Categories, Folders).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the vSphere object, VM name, or Tag name.",
						},
						"host_name": schema.StringAttribute{
							Optional:    true,
							Description: "Hostname or IP address of the vCenter/vSphere server.",
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "Type of object (e.g. VirtualMachine, Tag, Category, Folder, vCenterServer).",
						},
						"object_id": schema.StringAttribute{
							Optional:    true,
							Description: "vSphere MOB ID or Tag ID (e.g. vm-101).",
						},
						"urn": schema.StringAttribute{
							Optional:    true,
							Description: "Uniform Resource Name (URN) for the vSphere object.",
						},
					},
				},
			},
			"retention_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Retention policy type (RestorePoints or Days). Defaults to RestorePoints.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"retention_quantity": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of restore points or days to keep. Defaults to 7.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
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
								Description: "Local time to start the job (format: HH:MM, e.g. 22:00).",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"daily_kind": schema.StringAttribute{
								Optional:    true,
								Computed:    true,
								Description: "Daily schedule kind: Everyday, WeekDays, or SelectedDays.",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"days": schema.ListAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "Days of the week when daily_kind is SelectedDays (e.g. ['Monday', 'Wednesday']).",
							},
						},
					},
				},
			},
			"gfs_policy": schema.SingleNestedBlock{
				Description: "Long-term (GFS) retention policy settings.",
				Attributes: map[string]schema.Attribute{
					"is_enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "If true, GFS retention policy is enabled.",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
				},
				Blocks: map[string]schema.Block{
					"weekly": schema.SingleNestedBlock{
						Description: "Weekly GFS retention policy.",
						Attributes: map[string]schema.Attribute{
							"is_enabled": schema.BoolAttribute{
								Optional:    true,
								Computed:    true,
								Description: "If true, weekly GFS backups are enabled.",
							},
							"keep_for_weeks": schema.Int64Attribute{
								Optional:    true,
								Description: "Number of weekly full backups to keep.",
							},
						},
					},
					"monthly": schema.SingleNestedBlock{
						Description: "Monthly GFS retention policy.",
						Attributes: map[string]schema.Attribute{
							"is_enabled": schema.BoolAttribute{
								Optional:    true,
								Computed:    true,
								Description: "If true, monthly GFS backups are enabled.",
							},
							"keep_for_months": schema.Int64Attribute{
								Optional:    true,
								Description: "Number of monthly full backups to keep.",
							},
						},
					},
					"yearly": schema.SingleNestedBlock{
						Description: "Yearly GFS retention policy.",
						Attributes: map[string]schema.Attribute{
							"is_enabled": schema.BoolAttribute{
								Optional:    true,
								Computed:    true,
								Description: "If true, yearly GFS backups are enabled.",
							},
							"keep_for_years": schema.Int64Attribute{
								Optional:    true,
								Description: "Number of yearly full backups to keep.",
							},
						},
					},
				},
			},
			"guest_processing": schema.SingleNestedBlock{
				Description: "Guest OS application-aware processing and guest credentials settings.",
				Attributes: map[string]schema.Attribute{
					"app_aware_processing_enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "If true, application-aware processing is enabled for transactional applications.",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"guest_indexing_enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "If true, guest file system indexing is enabled.",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"guest_credentials_id": schema.StringAttribute{
						Optional:    true,
						Description: "ID (GUID) of stored Veeam Credential entry to use for guest OS authentication.",
					},
				},
			},
		},
	}
}

func (r *JobVMwareResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *JobVMwareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan JobVMwareResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	defaultHost := plan.VMwareHostName.ValueString()
	var includeSpecs []client.VMwareIncludeObject

	if len(plan.Includes) > 0 {
		for _, inc := range plan.Includes {
			hName := inc.HostName.ValueString()
			if hName == "" {
				hName = defaultHost
			}

			oType := inc.Type.ValueString()
			oID := inc.ObjectID.ValueString()
			urn := inc.URN.ValueString()

			if strings.EqualFold(oType, "vCenterServer") || strings.EqualFold(oType, "VCenterServer") {
				oType = "vCenterServer"
				if oID == "" {
					oID = hName
				}
				if urn == "" {
					urn = fmt.Sprintf("vc:%s", hName)
				}
			}

			includeSpecs = append(includeSpecs, client.VMwareIncludeObject{
				Platform: "VSphere",
				Name:     inc.Name.ValueString(),
				HostName: hName,
				Type:     oType,
				ObjectID: oID,
				URN:      urn,
			})
		}
	} else {
		// Fallback to single target convenience fields
		objectType := plan.VMwareObjectType.ValueString()
		if objectType == "" {
			objectType = "vCenterServer"
		}

		objectID := plan.VMwareObjectID.ValueString()
		urn := plan.VMwareURN.ValueString()

		if strings.EqualFold(objectType, "vCenterServer") || strings.EqualFold(objectType, "VCenterServer") {
			objectType = "vCenterServer"
			if objectID == "" {
				objectID = defaultHost
			}
			if urn == "" {
				urn = fmt.Sprintf("vc:%s", defaultHost)
			}
		}

		includeSpecs = append(includeSpecs, client.VMwareIncludeObject{
			Platform: "VSphere",
			Name:     defaultHost,
			HostName: defaultHost,
			Type:     objectType,
			ObjectID: objectID,
			URN:      urn,
		})
	}

	if plan.VMwareObjectType.IsUnknown() || plan.VMwareObjectType.IsNull() {
		if len(includeSpecs) > 0 {
			plan.VMwareObjectType = types.StringValue(includeSpecs[0].Type)
		} else {
			plan.VMwareObjectType = types.StringValue("")
		}
	}

	if plan.VMwareObjectID.IsUnknown() || plan.VMwareObjectID.IsNull() {
		if len(includeSpecs) > 0 {
			plan.VMwareObjectID = types.StringValue(includeSpecs[0].ObjectID)
		} else {
			plan.VMwareObjectID = types.StringValue("")
		}
	}

	if plan.VMwareURN.IsUnknown() || plan.VMwareURN.IsNull() {
		if len(includeSpecs) > 0 {
			plan.VMwareURN = types.StringValue(includeSpecs[0].URN)
		} else {
			plan.VMwareURN = types.StringValue("")
		}
	}

	retentionType := plan.RetentionType.ValueString()
	if retentionType == "" {
		retentionType = "RestorePoints"
	}

	retentionQty := int(plan.RetentionQuantity.ValueInt64())
	if retentionQty <= 0 {
		retentionQty = 7
	}

	var scheduleSpec *client.BackupScheduleModel
	if plan.Schedule != nil {
		runAuto := true
		if !plan.Schedule.RunAutomatically.IsNull() && !plan.Schedule.RunAutomatically.IsUnknown() {
			runAuto = plan.Schedule.RunAutomatically.ValueBool()
		}

		scheduleSpec = &client.BackupScheduleModel{
			RunAutomatically: runAuto,
		}
		plan.Schedule.RunAutomatically = types.BoolValue(runAuto)

		if plan.Schedule.Daily != nil {
			dailyEnabled := true
			if !plan.Schedule.Daily.IsEnabled.IsNull() && !plan.Schedule.Daily.IsEnabled.IsUnknown() {
				dailyEnabled = plan.Schedule.Daily.IsEnabled.ValueBool()
			}

			localTime := plan.Schedule.Daily.LocalTime.ValueString()
			if localTime == "" {
				localTime = "22:00"
			}

			dailyKind := plan.Schedule.Daily.DailyKind.ValueString()
			if dailyKind == "" {
				dailyKind = "Everyday"
			}

			var days []string
			for _, d := range plan.Schedule.Daily.Days {
				if d.ValueString() != "" {
					days = append(days, d.ValueString())
				}
			}

			scheduleSpec.Daily = &client.ScheduleDailyModel{
				IsEnabled: dailyEnabled,
				LocalTime: localTime,
				DailyKind: dailyKind,
				Days:      days,
			}

			plan.Schedule.Daily.IsEnabled = types.BoolValue(dailyEnabled)
			plan.Schedule.Daily.LocalTime = types.StringValue(localTime)
			plan.Schedule.Daily.DailyKind = types.StringValue(dailyKind)
		}
	} else {
		scheduleSpec = &client.BackupScheduleModel{
			RunAutomatically: false,
		}
	}

	gfsPolicySpec := buildGFSPolicySpec(plan.GFSPolicy)
	guestProcessingSpec := buildGuestProcessingSpec(plan.GuestProcessing)

	spec := client.CreateVSphereBackupJobSpec{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Type:        "VSphereBackup",
		VirtualMachines: client.BackupJobVirtualMachinesSpec{
			Includes: includeSpecs,
		},
		Storage: client.BackupJobStorageModel{
			BackupRepositoryID: plan.RepositoryID.ValueString(),
			BackupProxies: client.BackupProxiesSettingsModel{
				AutoSelectEnabled: true,
			},
			RetentionPolicy: client.BackupJobRetentionPolicySettingsModel{
				Type:     retentionType,
				Quantity: retentionQty,
			},
			GFSPolicy: gfsPolicySpec,
		},
		GuestProcessing: guestProcessingSpec,
		Schedule:        scheduleSpec,
	}

	created, err := r.client.CreateVSphereJob(ctx, spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Veeam VMware Backup Job", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.RetentionType = types.StringValue(retentionType)
	plan.RetentionQuantity = types.Int64Value(int64(retentionQty))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *JobVMwareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state JobVMwareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	job, err := r.client.GetJobByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Veeam Backup Job", err.Error())
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

	if job.Storage != nil && job.Storage.GFSPolicy != nil {
		gfsEnabled := job.Storage.GFSPolicy.IsEnabled
		if state.GFSPolicy != nil || gfsEnabled {
			if state.GFSPolicy == nil {
				state.GFSPolicy = &GFSPolicyModel{}
			}
			state.GFSPolicy.IsEnabled = types.BoolValue(gfsEnabled)
			if job.Storage.GFSPolicy.Weekly != nil && state.GFSPolicy.Weekly != nil {
				state.GFSPolicy.Weekly.IsEnabled = types.BoolValue(job.Storage.GFSPolicy.Weekly.IsEnabled)
				state.GFSPolicy.Weekly.KeepForWeeks = types.Int64Value(int64(job.Storage.GFSPolicy.Weekly.KeepForNumberOfWeeks))
			}
			if job.Storage.GFSPolicy.Monthly != nil && state.GFSPolicy.Monthly != nil {
				state.GFSPolicy.Monthly.IsEnabled = types.BoolValue(job.Storage.GFSPolicy.Monthly.IsEnabled)
				state.GFSPolicy.Monthly.KeepForMonths = types.Int64Value(int64(job.Storage.GFSPolicy.Monthly.KeepForNumberOfMonths))
			}
			if job.Storage.GFSPolicy.Yearly != nil && state.GFSPolicy.Yearly != nil {
				state.GFSPolicy.Yearly.IsEnabled = types.BoolValue(job.Storage.GFSPolicy.Yearly.IsEnabled)
				state.GFSPolicy.Yearly.KeepForYears = types.Int64Value(int64(job.Storage.GFSPolicy.Yearly.KeepForNumberOfYears))
			}
		} else {
			state.GFSPolicy = nil
		}
	}

	if job.GuestProcessing != nil {
		appAware := job.GuestProcessing.AppAwareProcessing.IsEnabled
		indexing := job.GuestProcessing.GuestFSIndexing.IsEnabled
		hasCreds := job.GuestProcessing.GuestCredentials != nil && job.GuestProcessing.GuestCredentials.Credentials != nil && job.GuestProcessing.GuestCredentials.Credentials.CredentialsID != "" && job.GuestProcessing.GuestCredentials.Credentials.CredentialsID != "00000000-0000-0000-0000-000000000000"

		if state.GuestProcessing != nil || appAware || indexing || hasCreds {
			if state.GuestProcessing == nil {
				state.GuestProcessing = &GuestProcessingModel{}
			}
			state.GuestProcessing.AppAwareProcessingEnabled = types.BoolValue(appAware)
			state.GuestProcessing.GuestIndexingEnabled = types.BoolValue(indexing)
			if hasCreds {
				state.GuestProcessing.GuestCredentialsID = types.StringValue(job.GuestProcessing.GuestCredentials.Credentials.CredentialsID)
			}
		} else {
			state.GuestProcessing = nil
		}
	}

	if job.Schedule != nil {
		runAuto := job.Schedule.RunAutomatically && !job.IsDisabled
		if state.Schedule != nil || runAuto {
			if state.Schedule == nil {
				state.Schedule = &ScheduleModel{}
			}
			state.Schedule.RunAutomatically = types.BoolValue(runAuto)

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
				if len(state.Schedule.Daily.Days) > 0 && len(job.Schedule.Daily.Days) > 0 {
					var days []types.String
					for _, d := range job.Schedule.Daily.Days {
						days = append(days, types.StringValue(d))
					}
					state.Schedule.Daily.Days = days
				} else {
					state.Schedule.Daily.Days = nil
				}
			} else {
				state.Schedule.Daily = nil
			}
		} else {
			state.Schedule = nil
		}
	} else if job.IsDisabled {
		if state.Schedule != nil {
			state.Schedule.RunAutomatically = types.BoolValue(false)
			if state.Schedule.Daily != nil {
				state.Schedule.Daily.IsEnabled = types.BoolValue(false)
			}
		} else {
			state.Schedule = nil
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *JobVMwareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan JobVMwareResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state JobVMwareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := state.ID.ValueString()
	if jobID == "" {
		resp.Diagnostics.AddError("Error Updating Veeam VMware Backup Job", "Missing job ID in state.")
		return
	}

	defaultHost := plan.VMwareHostName.ValueString()
	var includeSpecs []client.VMwareIncludeObject

	if len(plan.Includes) > 0 {
		for _, inc := range plan.Includes {
			hName := inc.HostName.ValueString()
			if hName == "" {
				hName = defaultHost
			}

			oType := inc.Type.ValueString()
			oID := inc.ObjectID.ValueString()
			urn := inc.URN.ValueString()

			if strings.EqualFold(oType, "vCenterServer") || strings.EqualFold(oType, "VCenterServer") {
				oType = "vCenterServer"
				if oID == "" {
					oID = hName
				}
				if urn == "" {
					urn = fmt.Sprintf("vc:%s", hName)
				}
			}

			includeSpecs = append(includeSpecs, client.VMwareIncludeObject{
				Platform: "VSphere",
				Name:     inc.Name.ValueString(),
				HostName: hName,
				Type:     oType,
				ObjectID: oID,
				URN:      urn,
			})
		}
	} else {
		objectType := plan.VMwareObjectType.ValueString()
		if objectType == "" {
			objectType = "vCenterServer"
		}

		objectID := plan.VMwareObjectID.ValueString()
		urn := plan.VMwareURN.ValueString()

		if strings.EqualFold(objectType, "vCenterServer") || strings.EqualFold(objectType, "VCenterServer") {
			objectType = "vCenterServer"
			if objectID == "" {
				objectID = defaultHost
			}
			if urn == "" {
				urn = fmt.Sprintf("vc:%s", defaultHost)
			}
		}

		includeSpecs = append(includeSpecs, client.VMwareIncludeObject{
			Platform: "VSphere",
			Name:     defaultHost,
			HostName: defaultHost,
			Type:     objectType,
			ObjectID: objectID,
			URN:      urn,
		})
	}

	if plan.VMwareObjectType.IsUnknown() || plan.VMwareObjectType.IsNull() {
		if len(includeSpecs) > 0 {
			plan.VMwareObjectType = types.StringValue(includeSpecs[0].Type)
		} else {
			plan.VMwareObjectType = types.StringValue("")
		}
	}

	if plan.VMwareObjectID.IsUnknown() || plan.VMwareObjectID.IsNull() {
		if len(includeSpecs) > 0 {
			plan.VMwareObjectID = types.StringValue(includeSpecs[0].ObjectID)
		} else {
			plan.VMwareObjectID = types.StringValue("")
		}
	}

	if plan.VMwareURN.IsUnknown() || plan.VMwareURN.IsNull() {
		if len(includeSpecs) > 0 {
			plan.VMwareURN = types.StringValue(includeSpecs[0].URN)
		} else {
			plan.VMwareURN = types.StringValue("")
		}
	}

	retentionType := plan.RetentionType.ValueString()
	if retentionType == "" {
		retentionType = "RestorePoints"
	}

	retentionQty := int(plan.RetentionQuantity.ValueInt64())
	if retentionQty <= 0 {
		retentionQty = 7
	}

	var scheduleSpec *client.BackupScheduleModel
	if plan.Schedule != nil {
		runAuto := true
		if !plan.Schedule.RunAutomatically.IsNull() && !plan.Schedule.RunAutomatically.IsUnknown() {
			runAuto = plan.Schedule.RunAutomatically.ValueBool()
		}

		scheduleSpec = &client.BackupScheduleModel{
			RunAutomatically: runAuto,
		}
		plan.Schedule.RunAutomatically = types.BoolValue(runAuto)

		if plan.Schedule.Daily != nil {
			dailyEnabled := true
			if !plan.Schedule.Daily.IsEnabled.IsNull() && !plan.Schedule.Daily.IsEnabled.IsUnknown() {
				dailyEnabled = plan.Schedule.Daily.IsEnabled.ValueBool()
			}

			localTime := plan.Schedule.Daily.LocalTime.ValueString()
			if localTime == "" {
				localTime = "22:00"
			}

			dailyKind := plan.Schedule.Daily.DailyKind.ValueString()
			if dailyKind == "" {
				dailyKind = "Everyday"
			}

			var days []string
			for _, d := range plan.Schedule.Daily.Days {
				if d.ValueString() != "" {
					days = append(days, d.ValueString())
				}
			}

			scheduleSpec.Daily = &client.ScheduleDailyModel{
				IsEnabled: dailyEnabled,
				LocalTime: localTime,
				DailyKind: dailyKind,
				Days:      days,
			}

			plan.Schedule.Daily.IsEnabled = types.BoolValue(dailyEnabled)
			plan.Schedule.Daily.LocalTime = types.StringValue(localTime)
			plan.Schedule.Daily.DailyKind = types.StringValue(dailyKind)
		}
	} else {
		scheduleSpec = &client.BackupScheduleModel{
			RunAutomatically: false,
		}
	}

	gfsPolicySpec := buildGFSPolicySpec(plan.GFSPolicy)
	guestProcessingSpec := buildGuestProcessingSpec(plan.GuestProcessing)

	spec := client.CreateVSphereBackupJobSpec{
		ID:          jobID,
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Type:        "VSphereBackup",
		VirtualMachines: client.BackupJobVirtualMachinesSpec{
			Includes: includeSpecs,
		},
		Storage: client.BackupJobStorageModel{
			BackupRepositoryID: plan.RepositoryID.ValueString(),
			BackupProxies: client.BackupProxiesSettingsModel{
				AutoSelectEnabled: true,
			},
			RetentionPolicy: client.BackupJobRetentionPolicySettingsModel{
				Type:     retentionType,
				Quantity: retentionQty,
			},
			GFSPolicy: gfsPolicySpec,
		},
		GuestProcessing: guestProcessingSpec,
		Schedule:        scheduleSpec,
	}

	updated, err := r.client.UpdateVSphereJob(ctx, jobID, spec)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Veeam VMware Backup Job", err.Error())
		return
	}

	plan.ID = types.StringValue(updated.ID)
	plan.RetentionType = types.StringValue(retentionType)
	plan.RetentionQuantity = types.Int64Value(int64(retentionQty))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func buildGFSPolicySpec(plan *GFSPolicyModel) *client.GFSPolicySettingsModel {
	if plan == nil {
		return nil
	}

	gfsEnabled := true
	if !plan.IsEnabled.IsNull() && !plan.IsEnabled.IsUnknown() {
		gfsEnabled = plan.IsEnabled.ValueBool()
	}
	plan.IsEnabled = types.BoolValue(gfsEnabled)

	spec := &client.GFSPolicySettingsModel{
		IsEnabled: gfsEnabled,
	}

	if plan.Weekly != nil {
		weeklyEnabled := true
		if !plan.Weekly.IsEnabled.IsNull() && !plan.Weekly.IsEnabled.IsUnknown() {
			weeklyEnabled = plan.Weekly.IsEnabled.ValueBool()
		}
		weeks := int(plan.Weekly.KeepForWeeks.ValueInt64())
		if weeks <= 0 {
			weeks = 4
		}
		spec.Weekly = &client.GFSPolicySettingsWeeklyModel{
			IsEnabled:            weeklyEnabled,
			KeepForNumberOfWeeks: weeks,
		}
		plan.Weekly.IsEnabled = types.BoolValue(weeklyEnabled)
		plan.Weekly.KeepForWeeks = types.Int64Value(int64(weeks))
	}

	if plan.Monthly != nil {
		monthlyEnabled := true
		if !plan.Monthly.IsEnabled.IsNull() && !plan.Monthly.IsEnabled.IsUnknown() {
			monthlyEnabled = plan.Monthly.IsEnabled.ValueBool()
		}
		months := int(plan.Monthly.KeepForMonths.ValueInt64())
		if months <= 0 {
			months = 12
		}
		spec.Monthly = &client.GFSPolicySettingsMonthlyModel{
			IsEnabled:             monthlyEnabled,
			KeepForNumberOfMonths: months,
		}
		plan.Monthly.IsEnabled = types.BoolValue(monthlyEnabled)
		plan.Monthly.KeepForMonths = types.Int64Value(int64(months))
	}

	if plan.Yearly != nil {
		yearlyEnabled := true
		if !plan.Yearly.IsEnabled.IsNull() && !plan.Yearly.IsEnabled.IsUnknown() {
			yearlyEnabled = plan.Yearly.IsEnabled.ValueBool()
		}
		years := int(plan.Yearly.KeepForYears.ValueInt64())
		if years <= 0 {
			years = 2
		}
		spec.Yearly = &client.GFSPolicySettingsYearlyModel{
			IsEnabled:            yearlyEnabled,
			KeepForNumberOfYears: years,
		}
		plan.Yearly.IsEnabled = types.BoolValue(yearlyEnabled)
		plan.Yearly.KeepForYears = types.Int64Value(int64(years))
	}

	return spec
}

func buildGuestProcessingSpec(plan *GuestProcessingModel) client.BackupJobGuestProcessingModel {
	appAwareEnabled := false
	guestIndexingEnabled := false
	var guestCreds *client.GuestOsCredentialsModel

	if plan != nil {
		if !plan.AppAwareProcessingEnabled.IsNull() && !plan.AppAwareProcessingEnabled.IsUnknown() {
			appAwareEnabled = plan.AppAwareProcessingEnabled.ValueBool()
		}
		if !plan.GuestIndexingEnabled.IsNull() && !plan.GuestIndexingEnabled.IsUnknown() {
			guestIndexingEnabled = plan.GuestIndexingEnabled.ValueBool()
		}
		credID := plan.GuestCredentialsID.ValueString()
		if credID != "" {
			guestCreds = &client.GuestOsCredentialsModel{
				UseAgentManagementCredentials: false,
				Credentials: &client.SpecifiedGuestOsCredentialsModel{
					CredentialsID:   credID,
					CredentialsType: "Standard",
				},
			}
		}

		plan.AppAwareProcessingEnabled = types.BoolValue(appAwareEnabled)
		plan.GuestIndexingEnabled = types.BoolValue(guestIndexingEnabled)
	}

	return client.BackupJobGuestProcessingModel{
		AppAwareProcessing: client.BackupApplicationAwareProcessingModel{
			IsEnabled: appAwareEnabled,
		},
		GuestFSIndexing: client.GuestFileSystemIndexingModel{
			IsEnabled: guestIndexingEnabled,
		},
		GuestCredentials: guestCreds,
	}
}

func (r *JobVMwareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state JobVMwareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteJob(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Veeam Backup Job", err.Error())
		return
	}
}

func (r *JobVMwareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
