package vpc

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	vpcclient "go.mws.cloud/go-sdk/service/vpc/client"
	vpcmodel "go.mws.cloud/go-sdk/service/vpc/model"
	vpcsdk "go.mws.cloud/go-sdk/service/vpc/sdk"

	"cq-source-mws/client"
)

func Routes() *schema.Table {
	return &schema.Table{
		Name:        "mws_vpc_routes",
		Description: "Routes of every network",
		Resolver:    fetchRoutes,
		Transform:   client.TransformResource(&vpcmodel.RouteOptionalResponse{}),
		Columns:     client.ChildColumns(networkColumn),
	}
}

func fetchRoutes(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	network, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	routes, err := vpcsdk.NewRoute(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := vpcclient.ListRoutesRequest{Project: c.ProjectName, Network: network}
	return client.List(ctx, req, routes.ListRoutes, res, c.Skip(serviceName))
}
