package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"terraform-provider-mongodb/internal/mongoclient/types"
)

func (r *ResourceUser) Create(ctx context.Context, plan types.User) error {
	var c *mongo.Client
	var exist bool
	var err error

	if isDefaultUser(plan.Username) {
		return fmt.Errorf("user %s is a default user and cannot be created", plan.Username)
	}

	c, err = r.connect()
	if err != nil {
		return err
	}

	defer func() {
		_ = c.Disconnect(ctx)
	}()

	exist, err = userExists(ctx, c, plan.Username)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %s", err)
	}

	if exist {
		return fmt.Errorf("user %s already exists", plan.Username)
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

	err = c.Database(types.DefaultDatabase).RunCommand(ctx, command).Err()
	if err != nil {
		return fmt.Errorf("failed to create user: %s", err)
	}

	return nil
}

func (r *ResourceUser) Exists(ctx context.Context, state types.User) (bool, error) {
	c, err := r.connect()
	if err != nil {
		return false, fmt.Errorf("failed to connect to MongoDB: %s", err)
	}

	defer func() {
		_ = c.Disconnect(ctx)
	}()

	return userExists(ctx, c, state.Username)
}

func (r *ResourceUser) Delete(ctx context.Context, state types.User) error {
	var c *mongo.Client
	var exist bool
	var err error

	if isDefaultUser(state.Username) {
		return fmt.Errorf("user %s is a default user and cannot be deleted", state.Username)
	}

	c, err = r.connect()
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %s", err)
	}

	defer func() {
		_ = c.Disconnect(ctx)
	}()

	exist, err = userExists(ctx, c, state.Username)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %s", err)
	}

	if !exist {
		return fmt.Errorf("user %s does not exist", state.Username)
	}

	err = c.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{
		{"dropUser", state.Username},
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to delete user: %s", err)
	}

	return nil
}

func (r *ResourceUser) Update(ctx context.Context, plan types.User) error {
	var c *mongo.Client
	var exist bool
	var err error

	if isDefaultUser(plan.Username) {
		return fmt.Errorf("user %s is a default user and cannot be updated", plan.Username)
	}

	c, err = r.connect()
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %s", err)
	}

	defer func() {
		_ = c.Disconnect(ctx)
	}()

	exist, err = userExists(ctx, c, plan.Username)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %s", err)
	}

	if !exist {
		return fmt.Errorf("user %s does not exist", plan.Username)
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

	err = c.Database(types.DefaultDatabase).RunCommand(ctx, command).Err()
	if err != nil {
		return fmt.Errorf("failed to update user: %s", err)
	}

	return nil
}

func (r *ResourceUser) ImportState(ctx context.Context, username string) (types.User, error) {
	var c *mongo.Client
	var users types.Users
	var err error

	if isDefaultUser(username) {
		return types.User{}, fmt.Errorf("user %s is a default user and cannot be imported", username)
	}

	c, err = r.connect()
	if err != nil {
		return types.User{}, fmt.Errorf("failed to connect to MongoDB: %s", err)
	}

	defer func() {
		_ = c.Disconnect(ctx)
	}()

	users, err = listUsers(ctx, c)
	if err != nil {
		return types.User{}, fmt.Errorf("failed to check if user exists: %s", err)
	}

	if !users.Exist(username) {
		return types.User{}, fmt.Errorf("user %s does not exist", username)
	}

	user := users.Get(username)

	roles := make([]types.Role, 0, len(user.Roles))
	for _, i := range user.Roles {
		roles = append(roles, types.Role{
			Role:     i.Role,
			Database: i.Database,
		})
	}

	u := types.User{
		Username: user.Username,
		Roles:    roles,
	}

	return u, nil
}

func (r *ResourceUser) connect() (*mongo.Client, error) {
	opts := options.Client().ApplyURI(r.Uri)
	return mongo.Connect(opts)
}
