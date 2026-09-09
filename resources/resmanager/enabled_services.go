package resmanager

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	rmclient "go.mws.cloud/go-sdk/service/rm/client"
	rmmodel "go.mws.cloud/go-sdk/service/rm/model"
	rmsdk "go.mws.cloud/go-sdk/service/rm/sdk"

	"cq-source-mws/client"
)

// serviceName is what a skipped project is logged under.
const serviceName = "resmanager"

// EnabledServices says which services a project has turned on, which is the
// answer to why most of its other tables are empty.
func EnabledServices() *schema.Table {
	return &schema.Table{
		Name:        "mws_resmanager_enabled_services",
		Description: "Services every synced project has enabled",
		Resolver:    fetchEnabledServices,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&rmmodel.EnabledServiceOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchEnabledServices(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	services, err := rmsdk.NewEnabledService(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := rmclient.ListEnabledServicesRequest{Project: c.ProjectName}
	return client.List(ctx, req, services.ListEnabledServices, res, c.Skip(serviceName))
}
