package compute

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	computeclient "go.mws.cloud/go-sdk/service/compute/client"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	computesdk "go.mws.cloud/go-sdk/service/compute/sdk"

	"cq-source-mws/client"
)

func DiskBackups() *schema.Table {
	return &schema.Table{
		Name:        "mws_compute_disk_backups",
		Description: "Disk backups of every synced project",
		Resolver:    fetchDiskBackups,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&computemodel.DiskBackupOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchDiskBackups(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	backups, err := computesdk.NewDiskBackup(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := computeclient.ListDiskBackupsRequest{Project: c.ProjectName}
	return client.List(ctx, req, backups.ListDiskBackups, res, c.Skip(serviceName))
}
