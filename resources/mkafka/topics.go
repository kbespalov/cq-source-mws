package mkafka

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	mkafkaclient "go.mws.cloud/go-sdk/service/mkafka/client"
	mkafkamodel "go.mws.cloud/go-sdk/service/mkafka/model"
	mkafkasdk "go.mws.cloud/go-sdk/service/mkafka/sdk"

	"cq-source-mws/client"
)

func Topics() *schema.Table {
	return &schema.Table{
		Name:        "mws_mkafka_topics",
		Description: "Topics of every managed Kafka cluster",
		Resolver:    fetchTopics,
		Transform:   client.TransformResource(&mkafkamodel.KafkaTopicResponse{}),
		Columns:     client.ChildColumns(clusterColumn),
	}
}

func fetchTopics(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	cluster, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	topics, err := mkafkasdk.NewTopic(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mkafkaclient.ListKafkaTopicsRequest{Project: c.ProjectName, Cluster: cluster}
	return client.List(ctx, req, topics.ListKafkaTopics, res, c.Skip(serviceName))
}
