package client

import "testing"

func TestLastSegment(t *testing.T) {
	for _, tc := range []struct {
		path, want string
	}{
		{"rm/projects/my-project", "my-project"},
		{"my-project", "my-project"},
		{"org/organizations/my-org/folders/my-folder", "my-folder"},
		{"", ""},
	} {
		if got := lastSegment(tc.path); got != tc.want {
			t.Errorf("lastSegment(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

func TestKeepMatchesFullPathOrName(t *testing.T) {
	p := Project{
		ID:           "rm/projects/my-project",
		Name:         "my-project",
		Folder:       "org/organizations/my-org/folders/my-folder",
		Organization: "org/organizations/my-org",
	}

	if !keep(p, nil, nil, nil) {
		t.Fatal("empty filters must keep every project")
	}
	if !keep(p, []string{"my-org"}, nil, nil) {
		t.Fatal("organization name should match")
	}
	if !keep(p, []string{"org/organizations/my-org"}, nil, nil) {
		t.Fatal("organization path should match")
	}
	if keep(p, []string{"other"}, nil, nil) {
		t.Fatal("a different organization must drop the project")
	}
	if !keep(p, nil, nil, []string{"my-project"}) {
		t.Fatal("project name should match")
	}
	if keep(p, nil, nil, []string{"other"}) {
		t.Fatal("a different project must be dropped")
	}
}

func TestUniqueBySkipsEmptyAndDuplicates(t *testing.T) {
	projects := []Project{
		{ID: "rm/projects/a", Folder: "org/organizations/o/folders/f", Organization: "org/organizations/o"},
		{ID: "rm/projects/b", Folder: "org/organizations/o/folders/f", Organization: "org/organizations/o"},
		{ID: "rm/projects/c", Folder: "", Organization: ""},
	}

	h := &ResourceHierarchy{projects: projects}
	if got := h.Organizations(); len(got) != 1 || got[0] != "org/organizations/o" {
		t.Errorf("Organizations() = %v", got)
	}
	if got := h.Folders(); len(got) != 1 || got[0] != "org/organizations/o/folders/f" {
		t.Errorf("Folders() = %v", got)
	}
	if got := h.OrganizationRows(); len(got) != 1 || got[0].ID != "rm/projects/a" {
		t.Errorf("OrganizationRows() kept the wrong first row: %+v", got)
	}
}
