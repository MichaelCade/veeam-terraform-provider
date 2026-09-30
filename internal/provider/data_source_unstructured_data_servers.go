package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/michael/terraform-provider-veeam/pkg/client"
)

var _ datasource.DataSource = &UnstructuredDataServersDataSource{}

type UnstructuredDataServersDataSource struct {
	client *client.Client
}

type UnstructuredDataServerModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Type                types.String `tfsdk:"type"`
	Path                types.String `tfsdk:"path"`
	AccessCredentialsID types.String `tfsdk:"access_credentials_id"`
	CacheRepositoryID   types.String `tfsdk:"cache_repository_id"`
}

type UnstructuredDataServersDataSourceModel struct {
	Servers []UnstructuredDataServerModel `tfsdk:"servers"`
}

func NewUnstructuredDataServersDataSource() datasource.DataSource {
	return &UnstructuredDataServersDataSource{}
}

func (d *UnstructuredDataServersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unstructured_data_servers"
}

func (d *UnstructuredDataServersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source to query all unstructured data backup sources (SMB shares, NFS exports, Object Storage) registered in Veeam Inventory.",
		Attributes: map[string]schema.Attribute{
			"servers": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of registered unstructured data sources.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique identifier (GUID) of the unstructured data source.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Display name or DNS name.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "Type of unstructured data server (e.g., SMBShare, NFSShare).",
						},
						"path": schema.StringAttribute{
							Computed:    true,
							Description: "Path or export string.",
						},
						"access_credentials_id": schema.StringAttribute{
							Computed:    true,
							Description: "ID of assigned credentials if required.",
						},
						"cache_repository_id": schema.StringAttribute{
							Computed:    true,
							Description: "ID of assigned cache repository.",
						},
					},
				},
			},
		},
	}
}

func (d *UnstructuredDataServersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UnstructuredDataServersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state UnstructuredDataServersDataSourceModel

	servers, err := d.client.GetUnstructuredDataServers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Fetching Unstructured Data Servers", err.Error())
		return
	}

	for _, srv := range servers {
		model := UnstructuredDataServerModel{
			ID:                  types.StringValue(srv.ID),
			Name:                types.StringValue(srv.Name),
			Type:                types.StringValue(srv.Type),
			Path:                types.StringValue(srv.Path),
			AccessCredentialsID: types.StringValue(srv.AccessCredentialsID),
			CacheRepositoryID:   types.StringValue(srv.CacheRepositoryID),
		}
		state.Servers = append(state.Servers, model)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
