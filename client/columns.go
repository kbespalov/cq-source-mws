package client

import (
	"fmt"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/cloudquery/plugin-sdk/v4/schema"
	mwserrors "go.mws.cloud/go-sdk/mws/errors"
)

// Resources carry no reference to the organization or folder they live under,
// so the multiplexed client supplies them. They are what makes a row joinable
// with mws_resmanager_projects and groupable by folder.
var (
	OrganizationColumn = schema.Column{
		Name:     "organization_id",
		Type:     arrow.BinaryTypes.String,
		Resolver: ResolveOrganization,
	}
	FolderColumn = schema.Column{
		Name:     "folder_id",
		Type:     arrow.BinaryTypes.String,
		Resolver: ResolveFolder,
	}
	ProjectColumn = schema.Column{
		Name:     "project_id",
		Type:     arrow.BinaryTypes.String,
		Resolver: ResolveProject,
	}
)

// ProjectHierarchyColumns is the set every project-scoped table carries.
func ProjectHierarchyColumns() []schema.Column {
	return []schema.Column{OrganizationColumn, FolderColumn, ProjectColumn}
}

// ProjectKeyColumns is ProjectHierarchyColumns for a table whose rows do not
// identify themselves. Nearly every resource has an id that already spells out
// its project, but a Kubernetes release channel is only a name, and the same
// "stable" comes back from every project, so the project joins its key.
func ProjectKeyColumns() []schema.Column {
	columns := ProjectHierarchyColumns()
	for i := range columns {
		if columns[i].Name == ProjectColumn.Name {
			columns[i].PrimaryKey = true
			columns[i].NotNull = true
		}
	}
	return columns
}

// ParentColumn joins a row back to the resource it was listed from, named for
// what that resource is: network_id under a network, cluster_id under a
// database cluster. Most of the platform nests this way, and there is no way
// to ask a project for its subnets or its cluster users directly.
func ParentColumn(name string) schema.Column {
	return schema.Column{
		Name:     name,
		Type:     arrow.BinaryTypes.String,
		Resolver: schema.ParentColumnResolver(idColumn),
	}
}

// ChildColumns is the set for a table nested under another resource. A nested
// table has no multiplexer of its own, it runs on the client of its parent, so
// the hierarchy columns resolve exactly as they do one level up.
func ChildColumns(parent string) []schema.Column {
	return append(ProjectHierarchyColumns(), ParentColumn(parent))
}

// ParentResourceName returns the name of the resource a nested table hangs
// off, which is what the nested list call takes as its path parameter: subnets
// are listed as .../networks/{network}/subnets.
//
// It reads the parent's id column rather than its model, so that one helper
// serves every nesting. The id is the primary key of every resource table, so
// it is resolved and non-null by the time a child runs.
func ParentResourceName(parent *schema.Resource) (string, error) {
	return ParentColumnValue(parent, idColumn)
}

// ParentColumnValue reads any column off the parent row. It exists for the one
// listing whose parent has no id: an mk8s release channel is named and nothing
// else, so its versions have to hang off that name.
func ParentColumnValue(parent *schema.Resource, column string) (string, error) {
	value := parent.Get(column)
	if value == nil || !value.IsValid() {
		return "", fmt.Errorf("parent row of %q carries no %s", parent.Table.Name, column)
	}
	return lastSegment(value.String()), nil
}

// Skip is the predicate List uses to tell a project that never enabled a
// service from one where something actually went wrong. Tables that list
// something global pass nil instead: there, a refusal is a real error.
func (c *Client) Skip(service string) func(error) bool {
	return func(err error) bool { return c.SkipProject(err, service) }
}

// SkipProject reports whether listing a service in the current project failed
// because the service is not there for us, rather than because something broke.
//
// An installation has hundreds of projects and most of them use a handful of
// services, so a project that never enabled compute answers 403 and a missing
// scope answers 404. Treating those as fatal would mean no sync ever completes.
// The cost is that a genuinely missing role looks the same as an unused
// service, which is why every skip is logged.
func (c *Client) SkipProject(err error, service string) bool {
	if !mwserrors.IsAPIErrorPermissionDeniedStatus(err) && !mwserrors.IsAPIErrorNotFoundStatus(err) {
		return false
	}
	c.Logger.Info().
		Str("service", service).
		Str("project", c.ProjectID).
		Err(err).
		Msg("skipping project: service is not enabled or not permitted")
	return true
}
