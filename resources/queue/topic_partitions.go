package queue

import (
	"context"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/cloudquery/plugin-sdk/v4/schema"
	queueclient "go.mws.cloud/go-sdk/service/queue/client"
	queuesdk "go.mws.cloud/go-sdk/service/queue/sdk"

	"cq-source-mws/client"
)

// A partition has no identity of its own beyond its number: no kind, no
// metadata, just where the topic's data starts, ends and how big it is. Like
// the Kubernetes release channels, its columns are spelled out and the topic
// it belongs to joins the key, because partition 0 exists in every topic.
func TopicPartitions() *schema.Table {
	columns := client.ChildColumns(topicColumn)
	for i := range columns {
		if columns[i].Name == topicColumn {
			columns[i].PrimaryKey = true
			columns[i].NotNull = true
		}
	}
	return &schema.Table{
		Name:        "mws_queue_topic_partitions",
		Description: "Partitions of every queues topic",
		Resolver:    fetchTopicPartitions,
		Columns: append(columns,
			schema.Column{
				Name:       "partition",
				Type:       arrow.PrimitiveTypes.Int64,
				Resolver:   schema.PathResolver("Partition"),
				PrimaryKey: true,
				NotNull:    true,
			},
			schema.Column{
				Name:     "offsets_earliest",
				Type:     arrow.PrimitiveTypes.Int64,
				Resolver: schema.PathResolver("Offsets.Earliest"),
			},
			schema.Column{
				Name:     "offsets_latest",
				Type:     arrow.PrimitiveTypes.Int64,
				Resolver: schema.PathResolver("Offsets.Latest"),
			},
			schema.Column{
				Name:     "size_bytes",
				Type:     arrow.PrimitiveTypes.Int64,
				Resolver: schema.PathResolver("SizeBytes"),
			},
		),
	}
}

func fetchTopicPartitions(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	topic, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	topics, err := queuesdk.NewTopic(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := queueclient.ListTopicPartitionsRequest{Project: c.ProjectName, Topic: topic}
	return client.List(ctx, req, topics.ListTopicPartitions, res, c.Skip(serviceName))
}
