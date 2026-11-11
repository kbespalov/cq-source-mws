package metastore

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	metastoreclient "go.mws.cloud/go-sdk/service/metastore/client"
	metastoremodel "go.mws.cloud/go-sdk/service/metastore/model"
	metastoresdk "go.mws.cloud/go-sdk/service/metastore/sdk"

	"cq-source-mws/client"
)

// IcebergRoleBindings lists who holds which role on Iceberg catalogs. The
// listing is scoped to the project rather than to a catalog, so unlike the
// role binding tables of KMS or Secret Manager this one hangs off the project.
func IcebergRoleBindings() *schema.Table {
	return &schema.Table{
		Name:        "mws_metastore_iceberg_role_bindings",
		Description: "Who holds which role on the Iceberg catalogs of every synced project",
		Resolver:    fetchIcebergRoleBindings,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&metastoremodel.IcebergRoleBindingOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchIcebergRoleBindings(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	bindings, err := metastoresdk.NewIcebergRoleBinding(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := metastoreclient.ListIcebergRoleBindingsRequest{Project: c.ProjectName}
	return client.List(ctx, req, bindings.ListIcebergRoleBindings, res, c.Skip(serviceName))
}
