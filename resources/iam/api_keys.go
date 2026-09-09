package iam

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	iamclient "go.mws.cloud/go-sdk/service/iam/client"
	iammodel "go.mws.cloud/go-sdk/service/iam/model"
	iamsdk "go.mws.cloud/go-sdk/service/iam/sdk"

	"cq-source-mws/client"
)

func ApiKeys() *schema.Table {
	return &schema.Table{
		Name:        "mws_iam_api_keys",
		Description: "API keys of every service account",
		Resolver:    fetchApiKeys,
		Transform:   client.TransformResource(&iammodel.ApiKeyResponse{}),
		Columns:     client.ChildColumns(serviceAccountColumn),
	}
}

func fetchApiKeys(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	account, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	keys, err := iamsdk.NewApiKey(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := iamclient.ListApiKeyRequest{Project: c.ProjectName, ServiceAccount: account}
	return client.List(ctx, req, keys.ListApiKey, res, c.Skip(serviceName))
}
