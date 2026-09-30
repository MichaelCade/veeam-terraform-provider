package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/michael/terraform-provider-veeam/pkg/client"
)

var _ datasource.DataSource = &JobsDataSource{}

type JobsDataSource struct {
	client *client.Client
}

type JobsDataSourceModel struct {
	Jobs []JobModel `tfsdk:"jobs"`
}

type JobModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	IsDisabled types.Bool   `tfsdk:"is_disabled"`
}

func NewJobsDataSource() datasource.DataSource {
	return &JobsDataSource{}
}

func (d *JobsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jobs"
}

func (d *JobsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of backup jobs configured in Veeam Backup & Replication.",
		Attributes: map[string]schema.Attribute{
			"jobs": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of backup jobs.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Job unique identifier (GUID).",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Job name.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "Job type (e.g., Backup, BackupSync, AgentBackup).",
						},
						"is_disabled": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the job schedule is disabled.",
						},
					},
				},
			},
		},
	}
}

func (d *JobsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *JobsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state JobsDataSourceModel

	jobs, err := d.client.GetJobs(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to Read Jobs", err.Error())
		return
	}

	for _, j := range jobs {
		state.Jobs = append(state.Jobs, JobModel{
			ID:         types.StringValue(j.ID),
			Name:       types.StringValue(j.Name),
			Type:       types.StringValue(j.Type),
			IsDisabled: types.BoolValue(j.IsDisabled),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
