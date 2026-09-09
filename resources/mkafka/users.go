package mkafka

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	mkafkaclient "go.mws.cloud/go-sdk/service/mkafka/client"
	mkafkamodel "go.mws.cloud/go-sdk/service/mkafka/model"
	mkafkasdk "go.mws.cloud/go-sdk/service/mkafka/sdk"

	"cq-source-mws/client"
)

func Users() *schema.Table {
	return &schema.Table{
		Name:        "mws_mkafka_users",
		Description: "Users of every managed Kafka cluster",
		Resolver:    fetchUsers,
		Transform:   client.TransformResource(&mkafkamodel.KafkaUserResponse{}),
		Columns:     client.ChildColumns(clusterColumn),
	}
}

func fetchUsers(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	cluster, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	users, err := mkafkasdk.NewKafkaUser(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mkafkaclient.ListKafkaUsersRequest{Project: c.ProjectName, Cluster: cluster}
	return client.List(ctx, req, users.ListKafkaUsers, res, c.Skip(serviceName))
}
