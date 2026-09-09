package mclickhouse

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	mclickhouseclient "go.mws.cloud/go-sdk/service/mclickhouse/client"
	mclickhousemodel "go.mws.cloud/go-sdk/service/mclickhouse/model"
	mclickhousesdk "go.mws.cloud/go-sdk/service/mclickhouse/sdk"

	"cq-source-mws/client"
)

func ClusterUsers() *schema.Table {
	return &schema.Table{
		Name:        "mws_mclickhouse_cluster_users",
		Description: "Users of every managed ClickHouse cluster",
		Resolver:    fetchClusterUsers,
		Transform:   client.TransformResource(&mclickhousemodel.ClickhouseClusterUserResponse{}),
		Columns:     client.ChildColumns(clusterColumn),
	}
}

func fetchClusterUsers(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	cluster, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	users, err := mclickhousesdk.NewClickhouseClusterUser(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mclickhouseclient.ListClickhouseClusterUsersRequest{Project: c.ProjectName, Cluster: cluster}
	return client.List(ctx, req, users.ListClickhouseClusterUsers, res, c.Skip(serviceName))
}
