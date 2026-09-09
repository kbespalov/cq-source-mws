package client

import (
	"github.com/cloudquery/plugin-sdk/v4/schema"
)

// GlobalMultiplex keeps the single unscoped client for tables whose API call
// takes no organization, folder or project.
func GlobalMultiplex() schema.Multiplexer {
	return func(meta schema.ClientMeta) []schema.ClientMeta {
		return []schema.ClientMeta{meta.(*Client)}
	}
}

func OrganizationMultiplex() schema.Multiplexer {
	return func(meta schema.ClientMeta) []schema.ClientMeta {
		c := meta.(*Client)
		rows := c.hierarchy.OrganizationRows()
		l := make([]schema.ClientMeta, len(rows))
		for i, p := range rows {
			l[i] = c.WithOrganization(p.Organization)
		}
		return l
	}
}

func FolderMultiplex() schema.Multiplexer {
	return func(meta schema.ClientMeta) []schema.ClientMeta {
		c := meta.(*Client)
		rows := c.hierarchy.FolderRows()
		l := make([]schema.ClientMeta, len(rows))
		for i, p := range rows {
			l[i] = c.WithOrganization(p.Organization).WithFolder(p.Folder)
		}
		return l
	}
}

func ProjectMultiplex() schema.Multiplexer {
	return func(meta schema.ClientMeta) []schema.ClientMeta {
		c := meta.(*Client)
		rows := c.hierarchy.Projects()
		l := make([]schema.ClientMeta, len(rows))
		for i, p := range rows {
			l[i] = c.WithOrganization(p.Organization).WithFolder(p.Folder).WithProject(p)
		}
		return l
	}
}
