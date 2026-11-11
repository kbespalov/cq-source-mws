package queue

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	commonmodel "go.mws.cloud/go-sdk/service/common/model"
	queueclient "go.mws.cloud/go-sdk/service/queue/client"
	queuesdk "go.mws.cloud/go-sdk/service/queue/sdk"

	"cq-source-mws/client"
)

func TopicRoleBindings() *schema.Table {
	return &schema.Table{
		Name:        "mws_queue_topic_role_bindings",
		Description: "Who holds which role on every queues topic",
		Resolver:    fetchTopicRoleBindings,
		Transform:   client.TransformResource(&commonmodel.CommonRoleBindingOptionalResponse{}),
		Columns:     client.ChildColumns(topicColumn),
	}
}

func fetchTopicRoleBindings(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	topic, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	bindings, err := queuesdk.NewTopicRoleBinding(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := queueclient.ListTopicRoleBindingsRequest{Project: c.ProjectName, Topic: topic}
	return client.List(ctx, req, bindings.ListTopicRoleBindings, res, c.Skip(serviceName))
}
