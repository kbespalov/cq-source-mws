package iam

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	iamclient "go.mws.cloud/go-sdk/service/iam/client"
	iammodel "go.mws.cloud/go-sdk/service/iam/model"
	iamsdk "go.mws.cloud/go-sdk/service/iam/sdk"

	"cq-source-mws/client"
)

func AuthorizedKeys() *schema.Table {
	return &schema.Table{
		Name:        "mws_iam_authorized_keys",
		Description: "Authorized keys of every service account",
		Resolver:    fetchAuthorizedKeys,
		Transform:   client.TransformResource(&iammodel.AuthorizedKeyOptionalResponse{}),
		Columns:     client.ChildColumns(serviceAccountColumn),
	}
}

func fetchAuthorizedKeys(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	account, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	keys, err := iamsdk.NewAuthorizedKey(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := iamclient.ListAuthorizedKeyRequest{Project: c.ProjectName, ServiceAccount: account}
	return client.List(ctx, req, keys.ListAuthorizedKey, res, c.Skip(serviceName))
}
