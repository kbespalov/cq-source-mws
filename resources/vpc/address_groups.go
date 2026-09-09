package vpc

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	vpcclient "go.mws.cloud/go-sdk/service/vpc/client"
	vpcmodel "go.mws.cloud/go-sdk/service/vpc/model"
	vpcsdk "go.mws.cloud/go-sdk/service/vpc/sdk"

	"cq-source-mws/client"
)

func AddressGroups() *schema.Table {
	return &schema.Table{
		Name:        "mws_vpc_address_groups",
		Description: "Address groups of every network",
		Resolver:    fetchAddressGroups,
		Transform:   client.TransformResource(&vpcmodel.VpcAddressGroupOptionalResponse{}),
		Columns:     client.ChildColumns(networkColumn),
	}
}

func fetchAddressGroups(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	network, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	groups, err := vpcsdk.NewAddressGroup(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := vpcclient.ListAddressGroupsRequest{Project: c.ProjectName, Network: network}
	return client.List(ctx, req, groups.ListAddressGroups, res, c.Skip(serviceName))
}
