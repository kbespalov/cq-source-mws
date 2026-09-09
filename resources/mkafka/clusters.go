// Package mkafka holds the tables of the managed Kafka service: the clusters
// of a project, along with the topics, users and connectors of each.
package mkafka

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	mkafkaclient "go.mws.cloud/go-sdk/service/mkafka/client"
	mkafkamodel "go.mws.cloud/go-sdk/service/mkafka/model"
	mkafkasdk "go.mws.cloud/go-sdk/service/mkafka/sdk"

	"cq-source-mws/client"
)

const (
	// serviceName is what a skipped project is logged under.
	serviceName = "mkafka"
	// clusterColumn joins a nested table back to its cluster.
	clusterColumn = "cluster_id"
)

func Clusters() *schema.Table {
	return &schema.Table{
		Name:        "mws_mkafka_clusters",
		Description: "Managed Kafka clusters of every synced project",
		Resolver:    fetchClusters,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&mkafkamodel.KafkaClusterResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
		Relations: []*schema.Table{
			Topics(),
			Users(),
			Connectors(),
		},
	}
}

func fetchClusters(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	clusters, err := mkafkasdk.NewKafka(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mkafkaclient.ListKafkaClustersRequest{Project: c.ProjectName}
	return client.List(ctx, req, clusters.ListKafkaClusters, res, c.Skip(serviceName))
}
