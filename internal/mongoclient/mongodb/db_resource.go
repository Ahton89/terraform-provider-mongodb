package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"terraform-provider-mongodb/internal/mongoclient/types"
)

func (r *ResourceDatabase) Create(ctx context.Context, plan types.Database) error {
	var c *mongo.Client
	var exist bool
	var err error

	c, err = r.connect()
	if err != nil {
		return fmt.Errorf("connection to MongoDB failed with error: %s", err)
	}

	defer func() {
		_ = c.Disconnect(ctx)
	}()

	if isDefaultDatabase(plan.Name) {
		return fmt.Errorf("database %s is a default database and cannot be created", plan.Name)
	}

	exist, err = databaseExists(ctx, c, plan.Name)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %s", err)
	}

	if exist {
		return fmt.Errorf("database %s already exists", plan.Name)
	}

	return createDatabase(ctx, c, plan.Name)
}

func (r *ResourceDatabase) Exists(ctx context.Context, state types.Database) (bool, error) {
	c, err := r.connect()
	if err != nil {
		return false, fmt.Errorf("connection to MongoDB failed with error: %s", err)
	}

	defer func() {
		_ = c.Disconnect(ctx)
	}()

	return databaseExists(ctx, c, state.Name)
}

func (r *ResourceDatabase) Delete(ctx context.Context, state types.Database) error {
	var c *mongo.Client
	var exist bool
	var err error

	c, err = r.connect()
	if err != nil {
		return fmt.Errorf("connection to MongoDB failed with error: %s", err)
	}

	defer func() {
		_ = c.Disconnect(ctx)
	}()

	exist, err = databaseExists(ctx, c, state.Name)
	if err != nil {
		return fmt.Errorf("database exist check failed with error: %s", err)
	}

	if !exist {
		return fmt.Errorf("database %s does not exist", state.Name)
	}

	return deleteDatabase(ctx, c, state.Name)
}

func (r *ResourceDatabase) ImportState(ctx context.Context, name string) (types.Database, error) {
	var c *mongo.Client
	var exist bool
	var err error

	c, err = r.connect()
	if err != nil {
		return types.Database{}, fmt.Errorf("connection to MongoDB failed with error: %s", err)
	}

	defer func() {
		_ = c.Disconnect(ctx)
	}()

	if isDefaultDatabase(name) {
		return types.Database{}, fmt.Errorf("database %s is a default database and cannot be imported", name)
	}

	exist, err = databaseExists(ctx, c, name)
	if err != nil {
		return types.Database{}, fmt.Errorf("database exist check failed with error: %s", err)
	}

	if !exist {
		return types.Database{}, fmt.Errorf("database %s does not exist", name)
	}

	return types.Database{Name: name}, nil
}

func (r *ResourceDatabase) connect() (*mongo.Client, error) {
	opts := options.Client().ApplyURI(r.Uri)
	return mongo.Connect(opts)
}
