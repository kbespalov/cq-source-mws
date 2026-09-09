package kms

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	kmsclient "go.mws.cloud/go-sdk/service/kms/client"
	kmsmodel "go.mws.cloud/go-sdk/service/kms/model"
	kmssdk "go.mws.cloud/go-sdk/service/kms/sdk"

	"cq-source-mws/client"
)

func CryptoKeyVersions() *schema.Table {
	return &schema.Table{
		Name:        "mws_kms_crypto_key_versions",
		Description: "Versions of every crypto key",
		Resolver:    fetchCryptoKeyVersions,
		Transform:   client.TransformResource(&kmsmodel.CryptoKeyVersionOptionalResponse{}),
		Columns:     client.ChildColumns(keyColumn),
	}
}

func fetchCryptoKeyVersions(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	key, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	versions, err := kmssdk.NewCryptoKeyVersion(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := kmsclient.ListCryptoKeyVersionsRequest{Project: c.ProjectName, Key: key}
	return client.List(ctx, req, versions.ListCryptoKeyVersions, res, c.Skip(serviceName))
}
