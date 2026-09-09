// Package mpostgres holds the tables of the managed PostgreSQL service: the
// clusters of a project, and what lives inside each of them.
//
// The API offers no way to ask a project for its databases or its users, only
// a cluster for its own, so those tables are children of mws_mpostgres_clusters
// the way VPC nests under a network.
package mpostgres

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	mpostgresclient "go.mws.cloud/go-sdk/service/mpostgres/client"
	mpostgresmodel "go.mws.cloud/go-sdk/service/mpostgres/model"
	mpostgressdk "go.mws.cloud/go-sdk/service/mpostgres/sdk"

	"cq-source-mws/client"
)

const (
	// serviceName is what a skipped project is logged under.
	serviceName = "mpostgres"
	// clusterColumn joins a nested table back to its cluster.
	clusterColumn = "cluster_id"
)

func Clusters() *schema.Table {
	return &schema.Table{
		Name:        "mws_mpostgres_clusters",
		Description: "Managed PostgreSQL clusters of every synced project",
		Resolver:    fetchClusters,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&mpostgresmodel.PostgresClusterResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
		Relations: []*schema.Table{
			ClusterUsers(),
			ClusterDatabases(),
			Backups(),
		},
	}
}

func fetchClusters(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	clusters, err := mpostgressdk.NewPostgresCluster(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mpostgresclient.ListPostgresClustersRequest{Project: c.ProjectName}
	return client.List(ctx, req, clusters.ListPostgresClusters, res, c.Skip(serviceName))
}
