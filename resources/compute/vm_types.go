package compute

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	computeclient "go.mws.cloud/go-sdk/service/compute/client"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	computesdk "go.mws.cloud/go-sdk/service/compute/sdk"

	"cq-source-mws/client"
)

// VmTypes is the catalog of hardware a virtual machine can be built on. It is
// the same list for everyone, so it is fetched once rather than per project,
// and it carries no hierarchy columns: it belongs to no project.
func VmTypes() *schema.Table {
	return &schema.Table{
		Name:        "mws_compute_vm_types",
		Description: "Catalog of virtual machine types",
		Resolver:    fetchVmTypes,
		Multiplex:   client.GlobalMultiplex(),
		Transform:   client.TransformResource(&computemodel.VmTypeOptionalResponse{}),
	}
}

func fetchVmTypes(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	vmTypes, err := computesdk.NewVmType(ctx, c.SDK)
	if err != nil {
		return err
	}

	return client.List(ctx, computeclient.ListVmTypesRequest{}, vmTypes.ListVmTypes, res, nil)
}
