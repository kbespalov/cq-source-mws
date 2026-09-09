package vpc

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	vpcclient "go.mws.cloud/go-sdk/service/vpc/client"
	vpcmodel "go.mws.cloud/go-sdk/service/vpc/model"
	vpcsdk "go.mws.cloud/go-sdk/service/vpc/sdk"

	"cq-source-mws/client"
)

func Addresses() *schema.Table {
	return &schema.Table{
		Name:        "mws_vpc_addresses",
		Description: "Internal addresses of every network",
		Resolver:    fetchAddresses,
		Transform:   client.TransformResource(&vpcmodel.AddressOptionalResponse{}),
		Columns:     client.ChildColumns(networkColumn),
	}
}

func fetchAddresses(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	network, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	addresses, err := vpcsdk.NewAddress(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := vpcclient.ListAddressesRequest{Project: c.ProjectName, Network: network}
	return client.List(ctx, req, addresses.ListAddresses, res, c.Skip(serviceName))
}
