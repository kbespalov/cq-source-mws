package vpc

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	vpcclient "go.mws.cloud/go-sdk/service/vpc/client"
	vpcmodel "go.mws.cloud/go-sdk/service/vpc/model"
	vpcsdk "go.mws.cloud/go-sdk/service/vpc/sdk"

	"cq-source-mws/client"
)

func FirewallRules() *schema.Table {
	return &schema.Table{
		Name:        "mws_vpc_firewall_rules",
		Description: "Firewall rules of every network",
		Resolver:    fetchFirewallRules,
		Transform:   client.TransformResource(&vpcmodel.FirewallRuleOptionalResponse{}),
		Columns:     client.ChildColumns(networkColumn),
	}
}

func fetchFirewallRules(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	network, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	rules, err := vpcsdk.NewFirewallRule(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := vpcclient.ListFirewallRulesRequest{Project: c.ProjectName, Network: network}
	return client.List(ctx, req, rules.ListFirewallRules, res, c.Skip(serviceName))
}
