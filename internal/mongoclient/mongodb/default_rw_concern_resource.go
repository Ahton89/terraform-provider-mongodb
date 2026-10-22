package mongodb

import (
	"context"
	"fmt"

	"terraform-provider-mongodb/internal/mongoclient/types"

	"github.com/avast/retry-go/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *ResourceDefaultRWConcern) Create(ctx context.Context, plan types.DefaultRWConcern) error {
	return r.run(ctx, func(c *mongo.Client) error {
		return setDefaultRWConcern(ctx, c,
			readConcernDocument(plan.DefaultReadConcern),
			writeConcernDocument(plan.DefaultWriteConcern),
		)
	})
}

func (r *ResourceDefaultRWConcern) Read(ctx context.Context) (types.DefaultRWConcern, error) {
	var d *types.RWConcernDefaults

	err := r.run(ctx, func(c *mongo.Client) error {
		var err error

		d, err = getDefaultRWConcern(ctx, c)
		if err != nil {
			return fmt.Errorf("get default read/write concern failed with error: %w", err)
		}

		return nil
	})

	if err != nil {
		return types.DefaultRWConcern{}, err
	}

	return d.Global(), nil
}

func (r *ResourceDefaultRWConcern) Update(ctx context.Context, plan, state types.DefaultRWConcern) error {
	if plan.DefaultWriteConcern == nil && state.DefaultWriteConcern != nil {
		return fmt.Errorf("the default write concern cannot be unset once it is set")
	}

	readConcern := readConcernDocument(plan.DefaultReadConcern)
	if plan.DefaultReadConcern == nil && state.DefaultReadConcern != nil {
		readConcern = bson.D{}
	}

	writeConcern := writeConcernDocument(plan.DefaultWriteConcern)

	if readConcern == nil && writeConcern == nil {
		return nil
	}

	return r.run(ctx, func(c *mongo.Client) error {
		return setDefaultRWConcern(ctx, c, readConcern, writeConcern)
	})
}

func (r *ResourceDefaultRWConcern) ImportState(ctx context.Context) (types.DefaultRWConcern, error) {
	d, err := r.Read(ctx)
	if err != nil {
		return types.DefaultRWConcern{}, err
	}

	if d.DefaultReadConcern == nil && d.DefaultWriteConcern == nil {
		return types.DefaultRWConcern{}, fmt.Errorf("neither default read concern nor default write concern is set, nothing to import")
	}

	d.ClearTimeouts()

	return d, nil
}

func (r *ResourceDefaultRWConcern) run(ctx context.Context, fn func(c *mongo.Client) error) error {
	return retry.Do(
		func() error {
			c, err := r.connect(ctx)
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

			return fn(c)
		},
		retry.Attempts(r.RetryAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(r.RetryDelay),
		retry.Context(ctx),
	)
}

func (r *ResourceDefaultRWConcern) connect(ctx context.Context) (*mongo.Client, error) {
	opts := options.Client().ApplyURI(r.Uri)

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
