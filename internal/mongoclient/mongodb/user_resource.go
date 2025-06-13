package mongodb

import (
	"context"
	"fmt"

	"github.com/avast/retry-go/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"terraform-provider-mongodb/internal/mongoclient/types"
)

func (r *ResourceUser) Create(ctx context.Context, plan types.User) error {
	if isDefaultUser(plan.Username) {
		return fmt.Errorf("user %s is a default user and cannot be created", plan.Username)
	}

	err := retry.Do(
		func() error {
			c, err := r.connect(ctx)
			if err != nil {
				return err
			}

			defer func() {
				_ = c.Disconnect(ctx)
			}()

			exist, err := userExists(ctx, c, plan.Username)
			if err != nil {
				return fmt.Errorf("failed to check if user exists: %s", err)
			}

			if exist {
				return retry.Unrecoverable(fmt.Errorf("user %s already exists", plan.Username))
			}

			roles := make([]bson.M, 0, len(plan.Roles))
			for _, i := range plan.Roles {
				roles = append(roles, bson.M{
					"role": i.Role,
					"db":   i.Database,
				})
			}

			command := bson.D{
				{"createUser", plan.Username},
				{"pwd", plan.Password},
				{"roles", roles},
			}

			return c.Database(types.DefaultDatabase).RunCommand(ctx, command).Err()
		},
		retry.Attempts(r.RetryAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(r.RetryDelay),
		retry.Context(ctx),
	)

	return err
}

func (r *ResourceUser) Exists(ctx context.Context, state types.User) (bool, error) {
	var exist bool

	err := retry.Do(
		func() error {
			var err error

			c, err := r.connect(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect to MongoDB: %s", err)
			}

			defer func() {
				_ = c.Disconnect(ctx)
			}()

			exist, err = userExists(ctx, c, state.Username)

			return err
		},
		retry.Attempts(r.RetryAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(r.RetryDelay),
		retry.Context(ctx),
	)

	return exist, err
}

func (r *ResourceUser) Delete(ctx context.Context, state types.User) error {
	if isDefaultUser(state.Username) {
		return fmt.Errorf("user %s is a default user and cannot be deleted", state.Username)
	}

	err := retry.Do(
		func() error {
			c, err := r.connect(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect to MongoDB: %s", err)
			}

			defer func() {
				_ = c.Disconnect(ctx)
			}()

			exist, err := userExists(ctx, c, state.Username)
			if err != nil {
				return fmt.Errorf("failed to check if user exists: %s", err)
			}

			if !exist {
				return retry.Unrecoverable(fmt.Errorf("user %s does not exist", state.Username))
			}

			return c.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{
				{"dropUser", state.Username},
			}).Err()
		},
		retry.Attempts(r.RetryAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(r.RetryDelay),
		retry.Context(ctx),
	)

	return err
}

func (r *ResourceUser) Update(ctx context.Context, plan types.User) error {
	if isDefaultUser(plan.Username) {
		return fmt.Errorf("user %s is a default user and cannot be updated", plan.Username)
	}

	err := retry.Do(
		func() error {
			c, err := r.connect(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect to MongoDB: %s", err)
			}

			defer func() {
				_ = c.Disconnect(ctx)
			}()

			exist, err := userExists(ctx, c, plan.Username)
			if err != nil {
				return fmt.Errorf("failed to check if user exists: %s", err)
			}

			if !exist {
				return retry.Unrecoverable(fmt.Errorf("user %s does not exist", plan.Username))
			}

			roles := make([]bson.M, 0, len(plan.Roles))
			for _, i := range plan.Roles {
				roles = append(roles, bson.M{
					"role": i.Role,
					"db":   i.Database,
				})
			}

			command := bson.D{
				{"updateUser", plan.Username},
				{"pwd", plan.Password},
				{"roles", roles},
			}

			return c.Database(types.DefaultDatabase).RunCommand(ctx, command).Err()
		},
		retry.Attempts(r.RetryAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(r.RetryDelay),
		retry.Context(ctx),
	)

	return err
}

func (r *ResourceUser) ImportState(ctx context.Context, username string) (types.User, error) {
	var u types.User

	if isDefaultUser(username) {
		return types.User{}, fmt.Errorf("user %s is a default user and cannot be imported", username)
	}

	err := retry.Do(
		func() error {
			c, err := r.connect(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect to MongoDB: %s", err)
			}

			defer func() {
				_ = c.Disconnect(ctx)
			}()

			users, err := listUsers(ctx, c)
			if err != nil {
				return fmt.Errorf("failed to check if user exists: %s", err)
			}

			if !users.Exist(username) {
				return retry.Unrecoverable(fmt.Errorf("user %s does not exist", username))
			}

			user := users.Get(username)

			roles := make([]types.Role, 0, len(user.Roles))
			for _, i := range user.Roles {
				roles = append(roles, types.Role{
					Role:     i.Role,
					Database: i.Database,
				})
			}

			u = types.User{
				Username: user.Username,
				Roles:    roles,
			}

			return nil
		},
		retry.Attempts(r.RetryAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(r.RetryDelay),
		retry.Context(ctx),
	)

	if err != nil {
		return types.User{}, err
	}

	return u, nil
}

func (r *ResourceUser) connect(ctx context.Context) (*mongo.Client, error) {
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
