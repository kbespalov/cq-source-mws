package mpostgres

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	mpostgresclient "go.mws.cloud/go-sdk/service/mpostgres/client"
	mpostgresmodel "go.mws.cloud/go-sdk/service/mpostgres/model"
	mpostgressdk "go.mws.cloud/go-sdk/service/mpostgres/sdk"

	"cq-source-mws/client"
)

func ClusterRoleBindings() *schema.Table {
	return &schema.Table{
		Name:        "mws_mpostgres_cluster_role_bindings",
		Description: "Who holds which role on every managed PostgreSQL cluster",
		Resolver:    fetchClusterRoleBindings,
		Transform:   client.TransformResource(&mpostgresmodel.PostgresRoleBindingResponse{}),
		Columns:     client.ChildColumns(clusterColumn),
	}
}

func fetchClusterRoleBindings(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	cluster, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	bindings, err := mpostgressdk.NewPostgresRoleBinding(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mpostgresclient.ListPostgresRoleBindingsRequest{Project: c.ProjectName, Cluster: cluster}
	return client.List(ctx, req, bindings.ListPostgresRoleBindings, res, c.Skip(serviceName))
}
