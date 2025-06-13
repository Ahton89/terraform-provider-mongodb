package mongodb

import (
	"context"
	"fmt"

	"github.com/avast/retry-go/v4"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"terraform-provider-mongodb/internal/mongoclient/types"
)

func (d *DataSourceReplicaSet) Read(ctx context.Context) (types.ReplicaSet, error) {
	var rsc *types.ReplicaSetConfig

	err := retry.Do(
		func() error {
			c, err := d.connect(ctx)
			if err != nil {
				return fmt.Errorf("connection to MongoDB failed with error: %s", err)
			}

			defer func() {
				_ = c.Disconnect(ctx)
			}()

			err = requiredVersion(ctx, c)
			if err != nil {
				return fmt.Errorf("required version check failed with error: %s", err)
			}

			rsc, err = getReplicaSetConfig(ctx, c)
			if err != nil {
				return fmt.Errorf("get replica set config failed with error: %s", err)
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

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to ping MongoDB: %s", err)
	}

	return client, nil
}
