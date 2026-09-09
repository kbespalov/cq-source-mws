package mk8s

import (
	"context"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/cloudquery/plugin-sdk/v4/schema"
	mk8sclient "go.mws.cloud/go-sdk/service/mk8s/client"
	mk8smodel "go.mws.cloud/go-sdk/service/mk8s/model"
	mk8ssdk "go.mws.cloud/go-sdk/service/mk8s/sdk"

	"cq-source-mws/client"
)

// A release channel is the exception to how everything else in the platform is
// shaped: no kind, no metadata, no identifier, just a name and whether it can
// be picked. There is nothing for TransformResource to build a table out of,
// so these two tables spell out their columns, and the project is part of the
// key because every project answers with the same handful of channel names.
//
// The listings take no page token either: a project has a few channels and a
// channel a few versions, and the API returns them all at once.

// nameColumn is the identity of a channel and of a version alike.
func nameColumn() schema.Column {
	return schema.Column{
		Name:       "name",
		Type:       arrow.BinaryTypes.String,
		Resolver:   schema.PathResolver("Name"),
		PrimaryKey: true,
		NotNull:    true,
	}
}

func descriptionColumn() schema.Column {
	return schema.Column{
		Name:     "description",
		Type:     arrow.BinaryTypes.String,
		Resolver: schema.PathResolver("Description"),
	}
}

func ReleaseChannels() *schema.Table {
	return &schema.Table{
		Name:        "mws_mk8s_release_channels",
		Description: "Kubernetes release channels offered to every synced project",
		Resolver:    fetchReleaseChannels,
		Multiplex:   client.ProjectMultiplex(),
		Columns: append(client.ProjectKeyColumns(),
			nameColumn(),
			descriptionColumn(),
			schema.Column{
				Name:     "enabled",
				Type:     arrow.FixedWidthTypes.Boolean,
				Resolver: schema.PathResolver("Enabled"),
			},
		),
		Relations: []*schema.Table{Versions()},
	}
}

func fetchReleaseChannels(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	channels, err := mk8ssdk.NewMk8sReleaseChannel(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mk8sclient.ListMk8sReleaseChannelsRequest{Project: c.ProjectName}
	return client.Fetch(func() ([]mk8smodel.ReleaseChannelResponse, error) {
		resp, err := channels.ListMk8sReleaseChannels(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.Items, nil
	}, res, c.Skip(serviceName))
}
