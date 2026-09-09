package secretmanager

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	secretmanagerclient "go.mws.cloud/go-sdk/service/secretmanager/client"
	secretmanagermodel "go.mws.cloud/go-sdk/service/secretmanager/model"
	secretmanagersdk "go.mws.cloud/go-sdk/service/secretmanager/sdk"

	"cq-source-mws/client"
)

func SecretVersions() *schema.Table {
	return &schema.Table{
		Name:        "mws_secretmanager_secret_versions",
		Description: "Versions of every secret",
		Resolver:    fetchSecretVersions,
		Transform:   client.TransformResource(&secretmanagermodel.SecretVersionOptionalResponse{}),
		Columns:     client.ChildColumns(secretColumn),
	}
}

func fetchSecretVersions(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	secret, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	versions, err := secretmanagersdk.NewSecretVersion(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := secretmanagerclient.ListSecretVersionsRequest{Project: c.ProjectName, Name: secret}
	return client.List(ctx, req, versions.ListSecretVersions, res, c.Skip(serviceName))
}
