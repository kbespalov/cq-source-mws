package vpc

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	vpcclient "go.mws.cloud/go-sdk/service/vpc/client"
	vpcmodel "go.mws.cloud/go-sdk/service/vpc/model"
	vpcsdk "go.mws.cloud/go-sdk/service/vpc/sdk"

	"cq-source-mws/client"
)

// ExternalAddresses is the public side of a project's network. Unlike the
// internal addresses, these belong to the project rather than to a network.
func ExternalAddresses() *schema.Table {
	return &schema.Table{
		Name:        "mws_vpc_external_addresses",
		Description: "External addresses of every synced project",
		Resolver:    fetchExternalAddresses,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&vpcmodel.ExternalAddressOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchExternalAddresses(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	addresses, err := vpcsdk.NewExternalAddress(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := vpcclient.ListExternalAddressesRequest{Project: c.ProjectName}
	return client.List(ctx, req, addresses.ListExternalAddresses, res, c.Skip(serviceName))
}
