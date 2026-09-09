// Package secretmanager holds the tables of the secret store: the secrets of a
// project, the versions each has gone through and who is allowed to read them.
// Only the metadata is listed; the payload of a secret is fetched one at a
// time and never lands here.
package secretmanager

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	secretmanagerclient "go.mws.cloud/go-sdk/service/secretmanager/client"
	secretmanagermodel "go.mws.cloud/go-sdk/service/secretmanager/model"
	secretmanagersdk "go.mws.cloud/go-sdk/service/secretmanager/sdk"

	"cq-source-mws/client"
)

const (
	// serviceName is what a skipped project is logged under.
	serviceName = "secretmanager"
	// secretColumn joins a nested table back to its secret.
	secretColumn = "secret_id"
)

func Secrets() *schema.Table {
	return &schema.Table{
		Name:        "mws_secretmanager_secrets",
		Description: "Secrets of every synced project",
		Resolver:    fetchSecrets,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&secretmanagermodel.SecretOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
		Relations: []*schema.Table{
			SecretVersions(),
			SecretRoleBindings(),
		},
	}
}

func fetchSecrets(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	secrets, err := secretmanagersdk.NewSecret(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := secretmanagerclient.ListSecretsRequest{Project: c.ProjectName}
	return client.List(ctx, req, secrets.ListSecrets, res, c.Skip(serviceName))
}
