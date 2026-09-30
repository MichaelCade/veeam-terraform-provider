package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/michael/terraform-provider-veeam/pkg/client"
)

var _ datasource.DataSource = &InventoryDataSource{}

type InventoryDataSource struct {
	client *client.Client
}

type InventoryDataSourceModel struct {
	HostName      types.String         `tfsdk:"host_name"`
	HierarchyType types.String         `tfsdk:"hierarchy_type"`
	NameFilter    types.String         `tfsdk:"name_filter"`
	Items         []InventoryItemModel `tfsdk:"items"`
}

type InventoryItemModel struct {
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	HostName types.String `tfsdk:"host_name"`
	ObjectID types.String `tfsdk:"object_id"`
	URN      types.String `tfsdk:"urn"`
}

func NewInventoryDataSource() datasource.DataSource {
	return &InventoryDataSource{}
}

func (d *InventoryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_inventory"
}

func (d *InventoryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Browses and searches VMware vSphere inventory items (VMs, Tags, Categories, Folders) registered in Veeam Backup & Replication.",
		Attributes: map[string]schema.Attribute{
			"host_name": schema.StringAttribute{
				Optional:    true,
				Description: "vCenter Server or ESXi Host name/IP to search within (e.g. 192.168.169.181).",
			},
			"hierarchy_type": schema.StringAttribute{
				Optional:    true,
				Description: "Inventory hierarchy type to browse (VmsAndTemplates, VmsAndTags, HostsAndClusters). Defaults to VmsAndTemplates.",
			},
			"name_filter": schema.StringAttribute{
				Optional:    true,
				Description: "Case-insensitive substring filter for matching item names (e.g. SQLServer-1 or Linux Desktop Backup).",
			},
			"items": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of matching inventory items.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the inventory object.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "Type of inventory object (VirtualMachine, Tag, Category, Folder, Host, vCenterServer).",
						},
						"host_name": schema.StringAttribute{
							Computed:    true,
							Description: "vCenter or ESXi host name.",
						},
						"object_id": schema.StringAttribute{
							Computed:    true,
							Description: "vSphere MOB ID or Tag ID (e.g. vm-101).",
						},
						"urn": schema.StringAttribute{
							Computed:    true,
							Description: "Uniform Resource Name (URN) for the object.",
						},
					},
				},
			},
		},
	}
}

func (d *InventoryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InventoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state InventoryDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostName := state.HostName.ValueString()
	hierarchyType := state.HierarchyType.ValueString()
	if hierarchyType == "" {
		hierarchyType = "VmsAndTemplates"
	}

	nameFilter := state.NameFilter.ValueString()

	rawItems, err := d.client.GetInventory(ctx, hostName, hierarchyType, nameFilter)
	if err != nil {
		resp.Diagnostics.AddError("Failed to Read Veeam Inventory", err.Error())
		return
	}

	matchedItems := make([]InventoryItemModel, 0)
	filterLower := strings.ToLower(nameFilter)

	for _, item := range rawItems {
		if filterLower != "" && !strings.Contains(strings.ToLower(item.Name), filterLower) {
			continue
		}

		matchedItems = append(matchedItems, InventoryItemModel{
			Name:     types.StringValue(item.Name),
			Type:     types.StringValue(item.Type),
			HostName: types.StringValue(item.HostName),
			ObjectID: types.StringValue(item.ObjectID),
			URN:      types.StringValue(item.URN),
		})
	}

	state.Items = matchedItems
	state.HierarchyType = types.StringValue(hierarchyType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
