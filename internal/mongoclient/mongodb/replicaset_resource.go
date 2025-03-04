package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"terraform-provider-mongodb/internal/mongoclient/types"
)

func (r *ResourceReplicaSet) Create(ctx context.Context, plan types.ReplicaSet) error {
	c, err := r.directConnect()
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

	err = c.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{
		{"replSetInitiate", plan},
	}).Err()
	if err != nil {
		return fmt.Errorf("create replica set failed with error: %s", err)
	}

	return r.waitForReplicaSetReady(ctx, plan.Name)
}

func (r *ResourceReplicaSet) Exists(ctx context.Context, state types.ReplicaSet) (bool, error) {
	var rsc *types.ReplicaSetConfig

	c, err := r.connect()
	if err != nil {
		return false, fmt.Errorf("connection to MongoDB failed with error: %s", err)
	}

	defer func() {
		_ = c.Disconnect(ctx)
	}()

	err = requiredVersion(ctx, c)
	if err != nil {
		return false, fmt.Errorf("required version check failed with error: %s", err)
	}

	rsc, err = getReplicaSetConfig(ctx, c)
	if err != nil {
		return false, fmt.Errorf("get replica set config failed with error: %s", err)
	}

	return rsc.Config.Name == state.Name, nil
}

func (r *ResourceReplicaSet) Update(ctx context.Context, state types.ReplicaSet) error {
	c, err := r.connect()
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

	status, err := getReplicaSetStatus(ctx, c)
	if err != nil {
		return fmt.Errorf("get replica set status failed with error: %s", err)
	}

	if !isReplicaSetReady(status, state.Name) {
		return fmt.Errorf("replica set %s not ready or corrupted", state.Name)
	}

	err = c.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{
		{"replSetReconfig", state},
	}).Err()
	if err != nil {
		return fmt.Errorf("updating replica set failed with error: %s", err)
	}

	return r.waitForReplicaSetReady(ctx, state.Name)
}

func (r *ResourceReplicaSet) ImportState(ctx context.Context, name string) (types.ReplicaSet, error) {
	var rsc *types.ReplicaSetConfig

	c, err := r.connect()
	if err != nil {
		return types.ReplicaSet{}, fmt.Errorf("connection to MongoDB failed with error: %s", err)
	}

	defer func() {
		_ = c.Disconnect(ctx)
	}()

	err = requiredVersion(ctx, c)
	if err != nil {
		return types.ReplicaSet{}, fmt.Errorf("required version check failed with error: %s", err)
	}

	rsc, err = getReplicaSetConfig(ctx, c)
	if err != nil {
		return types.ReplicaSet{}, fmt.Errorf("get replica set config failed with error: %s", err)
	}

	if rsc.Config.Name != name {
		return types.ReplicaSet{}, fmt.Errorf("replica set %s does not exist", name)
	}

	rsc.Config.RemoveDefaults()

	return rsc.Config, nil
}

func (r *ResourceReplicaSet) connect() (*mongo.Client, error) {
	opts := options.Client().ApplyURI(r.Uri)

	if opts.ReplicaSet == nil {
		return nil, fmt.Errorf("you can't use direct connection when working with replica set")
	}

	return mongo.Connect(opts)
}

func (r *ResourceReplicaSet) directConnect() (*mongo.Client, error) {
	opts := options.Client().ApplyURI(r.Uri)
	opts.ReplicaSet = nil
	opts.Hosts = []string{opts.Hosts[0]}
	opts.SetDirect(true)

	return mongo.Connect(opts)
}

func (r *ResourceReplicaSet) waitForReplicaSetReady(ctx context.Context, replicaSetName string) error {
	ticker := time.NewTicker(5 * time.Second)
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
			var status *types.ReplicaSetStatus
			var err error

			client, err = r.connect()
			if err != nil {
				continue
			}

			status, err = getReplicaSetStatus(ctx, client)
			_ = client.Disconnect(ctx)
			if err == nil && isReplicaSetReady(status, replicaSetName) {
				return nil
			}
		}
	}
}
