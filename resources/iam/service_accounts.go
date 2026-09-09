// Package iam holds the tables of the identity service: the service accounts
// of a project and the credentials issued to each of them, plus the catalog of
// roles the whole installation shares.
//
// The credentials are listed per service account, never per project, so those
// tables are children of mws_iam_service_accounts. What they do not hold is the
// key material: the SDK hands out api keys, hmac secrets and private keys only
// redacted, so those fields get no column at all.
package iam

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	iamclient "go.mws.cloud/go-sdk/service/iam/client"
	iammodel "go.mws.cloud/go-sdk/service/iam/model"
	iamsdk "go.mws.cloud/go-sdk/service/iam/sdk"

	"cq-source-mws/client"
)

const (
	// serviceName is what a skipped project is logged under.
	serviceName = "iam"
	// serviceAccountColumn joins a credential back to the account holding it.
	serviceAccountColumn = "service_account_id"
)

func ServiceAccounts() *schema.Table {
	return &schema.Table{
		Name:        "mws_iam_service_accounts",
		Description: "Service accounts of every synced project",
		Resolver:    fetchServiceAccounts,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&iammodel.ServiceAccountResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
		Relations: []*schema.Table{
			ApiKeys(),
			AuthorizedKeys(),
			HmacKeys(),
		},
	}
}

func fetchServiceAccounts(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	accounts, err := iamsdk.NewServiceAccount(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := iamclient.ListServiceAccountRequest{Project: c.ProjectName}
	return client.List(ctx, req, accounts.ListServiceAccount, res, c.Skip(serviceName))
}
