// Package resmanager holds the tables of the resource manager: the projects
// the credentials can reach, the services each has enabled, and the regions
// and zones of the installation.
package resmanager

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	"go.mws.cloud/go-sdk/mws/page"
	rmclient "go.mws.cloud/go-sdk/service/rm/client"
	rmmodel "go.mws.cloud/go-sdk/service/rm/model"
	rmsdk "go.mws.cloud/go-sdk/service/rm/sdk"

	"cq-source-mws/client"
)

func Projects() *schema.Table {
	return &schema.Table{
		Name:        "mws_resmanager_projects",
		Description: "Projects of the organizations the credentials can reach",
		Resolver:    fetchProjects,
		Multiplex:   client.OrganizationMultiplex(),
		Transform:   client.TransformResource(&rmmodel.ProjectResponse{}),
		Columns:     []schema.Column{client.OrganizationColumn},
	}
}

func fetchProjects(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	// The hierarchy is already filtered by organization_ids / folder_ids /
	// project_ids. Listing again without that filter would put every reachable
	// project in the table while the rest of the sync only walked a subset.
	allowed := make(map[string]struct{})
	for _, p := range c.Hierarchy().Projects() {
		if c.OrganizationID == "" || p.Organization == c.OrganizationID {
			allowed[p.ID] = struct{}{}
		}
	}

	projects, err := rmsdk.NewProject(ctx, c.SDK)
	if err != nil {
		return err
	}

	onlyAvailable := true
	pager := page.NewPager(
		rmclient.ListProjectsV3Request{OnlyAvailableProjects: &onlyAvailable},
		projects.ListProjectsV3,
	)

	for item, err := range pager.All(ctx) {
		if err != nil {
			return err
		}
		id := item.GetMetadata().GetId()
		if id == nil {
			continue
		}
		if _, ok := allowed[id.ID()]; !ok {
			continue
		}
		res <- item
	}
	return nil
}
