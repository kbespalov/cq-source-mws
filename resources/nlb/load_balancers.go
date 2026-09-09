// Package nlb holds the network load balancer table.
package nlb

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	nlbclient "go.mws.cloud/go-sdk/service/nlb/client"
	nlbmodel "go.mws.cloud/go-sdk/service/nlb/model"
	nlbsdk "go.mws.cloud/go-sdk/service/nlb/sdk"

	"cq-source-mws/client"
)

const serviceName = "nlb"

// LoadBalancers lists per project rather than per network, which the service
// also offers, so the table stays a sibling of the other project tables.
func LoadBalancers() *schema.Table {
	return &schema.Table{
		Name:        "mws_nlb_load_balancers",
		Description: "Network load balancers of every synced project",
		Resolver:    fetchLoadBalancers,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&nlbmodel.NlbOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchLoadBalancers(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	balancers, err := nlbsdk.NewNlb(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := nlbclient.ListNlbsByProjectIdRequest{Project: c.ProjectName}
	return client.List(ctx, req, balancers.ListNlbsByProjectId, res, c.Skip(serviceName))
}
