package mk8s

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	mk8sclient "go.mws.cloud/go-sdk/service/mk8s/client"
	mk8smodel "go.mws.cloud/go-sdk/service/mk8s/model"
	mk8ssdk "go.mws.cloud/go-sdk/service/mk8s/sdk"

	"cq-source-mws/client"
)

func NodeGroups() *schema.Table {
	return &schema.Table{
		Name:        "mws_mk8s_node_groups",
		Description: "Node groups of every managed Kubernetes cluster",
		Resolver:    fetchNodeGroups,
		Transform:   client.TransformResource(&mk8smodel.NodeGroupOptionalResponse{}),
		Columns:     client.ChildColumns(clusterColumn),
	}
}

func fetchNodeGroups(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	cluster, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	groups, err := mk8ssdk.NewMk8sNodeGroup(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mk8sclient.ListMk8sNodeGroupsRequest{Project: c.ProjectName, ClusterName: cluster}
	return client.List(ctx, req, groups.ListMk8sNodeGroups, res, c.Skip(serviceName))
}
