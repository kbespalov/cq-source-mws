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

// releaseChannelColumn is the one parent link that cannot be an id, because a
// release channel has none. Its name is what the listing takes as its path
// parameter and what identifies the parent row, so the child borrows that.
const releaseChannelColumn = "release_channel"

func Versions() *schema.Table {
	return &schema.Table{
		Name:        "mws_mk8s_versions",
		Description: "Kubernetes versions offered by every release channel",
		Resolver:    fetchVersions,
		Columns: append(client.ProjectKeyColumns(),
			schema.Column{
				Name:       releaseChannelColumn,
				Type:       arrow.BinaryTypes.String,
				Resolver:   schema.ParentColumnResolver("name"),
				PrimaryKey: true,
				NotNull:    true,
			},
			nameColumn(),
			descriptionColumn(),
			schema.Column{
				Name:     "default",
				Type:     arrow.FixedWidthTypes.Boolean,
				Resolver: schema.PathResolver("Default"),
			},
		),
	}
}

func fetchVersions(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	channel, err := client.ParentColumnValue(parent, "name")
	if err != nil {
		return err
	}
	versions, err := mk8ssdk.NewMk8sVersion(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mk8sclient.ListMk8sVersionsRequest{Project: c.ProjectName, ReleaseChannelName: channel}
	return client.Fetch(func() ([]mk8smodel.VersionResponse, error) {
		resp, err := versions.ListMk8sVersions(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.Items, nil
	}, res, c.Skip(serviceName))
}
