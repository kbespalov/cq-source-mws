// Package metastore holds the tables of the metadata service: the Iceberg
// catalogs a project exposes and who is allowed to use them.
package metastore

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	metastoreclient "go.mws.cloud/go-sdk/service/metastore/client"
	metastoremodel "go.mws.cloud/go-sdk/service/metastore/model"
	metastoresdk "go.mws.cloud/go-sdk/service/metastore/sdk"

	"cq-source-mws/client"
)

// serviceName is what a skipped project is logged under.
const serviceName = "metastore"

func IcebergCatalogs() *schema.Table {
	return &schema.Table{
		Name:        "mws_metastore_iceberg_catalogs",
		Description: "Iceberg catalogs of every synced project",
		Resolver:    fetchIcebergCatalogs,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&metastoremodel.IcebergCatalogOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchIcebergCatalogs(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	catalogs, err := metastoresdk.NewIceberg(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := metastoreclient.ListIcebergCatalogsRequest{Project: c.ProjectName}
	return client.List(ctx, req, catalogs.ListIcebergCatalogs, res, c.Skip(serviceName))
}
