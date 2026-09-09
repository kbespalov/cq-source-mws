package secretmanager

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	secretmanagerclient "go.mws.cloud/go-sdk/service/secretmanager/client"
	secretmanagermodel "go.mws.cloud/go-sdk/service/secretmanager/model"
	secretmanagersdk "go.mws.cloud/go-sdk/service/secretmanager/sdk"

	"cq-source-mws/client"
)

func SecretRoleBindings() *schema.Table {
	return &schema.Table{
		Name:        "mws_secretmanager_secret_role_bindings",
		Description: "Who holds which role on every secret",
		Resolver:    fetchSecretRoleBindings,
		Transform:   client.TransformResource(&secretmanagermodel.SecretRoleBindingOptionalResponse{}),
		Columns:     client.ChildColumns(secretColumn),
	}
}

func fetchSecretRoleBindings(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	secret, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	bindings, err := secretmanagersdk.NewSecretRoleBinding(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := secretmanagerclient.ListRoleBindingsRequest{Project: c.ProjectName, Name: secret}
	return client.List(ctx, req, bindings.ListRoleBindings, res, c.Skip(serviceName))
}
