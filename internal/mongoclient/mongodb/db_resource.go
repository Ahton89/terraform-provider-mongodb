package mongodb

import (
	"context"
	"fmt"

	"github.com/avast/retry-go/v4"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"terraform-provider-mongodb/internal/mongoclient/types"
)

func (r *ResourceDatabase) Create(ctx context.Context, plan types.Database) error {
	if isDefaultDatabase(plan.Name) {
		return fmt.Errorf("database %s is a default database and cannot be created", plan.Name)
	}

	err := retry.Do(
		func() error {
			c, err := r.connect(ctx)
			if err != nil {
				return fmt.Errorf("connection to MongoDB failed with error: %s", err)
			}

			defer func() {
				_ = c.Disconnect(ctx)
			}()

			exist, err := databaseExists(ctx, c, plan.Name)
			if err != nil {
				return fmt.Errorf("failed to check if database exists: %s", err)
			}

			if exist {
				return retry.Unrecoverable(fmt.Errorf("database %s already exists", plan.Name))
			}

			return createDatabase(ctx, c, plan.Name)
		},
		retry.Attempts(r.RetryAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(r.RetryDelay),
		retry.Context(ctx),
	)

	return err
}

func (r *ResourceDatabase) Exists(ctx context.Context, state types.Database) (bool, error) {
	var exist bool

	err := retry.Do(
		func() error {
			c, err := r.connect(ctx)
			if err != nil {
				return fmt.Errorf("connection to MongoDB failed with error: %s", err)
			}

			defer func() {
				_ = c.Disconnect(ctx)
			}()

			exist, err = databaseExists(ctx, c, state.Name)

			return err
		},
		retry.Attempts(r.RetryAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(r.RetryDelay),
		retry.Context(ctx),
	)

	return exist, err
}

func (r *ResourceDatabase) Delete(ctx context.Context, state types.Database) error {
	err := retry.Do(
		func() error {
			c, err := r.connect(ctx)
			if err != nil {
				return fmt.Errorf("connection to MongoDB failed with error: %s", err)
			}

			defer func() {
				_ = c.Disconnect(ctx)
			}()

			exist, err := databaseExists(ctx, c, state.Name)
			if err != nil {
				return fmt.Errorf("database exist check failed with error: %s", err)
			}

			if !exist {
				return retry.Unrecoverable(fmt.Errorf("database %s does not exist", state.Name))
			}

			return deleteDatabase(ctx, c, state.Name)
		},
		retry.Attempts(r.RetryAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(r.RetryDelay),
		retry.Context(ctx),
	)

	return err
}

func (r *ResourceDatabase) ImportState(ctx context.Context, name string) (types.Database, error) {
	if isDefaultDatabase(name) {
		return types.Database{}, fmt.Errorf("database %s is a default database and cannot be imported", name)
	}

	err := retry.Do(
		func() error {
			c, err := r.connect(ctx)
			if err != nil {
				return fmt.Errorf("connection to MongoDB failed with error: %s", err)
			}

			defer func() {
				_ = c.Disconnect(ctx)
			}()

			exist, err := databaseExists(ctx, c, name)
			if err != nil {
				return fmt.Errorf("database exist check failed with error: %s", err)
			}

			if !exist {
				return retry.Unrecoverable(fmt.Errorf("database %s does not exist", name))
			}

			return nil
		},
		retry.Attempts(r.RetryAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(r.RetryDelay),
		retry.Context(ctx),
	)

	return types.Database{Name: name}, err
}

func (r *ResourceDatabase) connect(ctx context.Context) (*mongo.Client, error) {
	opts := options.Client().ApplyURI(r.Uri)

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
