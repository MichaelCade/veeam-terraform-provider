package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/michael/terraform-provider-veeam/pkg/client"
)

var _ datasource.DataSource = &BackupRepositoriesDataSource{}

type BackupRepositoriesDataSource struct {
	client *client.Client
}

type BackupRepositoriesDataSourceModel struct {
	Repositories []RepositoryModel `tfsdk:"repositories"`
}

type RepositoryModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Path        types.String `tfsdk:"path"`
}

func NewBackupRepositoriesDataSource() datasource.DataSource {
	return &BackupRepositoriesDataSource{}
}

func (d *BackupRepositoriesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup_repositories"
}

func (d *BackupRepositoriesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of backup repositories configured in Veeam Backup & Replication.",
		Attributes: map[string]schema.Attribute{
			"repositories": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of repositories.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Repository unique identifier (GUID).",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Repository name.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Repository description.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "Repository type (e.g. WinLocal, LinuxLocal, S3).",
						},
						"path": schema.StringAttribute{
							Computed:    true,
							Description: "Repository storage path.",
						},
					},
				},
			},
		},
	}
}

func (d *BackupRepositoriesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BackupRepositoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state BackupRepositoriesDataSourceModel

	repos, err := d.client.GetRepositories(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to Read Backup Repositories", err.Error())
		return
	}

	for _, r := range repos {
		state.Repositories = append(state.Repositories, RepositoryModel{
			ID:          types.StringValue(r.ID),
			Name:        types.StringValue(r.Name),
			Description: types.StringValue(r.Description),
			Type:        types.StringValue(r.Type),
			Path:        types.StringValue(r.Path),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
