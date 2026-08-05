package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"terraform-provider-mongodb/internal/mongoclient/types"

	"github.com/avast/retry-go/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/topology"
)

func (r *ResourceReplicaSet) Create(ctx context.Context, plan types.ReplicaSet) error {
	err := retry.Do(
		func() error {
			c, err := r.directConnect(ctx)
			if err != nil {
				return fmt.Errorf("connection to MongoDB failed with error: %w", err)
			}

			defer func() {
				disconnectCtx, cancel := context.WithTimeout(ctx, defaultContextTimeout)
				_ = c.Disconnect(disconnectCtx)
				cancel()
			}()

			err = requiredVersion(ctx, c)
			if err != nil {
				return fmt.Errorf("required version check failed with error: %w", err)
			}

			err = c.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{
				{"replSetInitiate", plan},
			}).Err()
			if err != nil {
				var cmdErr mongo.CommandError
				// Return if already exists
				if errors.As(err, &cmdErr) && cmdErr.Code == 23 {
					rsc, cfgErr := getReplicaSetConfig(ctx, c)
					if cfgErr != nil {
						return fmt.Errorf("replica set already initialized, failed to read its config: %w", cfgErr)
					}
					if rsc.Config.Name != plan.Name {
						return retry.Unrecoverable(fmt.Errorf(
							"node is already a member of replica set %q, cannot initiate %q",
							rsc.Config.Name, plan.Name))
					}
					return r.waitForReplicaSetReady(ctx, plan.Name)
				}
				if errors.As(err, &cmdErr) && cmdErr.Code == 76 {
					return retry.Unrecoverable(fmt.Errorf("no replication enabled for replica set %s", plan.Name))
				}
				return fmt.Errorf("create replica set failed with error: %w", err)
			}

			return r.waitForReplicaSetReady(ctx, plan.Name)
		},
		retry.Attempts(r.RetryAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(r.RetryDelay),
		retry.Context(ctx),
	)

	return err
}

func (r *ResourceReplicaSet) Exists(ctx context.Context, state types.ReplicaSet) (types.ReplicaSet, bool, error) {
	var rsc *types.ReplicaSetConfig
	var initialized bool

	err := retry.Do(
		func() error {
			c, _, err := r.connect(ctx)
			if err != nil {
				// Try to connect directly
				// To cover case when mongo cluster was recreated, but state still exists
				var selectionErr topology.ServerSelectionError
				if errors.As(err, &selectionErr) {
					c, err = r.directConnect(ctx)
					if err != nil {
						return fmt.Errorf("direct connection to MongoDB failed with error: %w", err)
					}
				} else {
					return fmt.Errorf("connection to MongoDB failed with error: %w", err)
				}
			}

			defer func() {
				disconnectCtx, cancel := context.WithTimeout(ctx, defaultContextTimeout)
				_ = c.Disconnect(disconnectCtx)
				cancel()
			}()

			err = requiredVersion(ctx, c)
			if err != nil {
				return fmt.Errorf("required version check failed with error: %w", err)
			}

			rsc, err = getReplicaSetConfig(ctx, c)
			if err != nil {
				var commandErr mongo.CommandError
				if errors.As(err, &commandErr) && commandErr.Code == 94 {
					// NotYetInitialized: there is no replica set on the server yet.
					initialized = false
					return nil
				}
				if errors.As(err, &commandErr) && commandErr.Code == 76 {
					return retry.Unrecoverable(fmt.Errorf("no replication enabled for replica set %s", state.Name))
				}
				return fmt.Errorf("get replica set config failed with error: %w", err)
			}

			initialized = true
			return nil
		},
		retry.Attempts(r.RetryAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(r.RetryDelay),
		retry.Context(ctx),
	)

	if err != nil {
		return types.ReplicaSet{}, false, fmt.Errorf("failed to check if replica set exists: %w", err)
	}

	if !initialized || rsc.Config.Name != state.Name {
		return types.ReplicaSet{}, false, nil
	}

	return rsc.Config, true, nil
}

func (r *ResourceReplicaSet) Update(ctx context.Context, state types.ReplicaSet) error {
	err := retry.Do(
		func() error {
			c, _, err := r.connect(ctx)
			if err != nil {
				return fmt.Errorf("connection to MongoDB failed with error: %w", err)
			}

			defer func() {
				disconnectCtx, cancel := context.WithTimeout(ctx, defaultContextTimeout)
				_ = c.Disconnect(disconnectCtx)
				cancel()
			}()

			err = requiredVersion(ctx, c)
			if err != nil {
				return fmt.Errorf("required version check failed with error: %w", err)
			}

			status, err := getReplicaSetStatus(ctx, c)
			if err != nil {
				return fmt.Errorf("get replica set status failed with error: %w", err)
			}

			if !isReplicaSetReady(status, state.Name) {
				return fmt.Errorf("replica set %s not ready or corrupted", state.Name)
			}

			// Get current config version and increment it
			version, err := getReplicaSetConfigVersion(ctx, c)
			if err != nil {
				return fmt.Errorf("get replica set config version failed with error: %w", err)
			}

			version++

			// Set the new version to the config
			state.SetVersion(&version)

			err = c.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{
				{"replSetReconfig", state},
			}).Err()
			if err != nil {
				return fmt.Errorf("updating replica set failed with error: %w", err)
			}

			// Clear version in state
			state.ClearVersion()

			return r.waitForReplicaSetReady(ctx, state.Name)
		},
		retry.Attempts(r.RetryAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(r.RetryDelay),
		retry.Context(ctx),
	)

	return err
}

func (r *ResourceReplicaSet) ImportState(ctx context.Context, name string) (types.ReplicaSet, error) {
	var rsc *types.ReplicaSetConfig

	err := retry.Do(
		func() error {
			c, _, err := r.connect(ctx)
			if err != nil {
				return fmt.Errorf("connection to MongoDB failed with error: %w", err)
			}

			defer func() {
				disconnectCtx, cancel := context.WithTimeout(ctx, defaultContextTimeout)
				_ = c.Disconnect(disconnectCtx)
				cancel()
			}()

			err = requiredVersion(ctx, c)
			if err != nil {
				return fmt.Errorf("required version check failed with error: %w", err)
			}

			rsc, err = getReplicaSetConfig(ctx, c)
			if err != nil {
				return fmt.Errorf("get replica set config failed with error: %w", err)
			}

			if rsc.Config.Name != name {
				return retry.Unrecoverable(fmt.Errorf("replica set %s does not exist", name))
			}

			rsc.Config.ClearTimeouts()

			return nil
		},
		retry.Attempts(r.RetryAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(r.RetryDelay),
		retry.Context(ctx),
	)

	if err != nil {
		return types.ReplicaSet{}, err
	}

	return rsc.Config, nil
}

func (r *ResourceReplicaSet) connect(ctx context.Context) (*mongo.Client, bool, error) {
	opts := options.Client().ApplyURI(r.Uri)

	if opts.ReplicaSet == nil {
		return nil, false, fmt.Errorf("you can't use direct connection when working with replica set")
	}

	if err := opts.Validate(); err != nil {
		return nil, false, fmt.Errorf("invalid connection options: %w", err)
	}

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, true, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		disconnectCtx, cancel := context.WithTimeout(ctx, defaultContextTimeout)
		_ = client.Disconnect(disconnectCtx)
		cancel()

		return nil, true, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client, false, nil
}

func (r *ResourceReplicaSet) directConnect(ctx context.Context) (*mongo.Client, error) {
	opts := options.Client().ApplyURI(r.Uri)
	opts.ReplicaSet = nil
	if len(opts.Hosts) == 0 {
		return nil, fmt.Errorf("no hosts provided for direct connection")
	}
	opts.Hosts = []string{opts.Hosts[0]}
	opts.SetDirect(true)

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid connection options: %w", err)
	}

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		disconnectCtx, cancel := context.WithTimeout(ctx, defaultContextTimeout)
		_ = client.Disconnect(disconnectCtx)
		cancel()

		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client, nil
}

func (r *ResourceReplicaSet) waitForReplicaSetReady(ctx context.Context, replicaSetName string) error {
	ticker := time.NewTicker(replicaSetPollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case _, ok := <-ticker.C:
			if !ok {
				return fmt.Errorf("ticker stopped")
			}

			var client *mongo.Client
			var retryable bool
			var status *types.ReplicaSetStatus
			var err error

			client, retryable, err = r.connect(ctx)
			if err != nil {
				if retryable {
					continue
				}
				return fmt.Errorf("connection to MongoDB failed with error: %w", err)
			}

			status, err = getReplicaSetStatus(ctx, client)

			disconnectCtx, cancel := context.WithTimeout(ctx, defaultContextTimeout)
			_ = client.Disconnect(disconnectCtx)
			cancel()

			if err == nil && isReplicaSetReady(status, replicaSetName) {
				return nil
			}
		}
	}
}
