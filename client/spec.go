package client

import (
	"fmt"
	"time"

	"github.com/cloudquery/plugin-sdk/v4/scheduler"
)

type Spec struct {
	// Organizations to sync. Empty means every organization the credentials reach.
	OrganizationIDs []string `json:"organization_ids,omitempty"`

	// Folders to sync. Empty means every folder of the selected organizations.
	FolderIDs []string `json:"folder_ids,omitempty"`

	// Projects to sync. Empty means every project of the selected folders.
	//
	// An installation can hold hundreds of projects and most tables list per
	// project, so narrowing this list is the main lever on sync duration.
	ProjectIDs []string `json:"project_ids,omitempty"`

	// If `true`, will log HTTP calls.
	Debug bool `json:"debug,omitempty"`

	// Timeout of a single API call.
	Timeout time.Duration `json:"timeout,omitempty" jsonschema:"default=30s"`

	// The best effort maximum number of Go routines to use.
	Concurrency int `json:"concurrency,omitempty" jsonschema:"minimum=1"`

	// The scheduler to use when determining the priority of resources to sync.
	//
	// Available options: `dfs`, `round-robin`, `shuffle`, `shuffle-queue`
	Scheduler scheduler.Strategy `json:"scheduler,omitempty"`
}

func NewDefaultSpec() *Spec {
	return &Spec{
		Timeout:     30 * time.Second,
		Concurrency: scheduler.DefaultConcurrency,
		Scheduler:   scheduler.StrategyShuffle,
	}
}

func (s *Spec) SetDefaults() {
	if s.Timeout <= 0 {
		s.Timeout = 30 * time.Second
	}
	if s.Concurrency < 1 {
		s.Concurrency = scheduler.DefaultConcurrency
	}
}

func (s *Spec) Validate() error {
	if err := s.Scheduler.Validate(); err != nil {
		return fmt.Errorf("scheduler: %w", err)
	}
	return nil
}
