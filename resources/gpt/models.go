// Package gpt holds the tables of the inference service: the models a project
// may run and the deployments actually serving them.
package gpt

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	gptclient "go.mws.cloud/go-sdk/service/gpt/client"
	gptmodel "go.mws.cloud/go-sdk/service/gpt/model"
	gptsdk "go.mws.cloud/go-sdk/service/gpt/sdk"

	"cq-source-mws/client"
)

// serviceName is what a skipped project is logged under.
const serviceName = "gpt"

func Models() *schema.Table {
	return &schema.Table{
		Name:        "mws_gpt_models",
		Description: "Models every synced project may run",
		Resolver:    fetchModels,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&gptmodel.ModelResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchModels(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	models, err := gptsdk.NewModel(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := gptclient.ListModelsRequest{Project: c.ProjectName}
	return client.List(ctx, req, models.ListModels, res, c.Skip(serviceName))
}
