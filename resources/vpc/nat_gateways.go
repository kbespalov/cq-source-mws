package vpc

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	vpcclient "go.mws.cloud/go-sdk/service/vpc/client"
	vpcmodel "go.mws.cloud/go-sdk/service/vpc/model"
	vpcsdk "go.mws.cloud/go-sdk/service/vpc/sdk"

	"cq-source-mws/client"
)

func NatGateways() *schema.Table {
	return &schema.Table{
		Name:        "mws_vpc_nat_gateways",
		Description: "NAT gateways of every synced project",
		Resolver:    fetchNatGateways,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&vpcmodel.NatGatewayOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchNatGateways(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	gateways, err := vpcsdk.NewNatGateway(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := vpcclient.ListNatGatewaysRequest{Project: c.ProjectName}
	return client.List(ctx, req, gateways.ListNatGateways, res, c.Skip(serviceName))
}
