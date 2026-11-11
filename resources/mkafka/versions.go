package mkafka

import (
	"context"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/cloudquery/plugin-sdk/v4/schema"
	mkafkaclient "go.mws.cloud/go-sdk/service/mkafka/client"
	mkafkamodel "go.mws.cloud/go-sdk/service/mkafka/model"
	mkafkasdk "go.mws.cloud/go-sdk/service/mkafka/sdk"

	"cq-source-mws/client"
)

// A Kafka version is a plain version string with a deprecation flag: no kind,
// no metadata, so its columns are spelled out and the project joins the key,
// because every project answers with the same handful of versions.
func Versions() *schema.Table {
	return &schema.Table{
		Name:        "mws_mkafka_versions",
		Description: "Kafka versions offered to every synced project",
		Resolver:    fetchVersions,
		Multiplex:   client.ProjectMultiplex(),
		Columns: append(client.ProjectKeyColumns(),
			schema.Column{
				Name:       "version",
				Type:       arrow.BinaryTypes.String,
				Resolver:   schema.PathResolver("Version"),
				PrimaryKey: true,
				NotNull:    true,
			},
			schema.Column{
				Name:     "deprecated",
				Type:     arrow.FixedWidthTypes.Boolean,
				Resolver: schema.PathResolver("Deprecated"),
			},
		),
	}
}

func fetchVersions(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	versions, err := mkafkasdk.NewKafkaVersion(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mkafkaclient.GetAvailableVersionsForProjectRequest{Project: c.ProjectName}
	return client.Fetch(func() ([]mkafkamodel.KafkaVersionResponse, error) {
		return versions.GetAvailableVersionsForProject(ctx, req)
	}, res, c.Skip(serviceName))
}
