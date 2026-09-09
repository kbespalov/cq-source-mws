package compute

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	computeclient "go.mws.cloud/go-sdk/service/compute/client"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	computesdk "go.mws.cloud/go-sdk/service/compute/sdk"

	"cq-source-mws/client"
)

func VirtualMachines() *schema.Table {
	return &schema.Table{
		Name:        "mws_compute_virtual_machines",
		Description: "Virtual machines of every synced project",
		Resolver:    fetchVirtualMachines,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&computemodel.VirtualMachineOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchVirtualMachines(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	vms, err := computesdk.NewVirtualMachine(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := computeclient.ListVirtualMachinesRequest{Project: c.ProjectName}
	return client.List(ctx, req, vms.ListVirtualMachines, res, c.Skip(serviceName))
}
