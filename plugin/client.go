package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudquery/plugin-sdk/v4/message"
	"github.com/cloudquery/plugin-sdk/v4/plugin"
	"github.com/cloudquery/plugin-sdk/v4/scheduler"
	"github.com/cloudquery/plugin-sdk/v4/schema"
	"github.com/cloudquery/plugin-sdk/v4/state"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"cq-source-mws/client"
)

const maxMsgSize = 100 * 1024 * 1024 // 100 MiB

type Client struct {
	plugin.UnimplementedDestination
	scheduler  *scheduler.Scheduler
	syncClient *client.Client
	options    plugin.NewClientOptions
	allTables  schema.Tables
}

func NewClient(ctx context.Context, logger zerolog.Logger, specBytes []byte, options plugin.NewClientOptions) (plugin.Client, error) {
	c := &Client{
		options:   options,
		allTables: tables(),
	}
	if options.NoConnection {
		return c, nil
	}

	spec := client.NewDefaultSpec()
	if err := json.Unmarshal(specBytes, spec); err != nil {
		return nil, err
	}
	spec.SetDefaults()
	if err := spec.Validate(); err != nil {
		return nil, err
	}

	var err error
	c.syncClient, err = client.New(ctx, logger, spec)
	if err != nil {
		return nil, err
	}

	c.scheduler = scheduler.NewScheduler(
		scheduler.WithLogger(logger),
		scheduler.WithConcurrency(spec.Concurrency),
		scheduler.WithStrategy(spec.Scheduler),
	)
	return c, nil
}

func (c *Client) Close(ctx context.Context) error {
	if c.syncClient == nil {
		return nil
	}
	return c.syncClient.Close(ctx)
}

func (c *Client) Tables(_ context.Context, options plugin.TableOptions) (schema.Tables, error) {
	return c.allTables.FilterDfs(options.Tables, options.SkipTables, options.SkipDependentTables)
}

func (c *Client) Sync(ctx context.Context, options plugin.SyncOptions, res chan<- message.SyncMessage) error {
	if c.options.NoConnection {
		return fmt.Errorf("Client.NoConnection is not supported")
	}
	tables, err := c.allTables.FilterDfs(options.Tables, options.SkipTables, options.SkipDependentTables)
	if err != nil {
		return err
	}

	stateClient := state.Client(&state.NoOpClient{})
	if options.BackendOptions != nil {
		conn, err := grpc.NewClient(options.BackendOptions.Connection,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(maxMsgSize),
				grpc.MaxCallSendMsgSize(maxMsgSize),
			))
		if err != nil {
			return fmt.Errorf("dial grpc source plugin at %s: %w", options.BackendOptions.Connection, err)
		}
		stateClient, err = state.NewClient(ctx, conn, options.BackendOptions.TableName)
		if err != nil {
			return fmt.Errorf("create state client: %w", err)
		}
	}

	if err := c.scheduler.Sync(ctx, c.syncClient.WithBackend(stateClient), tables, res,
		scheduler.WithSyncDeterministicCQID(options.DeterministicCQID)); err != nil {
		return err
	}
	return stateClient.Flush(ctx)
}
