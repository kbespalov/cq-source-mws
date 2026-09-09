package plugin

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/cloudquery/plugin-sdk/v4/message"
	sdkplugin "github.com/cloudquery/plugin-sdk/v4/plugin"
	"github.com/cloudquery/plugin-sdk/v4/scheduler"
)

// TestSyncProjects runs a real sync against the installation the ambient
// credentials point at. It needs MWS_TOKEN or MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH.
func TestSyncProjects(t *testing.T) {
	if os.Getenv("MWS_TOKEN") == "" && os.Getenv("MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH") == "" {
		t.Skip("no MWS credentials in the environment")
	}

	ctx := context.Background()
	p := Plugin()

	strategy := scheduler.StrategyDFS
	spec, err := json.Marshal(map[string]any{
		"concurrency": 10,
		"scheduler":   strategy.String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Init(ctx, spec, sdkplugin.NewClientOptions{}); err != nil {
		t.Fatalf("init: %v", err)
	}

	messages, err := p.SyncAll(ctx, sdkplugin.SyncOptions{
		Tables: []string{"mws_resmanager_projects"},
	})
	if err != nil {
		t.Fatalf("sync: %v", err)
	}

	var rows int
	for _, msg := range messages {
		insert, ok := msg.(*message.SyncInsert)
		if !ok {
			continue
		}
		if rows == 0 {
			logSample(t, insert)
		}
		rows += int(insert.Record.NumRows())
	}

	if rows == 0 {
		t.Fatal("sync produced no rows")
	}
	t.Logf("synced %d projects", rows)
}

// logSample prints the columns that only resolve correctly if the custom
// transformers did their job, so an empty one is visible instead of silent.
func logSample(t *testing.T, insert *message.SyncInsert) {
	rec := insert.Record
	schema := rec.Schema()
	interesting := []string{"id", "display_name", "create_time", "spec_folder", "status_organization", "status_ready_state"}

	for row := 0; row < min(int(rec.NumRows()), 2); row++ {
		for _, name := range interesting {
			indices := schema.FieldIndices(name)
			if len(indices) == 0 {
				t.Errorf("column %q missing from record", name)
				continue
			}
			col := rec.Column(indices[0])
			if col.IsNull(row) {
				t.Logf("row %d %-20s NULL", row, name)
				continue
			}
			t.Logf("row %d %-20s %v", row, name, col.ValueStr(row))
		}
	}
}
