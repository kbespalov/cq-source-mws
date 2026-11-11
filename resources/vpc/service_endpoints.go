package vpc

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	vpcclient "go.mws.cloud/go-sdk/service/vpc/client"
	vpcmodel "go.mws.cloud/go-sdk/service/vpc/model"
	vpcsdk "go.mws.cloud/go-sdk/service/vpc/sdk"

	"cq-source-mws/client"
)

// ServiceEndpoints lists the private service endpoints of a network: where a
// subnet reaches a platform service without leaving the VPC.
func ServiceEndpoints() *schema.Table {
	return &schema.Table{
		Name:        "mws_vpc_service_endpoints",
		Description: "Service endpoints of every synced network",
		Resolver:    fetchServiceEndpoints,
		Transform:   client.TransformResource(&vpcmodel.ServiceEndpointOptionalResponse{}),
		Columns:     client.ChildColumns(networkColumn),
	}
}

func fetchServiceEndpoints(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	network, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	endpoints, err := vpcsdk.NewServiceEndpoint(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := vpcclient.ListServiceEndpointsRequest{Project: c.ProjectName, Network: network}
	return client.List(ctx, req, endpoints.ListServiceEndpoints, res, c.Skip(serviceName))
}
