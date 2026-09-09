package certmanager

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	certmanagerclient "go.mws.cloud/go-sdk/service/certmanager/client"
	certmanagermodel "go.mws.cloud/go-sdk/service/certmanager/model"
	certmanagersdk "go.mws.cloud/go-sdk/service/certmanager/sdk"

	"cq-source-mws/client"
)

func RoleBindings() *schema.Table {
	return &schema.Table{
		Name:        "mws_certmanager_certificate_role_bindings",
		Description: "Who holds which role on every certificate",
		Resolver:    fetchRoleBindings,
		Transform:   client.TransformResource(&certmanagermodel.CertificateRoleBindingOptionalResponse{}),
		Columns:     client.ChildColumns(certificateColumn),
	}
}

func fetchRoleBindings(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	certificate, err := client.ParentResourceName(parent)
	if err != nil {
		return err
	}
	bindings, err := certmanagersdk.NewCertificateRoleBinding(ctx, c.SDK)
	if err != nil {
		return err
	}

	// Alone among the role binding listings, this one answers without a page
	// token, so there is nothing to drain: the response is the whole set.
	req := certmanagerclient.ListCertificateRoleBindingsRequest{Project: c.ProjectName, Name: certificate}
	return client.Fetch(func() ([]certmanagermodel.CertificateRoleBindingOptionalResponse, error) {
		resp, err := bindings.ListCertificateRoleBindings(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.Items, nil
	}, res, c.Skip(serviceName))
}
