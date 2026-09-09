package vpc

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	vpcclient "go.mws.cloud/go-sdk/service/vpc/client"
	vpcmodel "go.mws.cloud/go-sdk/service/vpc/model"
	vpcsdk "go.mws.cloud/go-sdk/service/vpc/sdk"

	"cq-source-mws/client"
)

func EgressNats() *schema.Table {
	return &schema.Table{
		Name:        "mws_vpc_egress_nats",
		Description: "Egress NAT of every network",
		Resolver:    fetchEgressNats,
		Transform:   client.TransformResource(&vpcmodel.EgressNatOptionalResponse{}),
		Columns:     client.ChildColumns(networkColumn),
	}
}

func fetchEgressNats(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	network, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	nats, err := vpcsdk.NewEgressNat(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := vpcclient.ListEgressNatsRequest{Project: c.ProjectName, Network: network}
	return client.List(ctx, req, nats.ListEgressNats, res, c.Skip(serviceName))
}
