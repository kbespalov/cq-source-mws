package plugin

import (
	"strings"
	"testing"
)

// TestTablesHavePrimaryKey guards the reason TransformResource exists: the
// SDK models hide their identifiers in unexported fields, so a table built
// with the default transformer silently ends up without a usable key.
func TestTablesHavePrimaryKey(t *testing.T) {
	for _, table := range tables().FlattenTables() {
		if len(table.PrimaryKeys()) == 0 {
			t.Errorf("table %q has no primary key", table.Name)
		}
	}
}

// TestTopLevelTablesHaveAMultiplexer is what keeps a catalog table from
// running once per project. A nested table inherits its parent's client and
// is deliberately not checked.
func TestTopLevelTablesHaveAMultiplexer(t *testing.T) {
	for _, table := range tables() {
		if table.Multiplex == nil {
			t.Errorf("table %q has no multiplexer", table.Name)
		}
	}
}

func TestSchemas(t *testing.T) {
	seen := make(map[string]string)
	for _, table := range tables().FlattenTables() {
		t.Run(table.Name, func(t *testing.T) {
			if table.Description == "" {
				t.Error("missing description")
			}
			if !strings.HasPrefix(table.Name, "mws_") {
				t.Errorf("name %q does not start with mws_", table.Name)
			}
			if other, dup := seen[table.Name]; dup {
				t.Errorf("name already used by %s", other)
			}
			seen[table.Name] = table.Name

			columns := make(map[string]struct{}, len(table.Columns))
			for _, col := range table.Columns {
				if _, dup := columns[col.Name]; dup {
					t.Errorf("duplicate column %q", col.Name)
				}
				columns[col.Name] = struct{}{}
			}
		})
	}
}
