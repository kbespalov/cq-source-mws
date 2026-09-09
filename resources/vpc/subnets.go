package vpc

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	vpcclient "go.mws.cloud/go-sdk/service/vpc/client"
	vpcmodel "go.mws.cloud/go-sdk/service/vpc/model"
	vpcsdk "go.mws.cloud/go-sdk/service/vpc/sdk"

	"cq-source-mws/client"
)

func Subnets() *schema.Table {
	return &schema.Table{
		Name:        "mws_vpc_subnets",
		Description: "Subnets of every network",
		Resolver:    fetchSubnets,
		Transform:   client.TransformResource(&vpcmodel.SubnetOptionalResponse{}),
		Columns:     client.ChildColumns(networkColumn),
	}
}

func fetchSubnets(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	network, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	subnets, err := vpcsdk.NewSubnet(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := vpcclient.ListSubnetsRequest{Project: c.ProjectName, Network: network}
	return client.List(ctx, req, subnets.ListSubnets, res, c.Skip(serviceName))
}
