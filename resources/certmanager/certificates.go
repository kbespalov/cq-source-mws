// Package certmanager holds the tables of the certificate service: the
// certificates of a project and who is allowed to use them. The certificate
// content, private key included, is fetched one at a time and is not listed.
package certmanager

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	certmanagerclient "go.mws.cloud/go-sdk/service/certmanager/client"
	certmanagermodel "go.mws.cloud/go-sdk/service/certmanager/model"
	certmanagersdk "go.mws.cloud/go-sdk/service/certmanager/sdk"

	"cq-source-mws/client"
)

const (
	// serviceName is what a skipped project is logged under.
	serviceName = "certmanager"
	// certificateColumn joins a nested table back to its certificate.
	certificateColumn = "certificate_id"
)

func Certificates() *schema.Table {
	return &schema.Table{
		Name:        "mws_certmanager_certificates",
		Description: "Certificates of every synced project",
		Resolver:    fetchCertificates,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&certmanagermodel.CertificateOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
		Relations:   []*schema.Table{RoleBindings()},
	}
}

func fetchCertificates(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	certificates, err := certmanagersdk.NewCertificate(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := certmanagerclient.ListCertificatesRequest{Project: c.ProjectName}
	return client.List(ctx, req, certificates.ListCertificates, res, c.Skip(serviceName))
}
