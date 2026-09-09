// Package mk8s holds the tables of the managed Kubernetes service: the
// clusters of a project and the node groups of each, plus the release channels
// a cluster can follow and the versions each channel offers.
package mk8s

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	mk8sclient "go.mws.cloud/go-sdk/service/mk8s/client"
	mk8smodel "go.mws.cloud/go-sdk/service/mk8s/model"
	mk8ssdk "go.mws.cloud/go-sdk/service/mk8s/sdk"

	"cq-source-mws/client"
)

const (
	// serviceName is what a skipped project is logged under.
	serviceName = "mk8s"
	// clusterColumn joins a nested table back to its cluster.
	clusterColumn = "cluster_id"
)

func Clusters() *schema.Table {
	return &schema.Table{
		Name:        "mws_mk8s_clusters",
		Description: "Managed Kubernetes clusters of every synced project",
		Resolver:    fetchClusters,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&mk8smodel.ClusterOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
		Relations:   []*schema.Table{NodeGroups()},
	}
}

func fetchClusters(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	clusters, err := mk8ssdk.NewMk8sCluster(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mk8sclient.ListMk8sClustersRequest{Project: c.ProjectName}
	return client.List(ctx, req, clusters.ListMk8sClusters, res, c.Skip(serviceName))
}
