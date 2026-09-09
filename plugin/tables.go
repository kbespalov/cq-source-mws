package plugin

import (
	"github.com/cloudquery/plugin-sdk/v4/schema"
	"github.com/cloudquery/plugin-sdk/v4/transformers"

	"cq-source-mws/resources/certmanager"
	"cq-source-mws/resources/compute"
	"cq-source-mws/resources/gpt"
	"cq-source-mws/resources/iam"
	"cq-source-mws/resources/kms"
	"cq-source-mws/resources/mclickhouse"
	"cq-source-mws/resources/mk8s"
	"cq-source-mws/resources/mkafka"
	"cq-source-mws/resources/mpostgres"
	"cq-source-mws/resources/nlb"
	"cq-source-mws/resources/resmanager"
	"cq-source-mws/resources/secretmanager"
	"cq-source-mws/resources/vpc"
)

func tables() schema.Tables {
	tables := schema.Tables{
		resmanager.Projects(),
		compute.VirtualMachines(),
		compute.Disks(),
		compute.Snapshots(),
		compute.Images(),
		compute.DiskBackups(),
		compute.VmTypes(),
		compute.DiskTypes(),
		vpc.Networks(),
		vpc.NatGateways(),
		vpc.ExternalAddresses(),
		nlb.LoadBalancers(),
		mpostgres.Clusters(),
		mclickhouse.Clusters(),
		mkafka.Clusters(),
		mk8s.Clusters(),
		mk8s.ReleaseChannels(),
		iam.ServiceAccounts(),
		iam.GlobalRoles(),
		kms.CryptoKeys(),
		secretmanager.Secrets(),
		certmanager.Certificates(),
		gpt.Models(),
		gpt.Deployments(),
		resmanager.EnabledServices(),
		resmanager.Regions(),
		resmanager.Zones(),
	}
	if err := transformers.TransformTables(tables); err != nil {
		panic(err)
	}
	for _, t := range tables {
		schema.AddCqIDs(t)
	}
	return tables
}
