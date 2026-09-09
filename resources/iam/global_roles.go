package iam

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	iamclient "go.mws.cloud/go-sdk/service/iam/client"
	iammodel "go.mws.cloud/go-sdk/service/iam/model"
	iamsdk "go.mws.cloud/go-sdk/service/iam/sdk"

	"cq-source-mws/client"
)

// GlobalRoles is the catalog of roles the installation offers. It is the same
// for every project, so it is fetched once rather than per project, and a
// refusal here is a real error rather than a project that never enabled iam.
func GlobalRoles() *schema.Table {
	return &schema.Table{
		Name:        "mws_iam_global_roles",
		Description: "Roles the installation offers",
		Resolver:    fetchGlobalRoles,
		Multiplex:   client.GlobalMultiplex(),
		Transform:   client.TransformResource(&iammodel.GlobalRoleV2Response{}),
	}
}

func fetchGlobalRoles(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	roles, err := iamsdk.NewRole(ctx, c.SDK)
	if err != nil {
		return err
	}

	// The listing insists on a type and rejects the request without one. Of
	// the two it accepts, public holds the roles a project can actually
	// grant; backoffice is the operator's own, and a row from it would be
	// indistinguishable here from a role anyone could use.
	roleType := "public"
	req := iamclient.ListGlobalRoleV2Request{Type: &roleType}
	return client.List(ctx, req, roles.ListGlobalRoleV2, res, nil)
}
