package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/michael/terraform-provider-veeam/pkg/client"
)

var _ datasource.DataSource = &ManagedServersDataSource{}

type ManagedServersDataSource struct {
	client *client.Client
}

type ManagedServersDataSourceModel struct {
	Servers []ManagedServerModel `tfsdk:"servers"`
}

type ManagedServerModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Status      types.String `tfsdk:"status"`
}

func NewManagedServersDataSource() datasource.DataSource {
	return &ManagedServersDataSource{}
}

func (d *ManagedServersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_servers"
}

func (d *ManagedServersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of managed infrastructure servers (vCenter, Windows, Linux, Hyper-V) configured in Veeam Backup & Replication.",
		Attributes: map[string]schema.Attribute{
			"servers": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of managed servers.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Server unique identifier (GUID).",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Server hostname or IP address.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Server description.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "Server type (e.g. VMware, WindowsHost, LinuxHost, HyperV).",
						},
						"status": schema.StringAttribute{
							Computed:    true,
							Description: "Server connection status.",
						},
					},
				},
			},
		},
	}
}

func (d *ManagedServersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ManagedServersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ManagedServersDataSourceModel

	servers, err := d.client.GetManagedServers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to Read Managed Servers", err.Error())
		return
	}

	for _, s := range servers {
		state.Servers = append(state.Servers, ManagedServerModel{
			ID:          types.StringValue(s.ID),
			Name:        types.StringValue(s.Name),
			Description: types.StringValue(s.Description),
			Type:        types.StringValue(s.Type),
			Status:      types.StringValue(s.Status),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
