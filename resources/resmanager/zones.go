package resmanager

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	rmclient "go.mws.cloud/go-sdk/service/rm/client"
	rmmodel "go.mws.cloud/go-sdk/service/rm/model"
	rmsdk "go.mws.cloud/go-sdk/service/rm/sdk"

	"cq-source-mws/client"
)

func Zones() *schema.Table {
	return &schema.Table{
		Name:        "mws_resmanager_zones",
		Description: "Availability zones of the installation",
		Resolver:    fetchZones,
		Multiplex:   client.GlobalMultiplex(),
		Transform:   client.TransformResource(&rmmodel.ZoneResponse{}),
	}
}

func fetchZones(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	zones, err := rmsdk.NewZone(ctx, c.SDK)
	if err != nil {
		return err
	}

	// The request can narrow to one region; left unset it answers with the
	// zones of every one.
	return client.List(ctx, rmclient.ListZonesRequest{}, zones.ListZones, res, nil)
}
