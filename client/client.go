// Package client talks to the MWS Cloud Platform and turns its models into
// CloudQuery tables.
//
// Every resource on the platform is shaped the same way — kind, metadata, spec,
// status — but the generated Go types hide identifiers, references, quantities
// and optionals from reflection. TransformResource is the walk that turns those
// wrappers into columns a warehouse can query.
package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudquery/plugin-sdk/v4/state"
	"github.com/rs/zerolog"
	"go.mws.cloud/go-sdk/mws"
)

type Client struct {
	hierarchy *ResourceHierarchy

	OrganizationID string
	FolderID       string
	// ProjectID is the full path, "rm/projects/my-project", and is what tables
	// store so rows join back to mws_resmanager_projects.
	ProjectID string
	// ProjectName is the trailing segment that service APIs take as their
	// {project} path parameter.
	ProjectName string

	Backend state.Client
	Logger  zerolog.Logger
	SDK     *mws.SDK
}

func (c *Client) ID() string {
	parts := make([]string, 0, 3)
	if c.OrganizationID != "" {
		parts = append(parts, "org:"+c.OrganizationID)
	}
	if c.FolderID != "" {
		parts = append(parts, "folder:"+c.FolderID)
	}
	if c.ProjectID != "" {
		parts = append(parts, "project:"+c.ProjectID)
	}
	return strings.Join(parts, "|")
}

func (c *Client) Hierarchy() *ResourceHierarchy { return c.hierarchy }

func (c *Client) WithBackend(backend state.Client) *Client {
	nc := *c
	nc.Backend = backend
	return &nc
}

func (c *Client) WithOrganization(id string) *Client {
	nc := *c
	nc.Logger = c.Logger.With().Str("organization", id).Logger()
	nc.OrganizationID = id
	return &nc
}

func (c *Client) WithFolder(id string) *Client {
	nc := *c
	nc.Logger = c.Logger.With().Str("folder", id).Logger()
	nc.FolderID = id
	return &nc
}

func (c *Client) WithProject(p Project) *Client {
	nc := *c
	nc.Logger = c.Logger.With().Str("project", p.ID).Logger()
	nc.ProjectID = p.ID
	nc.ProjectName = p.Name
	return &nc
}

func New(ctx context.Context, logger zerolog.Logger, spec *Spec) (*Client, error) {
	opts := []mws.LoadSDKOption{
		mws.WithTimeout(spec.Timeout),
		mws.WithUserAgent("cq-source-mws"),
	}
	if spec.Debug {
		opts = append(opts, mws.WithTraceEnabled())
	}

	sdk, err := mws.Load(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("initialize MWS SDK: %w", err)
	}

	c := &Client{
		SDK:    sdk,
		Logger: logger,
	}

	c.hierarchy, err = NewResourceHierarchy(ctx, logger, sdk, spec.OrganizationIDs, spec.FolderIDs, spec.ProjectIDs)
	if err != nil {
		return nil, fmt.Errorf("fetch resource hierarchy: %w", err)
	}

	if len(spec.OrganizationIDs) == 0 && len(spec.FolderIDs) == 0 && len(spec.ProjectIDs) == 0 {
		logger.Warn().Msg("no organization_ids, folder_ids or project_ids specified – syncing every reachable project")
	}

	return c, nil
}

// Close cancels in-flight SDK requests and releases the HTTP client. The
// plugin calls this when a sync finishes; leaving it out leaks connections
// and the credential refresher the SDK starts in the background.
func (c *Client) Close(ctx context.Context) error {
	if c == nil || c.SDK == nil {
		return nil
	}
	return c.SDK.Close(ctx)
}
