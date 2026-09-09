package compute

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	computeclient "go.mws.cloud/go-sdk/service/compute/client"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	computesdk "go.mws.cloud/go-sdk/service/compute/sdk"

	"cq-source-mws/client"
)

func Snapshots() *schema.Table {
	return &schema.Table{
		Name:        "mws_compute_snapshots",
		Description: "Disk snapshots of every synced project",
		Resolver:    fetchSnapshots,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&computemodel.SnapshotOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchSnapshots(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	snapshots, err := computesdk.NewSnapshot(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := computeclient.ListSnapshotsRequest{Project: c.ProjectName}
	return client.List(ctx, req, snapshots.ListSnapshots, res, c.Skip(serviceName))
}
