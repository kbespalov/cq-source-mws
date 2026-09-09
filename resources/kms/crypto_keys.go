// Package kms holds the tables of the key management service: the keys of a
// project, the versions each key has rotated through and who is allowed to use
// them. Key material never leaves the service, so none of it is here.
package kms

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	kmsclient "go.mws.cloud/go-sdk/service/kms/client"
	kmsmodel "go.mws.cloud/go-sdk/service/kms/model"
	kmssdk "go.mws.cloud/go-sdk/service/kms/sdk"

	"cq-source-mws/client"
)

const (
	// serviceName is what a skipped project is logged under.
	serviceName = "kms"
	// keyColumn joins a nested table back to its key.
	keyColumn = "crypto_key_id"
)

func CryptoKeys() *schema.Table {
	return &schema.Table{
		Name:        "mws_kms_crypto_keys",
		Description: "Crypto keys of every synced project",
		Resolver:    fetchCryptoKeys,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&kmsmodel.CryptoKeyOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
		Relations: []*schema.Table{
			CryptoKeyVersions(),
			CryptoKeyRoleBindings(),
		},
	}
}

func fetchCryptoKeys(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	keys, err := kmssdk.NewCryptoKey(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := kmsclient.ListCryptoKeysRequest{Project: c.ProjectName}
	return client.List(ctx, req, keys.ListCryptoKeys, res, c.Skip(serviceName))
}
