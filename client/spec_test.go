package client

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/cloudquery/plugin-sdk/v4/scheduler"
)

func TestSpecDefaultsAndValidate(t *testing.T) {
	spec := NewDefaultSpec()
	if err := json.Unmarshal([]byte(`{"concurrency":0}`), spec); err != nil {
		t.Fatal(err)
	}
	spec.SetDefaults()
	if spec.Concurrency < 1 {
		t.Fatalf("concurrency after defaults: %d", spec.Concurrency)
	}
	if spec.Timeout != 30*time.Second {
		t.Fatalf("timeout after defaults: %s", spec.Timeout)
	}
	if err := spec.Validate(); err != nil {
		t.Fatal(err)
	}

	spec.Scheduler = scheduler.Strategy(99)
	if err := spec.Validate(); err == nil {
		t.Fatal("unknown scheduler should fail validation")
	}
}
