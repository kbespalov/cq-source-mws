package kms

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	kmsclient "go.mws.cloud/go-sdk/service/kms/client"
	kmsmodel "go.mws.cloud/go-sdk/service/kms/model"
	kmssdk "go.mws.cloud/go-sdk/service/kms/sdk"

	"cq-source-mws/client"
)

func CryptoKeyRoleBindings() *schema.Table {
	return &schema.Table{
		Name:        "mws_kms_crypto_key_role_bindings",
		Description: "Who holds which role on every crypto key",
		Resolver:    fetchCryptoKeyRoleBindings,
		Transform:   client.TransformResource(&kmsmodel.CryptoKeyRoleBindingOptionalResponse{}),
		Columns:     client.ChildColumns(keyColumn),
	}
}

func fetchCryptoKeyRoleBindings(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	key, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	bindings, err := kmssdk.NewCryptoKeyRoleBinding(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := kmsclient.ListCryptoKeyRoleBindingsRequest{Project: c.ProjectName, Key: key}
	return client.List(ctx, req, bindings.ListCryptoKeyRoleBindings, res, c.Skip(serviceName))
}
