package client

import (
	"testing"

	"github.com/cloudquery/plugin-sdk/v4/schema"
)

func TestProjectMultiplexScopesEachClient(t *testing.T) {
	c := &Client{
		hierarchy: &ResourceHierarchy{projects: []Project{
			{ID: "rm/projects/a", Name: "a", Folder: "org/organizations/o/folders/f", Organization: "org/organizations/o"},
			{ID: "rm/projects/b", Name: "b", Folder: "org/organizations/o/folders/g", Organization: "org/organizations/o"},
		}},
	}

	clients := ProjectMultiplex()(c)
	if len(clients) != 2 {
		t.Fatalf("got %d clients, want 2", len(clients))
	}

	first := clients[0].(*Client)
	if first.ProjectID != "rm/projects/a" || first.ProjectName != "a" {
		t.Errorf("first client: id=%q name=%q", first.ProjectID, first.ProjectName)
	}
	if first.FolderID != "org/organizations/o/folders/f" || first.OrganizationID != "org/organizations/o" {
		t.Errorf("first client hierarchy: org=%q folder=%q", first.OrganizationID, first.FolderID)
	}
	if first.ID() != "org:org/organizations/o|folder:org/organizations/o/folders/f|project:rm/projects/a" {
		t.Errorf("client id = %q", first.ID())
	}
}

func TestGlobalMultiplexKeepsTheRootClient(t *testing.T) {
	c := &Client{}
	clients := GlobalMultiplex()(c)
	if len(clients) != 1 || clients[0] != schema.ClientMeta(c) {
		t.Fatalf("global multiplex should return the same client, got %#v", clients)
	}
}
