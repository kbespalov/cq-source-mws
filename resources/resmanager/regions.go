package resmanager

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	rmclient "go.mws.cloud/go-sdk/service/rm/client"
	rmmodel "go.mws.cloud/go-sdk/service/rm/model"
	rmsdk "go.mws.cloud/go-sdk/service/rm/sdk"

	"cq-source-mws/client"
)

// Regions and Zones describe the installation itself rather than anything a
// project owns, so they are fetched once instead of per project, and a refusal
// is a real error rather than a service someone never enabled.
func Regions() *schema.Table {
	return &schema.Table{
		Name:        "mws_resmanager_regions",
		Description: "Regions of the installation",
		Resolver:    fetchRegions,
		Multiplex:   client.GlobalMultiplex(),
		Transform:   client.TransformResource(&rmmodel.RegionResponse{}),
	}
}

func fetchRegions(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	regions, err := rmsdk.NewRegion(ctx, c.SDK)
	if err != nil {
		return err
	}

	return client.List(ctx, rmclient.ListRegionsRequest{}, regions.ListRegions, res, nil)
}
