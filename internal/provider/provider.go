package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/michael/terraform-provider-veeam/pkg/client"
)

var _ provider.Provider = &VeeamProvider{}

type VeeamProvider struct {
	version string
}

type VeeamProviderModel struct {
	Endpoint   types.String `tfsdk:"endpoint"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	APIVersion types.String `tfsdk:"api_version"`
	Insecure   types.Bool   `tfsdk:"insecure"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &VeeamProvider{
			version: version,
		}
	}
}

func (p *VeeamProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "veeam"
	resp.Version = p.version
}

func (p *VeeamProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for interacting with Veeam Backup & Replication REST API.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Required:    true,
				Description: "The REST API endpoint URL of the Veeam Backup & Replication server (e.g., https://vbr.lab.local:9419).",
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "Username for Veeam VBR authentication.",
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Password for Veeam VBR authentication.",
			},
			"api_version": schema.StringAttribute{
				Optional:    true,
				Description: "Veeam REST API version header (default: 1.1-rev0).",
			},
			"insecure": schema.BoolAttribute{
				Optional:    true,
				Description: "Set to true to skip TLS certificate verification (useful for lab environments with self-signed certificates).",
			},
		},
	}
}

func (p *VeeamProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data VeeamProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := data.Endpoint.ValueString()
	username := data.Username.ValueString()
	password := data.Password.ValueString()
	apiVersion := data.APIVersion.ValueString()
	insecure := data.Insecure.ValueBool()

	c, err := client.NewClient(client.Config{
		Endpoint:   endpoint,
		Username:   username,
		Password:   password,
		APIVersion: apiVersion,
		Insecure:   insecure,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Initialization Failed", err.Error())
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *VeeamProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewBackupRepositoriesDataSource,
		NewScaleOutRepositoriesDataSource,
		NewManagedServersDataSource,
		NewJobsDataSource,
		NewInventoryDataSource,
		NewUnstructuredDataServersDataSource,
	}
}

func (p *VeeamProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCredentialResource,
		NewJobVMwareResource,
		NewJobHyperVResource,
		NewJobProxmoxResource,
		NewJobNutanixResource,
		NewJobFileShareResource,
		NewManagedServerLinuxResource,
		NewManagedServerWindowsResource,
		NewRepositoryLinuxResource,
		NewRepositoryWindowsResource,
		NewRepositorySmbResource,
		NewRepositoryNfsResource,
		NewRepositoryS3CompatibleResource,
		NewRepositoryVeeamVaultResource,
		NewRepositoryScaleOutResource,
		NewUnstructuredDataSmbShareResource,
		NewUnstructuredDataNfsShareResource,
	}
}
