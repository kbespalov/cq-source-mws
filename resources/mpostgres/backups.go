package mpostgres

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	mpostgresclient "go.mws.cloud/go-sdk/service/mpostgres/client"
	mpostgresmodel "go.mws.cloud/go-sdk/service/mpostgres/model"
	mpostgressdk "go.mws.cloud/go-sdk/service/mpostgres/sdk"

	"cq-source-mws/client"
)

func Backups() *schema.Table {
	return &schema.Table{
		Name:        "mws_mpostgres_backups",
		Description: "Backups of every managed PostgreSQL cluster",
		Resolver:    fetchBackups,
		Transform:   client.TransformResource(&mpostgresmodel.PostgresBackupResponse{}),
		Columns:     client.ChildColumns(clusterColumn),
	}
}

func fetchBackups(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	cluster, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	backups, err := mpostgressdk.NewPostgresBackup(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mpostgresclient.ListPostgresBackupRequest{Project: c.ProjectName, Cluster: cluster}
	return client.List(ctx, req, backups.ListPostgresBackup, res, c.Skip(serviceName))
}
