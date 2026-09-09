package iam

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	iamclient "go.mws.cloud/go-sdk/service/iam/client"
	iammodel "go.mws.cloud/go-sdk/service/iam/model"
	iamsdk "go.mws.cloud/go-sdk/service/iam/sdk"

	"cq-source-mws/client"
)

func HmacKeys() *schema.Table {
	return &schema.Table{
		Name:        "mws_iam_hmac_keys",
		Description: "HMAC keys of every service account",
		Resolver:    fetchHmacKeys,
		Transform:   client.TransformResource(&iammodel.HmacKeyResponse{}),
		Columns:     client.ChildColumns(serviceAccountColumn),
	}
}

func fetchHmacKeys(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	account, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	keys, err := iamsdk.NewServiceAccountHmacKey(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := iamclient.ListHmacKeyRequest{Project: c.ProjectName, ServiceAccount: account}
	return client.List(ctx, req, keys.ListHmacKey, res, c.Skip(serviceName))
}
