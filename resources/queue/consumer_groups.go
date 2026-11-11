package queue

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	queueclient "go.mws.cloud/go-sdk/service/queue/client"
	queuemodel "go.mws.cloud/go-sdk/service/queue/model"
	queuesdk "go.mws.cloud/go-sdk/service/queue/sdk"

	"cq-source-mws/client"
)

func ConsumerGroups() *schema.Table {
	return &schema.Table{
		Name:        "mws_queue_consumer_groups",
		Description: "Consumer groups of every synced project",
		Resolver:    fetchConsumerGroups,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&queuemodel.ConsumerGroupResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchConsumerGroups(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	groups, err := queuesdk.NewConsumerGroup(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := queueclient.ListGroupsRequest{Project: c.ProjectName}
	return client.List(ctx, req, groups.ListGroups, res, c.Skip(serviceName))
}
