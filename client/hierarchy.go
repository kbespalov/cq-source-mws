package client

import (
	"context"
	"fmt"
	"slices"

	"github.com/rs/zerolog"
	"go.mws.cloud/go-sdk/mws"
	"go.mws.cloud/go-sdk/mws/page"
	rmclient "go.mws.cloud/go-sdk/service/rm/client"
	rmsdk "go.mws.cloud/go-sdk/service/rm/sdk"
)

// Project is one node of the hierarchy, carrying the path of every ancestor so
// that a multiplexed client can be built from it alone.
type Project struct {
	ID string // rm/projects/my-project
	// Name is the trailing segment of the ID. Service APIs take it as the
	// {project} path parameter, e.g. GET /compute/v1/projects/{project}/...
	Name         string
	DisplayName  string
	Folder       string // org/organizations/my-org/folders/my-folder
	Organization string // organizations/my-org
}

// ResourceHierarchy is the organization/folder/project tree, discovered once at
// startup. MWS has no public list API for organizations or folders, but every
// project reports both of its ancestors, so listing projects is enough to
// reconstruct the whole tree.
type ResourceHierarchy struct {
	projects []Project
}

func NewResourceHierarchy(ctx context.Context, logger zerolog.Logger, sdk *mws.SDK, orgIDs, folderIDs, projectIDs []string) (*ResourceHierarchy, error) {
	projects, err := rmsdk.NewProject(ctx, sdk)
	if err != nil {
		return nil, fmt.Errorf("create resource manager client: %w", err)
	}

	onlyAvailable := true
	pager := page.NewPager(
		rmclient.ListProjectsV3Request{OnlyAvailableProjects: &onlyAvailable},
		projects.ListProjectsV3,
	)

	h := &ResourceHierarchy{}
	for item, err := range pager.All(ctx) {
		if err != nil {
			return nil, fmt.Errorf("list projects: %w", err)
		}
		folder := item.GetSpec().Folder
		id := item.GetMetadata().GetId()
		p := Project{
			ID:          id.ID(),
			Name:        string(id.ResourceName()),
			DisplayName: item.GetMetadata().GetDisplayNameOr(""),
			Folder:      RefPath(&folder),
		}
		if org := item.GetStatus().GetOrganization(); org != nil {
			p.Organization = RefPath(org)
		}
		if !keep(p, orgIDs, folderIDs, projectIDs) {
			continue
		}
		h.projects = append(h.projects, p)
	}

	logger.Info().
		Int("organizations", len(h.Organizations())).
		Int("folders", len(h.Folders())).
		Int("projects", len(h.projects)).
		Msg("fetched resource hierarchy")

	return h, nil
}

// keep reports whether a project survives the spec filters. A filter matches
// either the full path or the trailing name, so both "organizations/my-org"
// and "my-org" select the same organization.
func keep(p Project, orgIDs, folderIDs, projectIDs []string) bool {
	return matches(p.Organization, orgIDs) &&
		matches(p.Folder, folderIDs) &&
		matches(p.ID, projectIDs)
}

func matches(path string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	return slices.ContainsFunc(filters, func(f string) bool {
		return f == path || f == lastSegment(path)
	})
}

func lastSegment(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

func (h *ResourceHierarchy) Projects() []Project { return h.projects }

func (h *ResourceHierarchy) Organizations() []string {
	return unique(h.projects, func(p Project) string { return p.Organization })
}

func (h *ResourceHierarchy) Folders() []string {
	return unique(h.projects, func(p Project) string { return p.Folder })
}

// FolderRows returns one row per folder, carrying its organization.
func (h *ResourceHierarchy) FolderRows() []Project {
	return uniqueBy(h.projects, func(p Project) string { return p.Folder })
}

// OrganizationRows returns one row per organization.
func (h *ResourceHierarchy) OrganizationRows() []Project {
	return uniqueBy(h.projects, func(p Project) string { return p.Organization })
}

func unique(projects []Project, key func(Project) string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, p := range projects {
		k := key(p)
		if k == "" {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	return out
}

func uniqueBy(projects []Project, key func(Project) string) []Project {
	seen := map[string]struct{}{}
	var out []Project
	for _, p := range projects {
		k := key(p)
		if k == "" {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, p)
	}
	return out
}
