// Package auditlogs holds the tables of the audit logs service: the collectors
// that gather a project's audit trail and the storages it is written to.
package auditlogs

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	auditlogsclient "go.mws.cloud/go-sdk/service/auditlogs/client"
	auditlogsmodel "go.mws.cloud/go-sdk/service/auditlogs/model"
	auditlogssdk "go.mws.cloud/go-sdk/service/auditlogs/sdk"

	"cq-source-mws/client"
)

// serviceName is what a skipped project is logged under.
const serviceName = "auditlogs"

func Collectors() *schema.Table {
	return &schema.Table{
		Name:        "mws_auditlogs_collectors",
		Description: "Audit log collectors of every synced project",
		Resolver:    fetchCollectors,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&auditlogsmodel.CollectorResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchCollectors(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	collectors, err := auditlogssdk.NewCollector(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := auditlogsclient.ListCollectorsRequest{Project: c.ProjectName}
	return client.List(ctx, req, collectors.ListCollectors, res, c.Skip(serviceName))
}
