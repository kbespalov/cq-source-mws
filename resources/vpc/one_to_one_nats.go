package vpc

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	vpcclient "go.mws.cloud/go-sdk/service/vpc/client"
	vpcmodel "go.mws.cloud/go-sdk/service/vpc/model"
	vpcsdk "go.mws.cloud/go-sdk/service/vpc/sdk"

	"cq-source-mws/client"
)

func OneToOneNats() *schema.Table {
	return &schema.Table{
		Name:        "mws_vpc_one_to_one_nats",
		Description: "One-to-one NAT of every network",
		Resolver:    fetchOneToOneNats,
		Transform:   client.TransformResource(&vpcmodel.OneToOneNatOptionalResponse{}),
		Columns:     client.ChildColumns(networkColumn),
	}
}

func fetchOneToOneNats(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	network, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	nats, err := vpcsdk.NewOneToOneNat(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := vpcclient.ListOneToOneNatsRequest{Project: c.ProjectName, Network: network}
	return client.List(ctx, req, nats.ListOneToOneNats, res, c.Skip(serviceName))
}
