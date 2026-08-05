package mongodb

import (
	"context"
	"fmt"

	"terraform-provider-mongodb/internal/mongoclient/types"

	"github.com/avast/retry-go/v4"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (d *DataSourceReplicaSet) Read(ctx context.Context) (types.ReplicaSet, error) {
	var rsc *types.ReplicaSetConfig

	err := retry.Do(
		func() error {
			c, err := d.connect(ctx)
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

			return nil
		},
		retry.Attempts(d.RetryAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(d.RetryDelay),
		retry.Context(ctx),
	)

	if err != nil {
		return types.ReplicaSet{}, err
	}

	return rsc.Config, nil
}

func (d *DataSourceReplicaSet) connect(ctx context.Context) (*mongo.Client, error) {
	opts := options.Client().ApplyURI(d.Uri)

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
