// Package compute holds the tables of the compute service: the virtual
// machines of a project, the storage attached to them and the images and
// snapshots they are built from, plus the catalogs of hardware to pick from.
package compute

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	computeclient "go.mws.cloud/go-sdk/service/compute/client"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	computesdk "go.mws.cloud/go-sdk/service/compute/sdk"

	"cq-source-mws/client"
)

// serviceName is what a skipped project is logged under. A project that never
// enabled compute answers 403 to all of these listings.
const serviceName = "compute"

func Disks() *schema.Table {
	return &schema.Table{
		Name:        "mws_compute_disks",
		Description: "Disks of every synced project",
		Resolver:    fetchDisks,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&computemodel.DiskOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchDisks(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	disks, err := computesdk.NewDisk(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := computeclient.ListDisksRequest{Project: c.ProjectName}
	return client.List(ctx, req, disks.ListDisks, res, c.Skip(serviceName))
}
