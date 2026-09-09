package mpostgres

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	mpostgresclient "go.mws.cloud/go-sdk/service/mpostgres/client"
	mpostgresmodel "go.mws.cloud/go-sdk/service/mpostgres/model"
	mpostgressdk "go.mws.cloud/go-sdk/service/mpostgres/sdk"

	"cq-source-mws/client"
)

func ClusterUsers() *schema.Table {
	return &schema.Table{
		Name:        "mws_mpostgres_cluster_users",
		Description: "Users of every managed PostgreSQL cluster",
		Resolver:    fetchClusterUsers,
		Transform:   client.TransformResource(&mpostgresmodel.PostgresClusterUserResponse{}),
		Columns:     client.ChildColumns(clusterColumn),
	}
}

func fetchClusterUsers(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	cluster, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	users, err := mpostgressdk.NewPostgresClusterUser(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mpostgresclient.ListPostgresClusterUsersRequest{Project: c.ProjectName, Cluster: cluster}
	return client.List(ctx, req, users.ListPostgresClusterUsers, res, c.Skip(serviceName))
}
