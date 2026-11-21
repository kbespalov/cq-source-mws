package mclickhouse

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	mclickhouseclient "go.mws.cloud/go-sdk/service/mclickhouse/client"
	mclickhousemodel "go.mws.cloud/go-sdk/service/mclickhouse/model"
	mclickhousesdk "go.mws.cloud/go-sdk/service/mclickhouse/sdk"

	"cq-source-mws/client"
)

// VmTypes is the catalog of hardware a ClickHouse cluster can run on. The
// listing takes a project, so the table is per project even though the list
// itself is the same for everyone.
func VmTypes() *schema.Table {
	return &schema.Table{
		Name:        "mws_mclickhouse_vm_types",
		Description: "ClickHouse virtual machine types of every synced project",
		Resolver:    fetchVmTypes,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&mclickhousemodel.ClickhouseVmTypeResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchVmTypes(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	vmTypes, err := mclickhousesdk.NewVmType(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mclickhouseclient.GetClickhouseClusterVmTypesRequest{Project: c.ProjectName}
	return client.Fetch(func() ([]mclickhousemodel.ClickhouseVmTypeResponse, error) {
		resp, err := vmTypes.GetClickhouseClusterVmTypes(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.GetItems(), nil
	}, res, c.Skip(serviceName))
}
