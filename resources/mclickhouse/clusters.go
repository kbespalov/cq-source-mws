// Package mclickhouse holds the tables of the managed ClickHouse service: the
// clusters of a project, their users and their backups.
package mclickhouse

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	mclickhouseclient "go.mws.cloud/go-sdk/service/mclickhouse/client"
	mclickhousemodel "go.mws.cloud/go-sdk/service/mclickhouse/model"
	mclickhousesdk "go.mws.cloud/go-sdk/service/mclickhouse/sdk"

	"cq-source-mws/client"
)

const (
	// serviceName is what a skipped project is logged under.
	serviceName = "mclickhouse"
	// clusterColumn joins a nested table back to its cluster.
	clusterColumn = "cluster_id"
)

func Clusters() *schema.Table {
	return &schema.Table{
		Name:        "mws_mclickhouse_clusters",
		Description: "Managed ClickHouse clusters of every synced project",
		Resolver:    fetchClusters,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&mclickhousemodel.ClickhouseClusterOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
		Relations: []*schema.Table{
			ClusterUsers(),
			Backups(),
		},
	}
}

func fetchClusters(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	clusters, err := mclickhousesdk.NewClickhouseCluster(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mclickhouseclient.ListClickhouseClustersRequest{Project: c.ProjectName}
	return client.List(ctx, req, clusters.ListClickhouseClusters, res, c.Skip(serviceName))
}
