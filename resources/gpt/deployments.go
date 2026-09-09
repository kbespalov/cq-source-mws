package gpt

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	gptclient "go.mws.cloud/go-sdk/service/gpt/client"
	gptmodel "go.mws.cloud/go-sdk/service/gpt/model"
	gptsdk "go.mws.cloud/go-sdk/service/gpt/sdk"

	"cq-source-mws/client"
)

func Deployments() *schema.Table {
	return &schema.Table{
		Name:        "mws_gpt_deployments",
		Description: "Model deployments of every synced project",
		Resolver:    fetchDeployments,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&gptmodel.DeploymentResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchDeployments(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	deployments, err := gptsdk.NewDeployment(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := gptclient.ListDeploymentsRequest{Project: c.ProjectName}
	return client.List(ctx, req, deployments.ListDeployments, res, c.Skip(serviceName))
}
