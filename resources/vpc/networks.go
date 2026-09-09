// Package vpc holds the tables of the network service.
//
// VPC nests most of what it offers under a network: a subnet is listed as
// /projects/{project}/networks/{network}/subnets, and there is no way to ask
// for the subnets of a project. Those tables are therefore children of
// mws_vpc_networks, which is both what the API forces and what reads best,
// since a subnet without its network means little. Only networks themselves,
// NAT gateways and external addresses hang off the project directly.
package vpc

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	vpcclient "go.mws.cloud/go-sdk/service/vpc/client"
	vpcmodel "go.mws.cloud/go-sdk/service/vpc/model"
	vpcsdk "go.mws.cloud/go-sdk/service/vpc/sdk"

	"cq-source-mws/client"
)

const (
	// serviceName is what a skipped project is logged under.
	serviceName = "vpc"
	// networkColumn joins a nested table back to its network.
	networkColumn = "network_id"
)

func Networks() *schema.Table {
	return &schema.Table{
		Name:        "mws_vpc_networks",
		Description: "Networks of every synced project",
		Resolver:    fetchNetworks,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&vpcmodel.NetworkOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
		Relations: []*schema.Table{
			Subnets(),
			Routes(),
			FirewallRules(),
			Addresses(),
			AddressGroups(),
			EgressNats(),
			OneToOneNats(),
		},
	}
}

func fetchNetworks(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	networks, err := vpcsdk.NewNetwork(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := vpcclient.ListNetworksRequest{Project: c.ProjectName}
	return client.List(ctx, req, networks.ListNetworks, res, c.Skip(serviceName))
}
