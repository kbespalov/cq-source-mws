// Package queue holds the tables of the queues service: the topics of a
// project, the partitions and access bindings of each, and the consumer
// groups reading from them.
package queue

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	queueclient "go.mws.cloud/go-sdk/service/queue/client"
	queuemodel "go.mws.cloud/go-sdk/service/queue/model"
	queuesdk "go.mws.cloud/go-sdk/service/queue/sdk"

	"cq-source-mws/client"
)

const (
	// serviceName is what a skipped project is logged under.
	serviceName = "queue"
	// topicColumn joins a nested table back to its topic.
	topicColumn = "topic_id"
)

func Topics() *schema.Table {
	return &schema.Table{
		Name:        "mws_queue_topics",
		Description: "Queues topics of every synced project",
		Resolver:    fetchTopics,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&queuemodel.TopicResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
		Relations: []*schema.Table{
			TopicPartitions(),
			TopicRoleBindings(),
		},
	}
}

func fetchTopics(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	topics, err := queuesdk.NewTopic(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := queueclient.ListTopicsRequest{Project: c.ProjectName}
	return client.List(ctx, req, topics.ListTopics, res, c.Skip(serviceName))
}
