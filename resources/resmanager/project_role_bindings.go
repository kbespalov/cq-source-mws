package resmanager

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	commonmodel "go.mws.cloud/go-sdk/service/common/model"
	rmclient "go.mws.cloud/go-sdk/service/rm/client"
	rmsdk "go.mws.cloud/go-sdk/service/rm/sdk"

	"cq-source-mws/client"
)

// ProjectRoleBindings lists who holds which role on every synced project. The
// listing is scoped to the project itself, so the table hangs off the project
// the way the service role binding tables hang off their resource.
func ProjectRoleBindings() *schema.Table {
	return &schema.Table{
		Name:        "mws_resmanager_project_role_bindings",
		Description: "Who holds which role on every synced project",
		Resolver:    fetchProjectRoleBindings,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&commonmodel.CommonRoleBindingOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchProjectRoleBindings(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	bindings, err := rmsdk.NewProjectRoleBinding(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := rmclient.ListProjectRoleBindingsRequest{Project: c.ProjectName}
	return client.List(ctx, req, bindings.ListProjectRoleBindings, res, c.Skip(serviceName))
}
