package mkafka

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	mkafkaclient "go.mws.cloud/go-sdk/service/mkafka/client"
	mkafkamodel "go.mws.cloud/go-sdk/service/mkafka/model"
	mkafkasdk "go.mws.cloud/go-sdk/service/mkafka/sdk"

	"cq-source-mws/client"
)

func Connectors() *schema.Table {
	return &schema.Table{
		Name:        "mws_mkafka_connectors",
		Description: "Connectors of every managed Kafka cluster",
		Resolver:    fetchConnectors,
		Transform:   client.TransformResource(&mkafkamodel.KafkaConnectorOptionalResponse{}),
		Columns:     client.ChildColumns(clusterColumn),
	}
}

func fetchConnectors(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	cluster, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	connectors, err := mkafkasdk.NewKafkaConnector(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mkafkaclient.ListKafkaConnectorsRequest{Project: c.ProjectName, Cluster: cluster}
	return client.List(ctx, req, connectors.ListKafkaConnectors, res, c.Skip(serviceName))
}
