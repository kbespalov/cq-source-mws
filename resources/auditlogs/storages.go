package auditlogs

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	auditlogsclient "go.mws.cloud/go-sdk/service/auditlogs/client"
	auditlogsmodel "go.mws.cloud/go-sdk/service/auditlogs/model"
	auditlogssdk "go.mws.cloud/go-sdk/service/auditlogs/sdk"

	"cq-source-mws/client"
)

func Storages() *schema.Table {
	return &schema.Table{
		Name:        "mws_auditlogs_storages",
		Description: "Audit log storages of every synced project",
		Resolver:    fetchStorages,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&auditlogsmodel.StorageResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchStorages(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	storages, err := auditlogssdk.NewStorage(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := auditlogsclient.ListStoragesRequest{Project: c.ProjectName}
	return client.List(ctx, req, storages.ListStorages, res, c.Skip(serviceName))
}
