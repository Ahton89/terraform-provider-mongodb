package v6

import (
	"context"
	"fmt"

	"terraform-provider-mongodb/internal/mongoclient/types"
)

func (r *ResourceDatabase) Create(ctx context.Context, plan types.Database) error {
	if isDefaultDatabase(plan.Name) {
		return fmt.Errorf("database %s is a default database and cannot be created", plan.Name)
	}

	exist, err := databaseExists(ctx, r.Client, plan.Name)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %s", err)
	}

	if exist {
		return fmt.Errorf("database %s already exists", plan.Name)
	}

	return createDatabase(ctx, r.Client, plan.Name)
}

func (r *ResourceDatabase) Exists(ctx context.Context, state types.Database) (bool, error) {
	return databaseExists(ctx, r.Client, state.Name)
}

func (r *ResourceDatabase) Delete(ctx context.Context, state types.Database) error {
	exist, err := databaseExists(ctx, r.Client, state.Name)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %s", err)
	}

	if !exist {
		return fmt.Errorf("database %s does not exist", state.Name)
	}

	return deleteDatabase(ctx, r.Client, state.Name)
}

func (r *ResourceDatabase) ImportState(ctx context.Context, name string) (types.Database, error) {
	if isDefaultDatabase(name) {
		return types.Database{}, fmt.Errorf("database %s is a default database and cannot be imported", name)
	}

	exist, err := databaseExists(ctx, r.Client, name)
	if err != nil {
		return types.Database{}, fmt.Errorf("failed to check if database exists: %s", err)
	}

	if !exist {
		return types.Database{}, fmt.Errorf("database %s does not exist", name)
	}

	return types.Database{Name: name}, nil
}
