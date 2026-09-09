package mpostgres

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	mpostgresclient "go.mws.cloud/go-sdk/service/mpostgres/client"
	mpostgresmodel "go.mws.cloud/go-sdk/service/mpostgres/model"
	mpostgressdk "go.mws.cloud/go-sdk/service/mpostgres/sdk"

	"cq-source-mws/client"
)

func ClusterDatabases() *schema.Table {
	return &schema.Table{
		Name:        "mws_mpostgres_cluster_databases",
		Description: "Databases of every managed PostgreSQL cluster",
		Resolver:    fetchClusterDatabases,
		Transform:   client.TransformResource(&mpostgresmodel.PostgresClusterDatabaseResponse{}),
		Columns:     client.ChildColumns(clusterColumn),
	}
}

func fetchClusterDatabases(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	cluster, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	databases, err := mpostgressdk.NewPostgresClusterDatabase(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mpostgresclient.ListPostgresClusterDatabasesRequest{Project: c.ProjectName, Cluster: cluster}
	return client.List(ctx, req, databases.ListPostgresClusterDatabases, res, c.Skip(serviceName))
}
