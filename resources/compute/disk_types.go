package compute

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	computeclient "go.mws.cloud/go-sdk/service/compute/client"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	computesdk "go.mws.cloud/go-sdk/service/compute/sdk"

	"cq-source-mws/client"
)

// DiskTypes is the catalog of storage a disk can be created on. Like the
// virtual machine types it is global, and unlike every other resource here it
// has no status: a catalog entry has nothing to observe.
func DiskTypes() *schema.Table {
	return &schema.Table{
		Name:        "mws_compute_disk_types",
		Description: "Catalog of disk types",
		Resolver:    fetchDiskTypes,
		Multiplex:   client.GlobalMultiplex(),
		Transform:   client.TransformResource(&computemodel.DiskTypeResponse{}),
	}
}

func fetchDiskTypes(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	diskTypes, err := computesdk.NewDiskType(ctx, c.SDK)
	if err != nil {
		return err
	}

	return client.List(ctx, computeclient.ListDiskTypesRequest{}, diskTypes.ListDiskTypes, res, nil)
}
