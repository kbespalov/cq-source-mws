package mkafka

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	mkafkaclient "go.mws.cloud/go-sdk/service/mkafka/client"
	mkafkamodel "go.mws.cloud/go-sdk/service/mkafka/model"
	mkafkasdk "go.mws.cloud/go-sdk/service/mkafka/sdk"

	"cq-source-mws/client"
)

// VmTypes is the catalog of hardware a Kafka cluster can run on. The listing
// takes a project, so the table is per project even though the list itself is
// the same for everyone.
func VmTypes() *schema.Table {
	return &schema.Table{
		Name:        "mws_mkafka_vm_types",
		Description: "Kafka virtual machine types of every synced project",
		Resolver:    fetchVmTypes,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&mkafkamodel.KafkaVmTypeResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchVmTypes(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	vmTypes, err := mkafkasdk.NewKafkaVmType(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := mkafkaclient.GetKafkaVmTypesRequest{Project: c.ProjectName}
	return client.Fetch(func() ([]mkafkamodel.KafkaVmTypeResponse, error) {
		resp, err := vmTypes.GetKafkaVmTypes(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.GetItems(), nil
	}, res, c.Skip(serviceName))
}
