package mclickhouse

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	mclickhouseclient "go.mws.cloud/go-sdk/service/mclickhouse/client"
	mclickhousemodel "go.mws.cloud/go-sdk/service/mclickhouse/model"
	mclickhousesdk "go.mws.cloud/go-sdk/service/mclickhouse/sdk"

	"cq-source-mws/client"
)

func Backups() *schema.Table {
	return &schema.Table{
		Name:        "mws_mclickhouse_backups",
		Description: "Backups of every managed ClickHouse cluster",
		Resolver:    fetchBackups,
		Transform:   client.TransformResource(&mclickhousemodel.ClickhouseBackupOptionalResponse{}),
		Columns:     client.ChildColumns(clusterColumn),
	}
}

func fetchBackups(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	cluster, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	backups, err := mclickhousesdk.NewClickhouseClusterBackup(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mclickhouseclient.ListClickhouseClusterBackupsRequest{Project: c.ProjectName, Cluster: cluster}
	return client.List(ctx, req, backups.ListClickhouseClusterBackups, res, c.Skip(serviceName))
}
