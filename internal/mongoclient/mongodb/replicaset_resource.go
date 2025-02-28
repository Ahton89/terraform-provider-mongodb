package mongodb

import (
	"context"
	"errors"
	"fmt"

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

	err = c.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{
		{"replSetInitiate", plan},
	}).Err()
	if err != nil {
		return fmt.Errorf("create replica set failed with error: %s", err)
	}

	return nil
}

func (r *ResourceReplicaSet) Exists(ctx context.Context, state types.ReplicaSet) (bool, error) {
	rsc := types.ReplicaSetResponse{}

	c, err := r.connect()
	if err != nil {
		return false, fmt.Errorf("connection to MongoDB failed with error: %s", err)
	}

	defer func() {
		_ = c.Disconnect(ctx)
	}()

	err = c.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{
		{"replSetGetConfig", 1},
	}).Decode(&rsc)

	if err != nil {
		var commandErr mongo.CommandError

		// NotYetInitialized
		if errors.As(err, &commandErr) && commandErr.Code == 94 {
			return false, fmt.Errorf("replica set not initialized. Please create, plan and apply mongodb_replicaset resource first")
		}

		// NoReplicationEnabled
		if errors.As(err, &commandErr) && commandErr.Code == 76 {
			return false, fmt.Errorf("replication not enabled. Please add replSetName in your mongod.conf file, then create, plan and apply mongodb_replicaset resource first")
		}

		return false, fmt.Errorf("get replica set config failed with error: %s", err)
	}

	return rsc.Config.Name == state.Name, nil
}

func (r *ResourceReplicaSet) Ready(ctx context.Context, state types.ReplicaSet) (bool, error) {
	c, err := r.connect()
	if err != nil {
		return false, fmt.Errorf("connection to MongoDB failed with error: %s", err)
	}

	defer func() {
		_ = c.Disconnect(ctx)
	}()

	status, err := getReplicaSetStatus(ctx, c)
	if err != nil {
		return false, fmt.Errorf("get replica set status failed with error: %s", err)
	}

	return isReplicaSetReady(status, state.Name), nil
}

func (r *ResourceReplicaSet) Update(ctx context.Context, state types.ReplicaSet) error {
	c, err := r.connect()
	if err != nil {
		return fmt.Errorf("connection to MongoDB failed with error: %s", err)
	}

	defer func() {
		_ = c.Disconnect(ctx)
	}()

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

	return nil
}

func (r *ResourceReplicaSet) ImportState(ctx context.Context, name string) (types.ReplicaSet, error) {
	rsc := types.ReplicaSetResponse{}

	c, err := r.connect()
	if err != nil {
		return types.ReplicaSet{}, fmt.Errorf("connection to MongoDB failed with error: %s", err)
	}

	defer func() {
		_ = c.Disconnect(ctx)
	}()

	err = c.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{
		{"replSetGetConfig", 1},
	}).Decode(&rsc)

	if err != nil {
		var commandErr mongo.CommandError

		// NotYetInitialized
		if errors.As(err, &commandErr) && commandErr.Code == 94 {
			return types.ReplicaSet{}, fmt.Errorf("replica set not initialized. Please create, plan and apply mongodb_replicaset resource first")
		}

		// NoReplicationEnabled
		if errors.As(err, &commandErr) && commandErr.Code == 76 {
			return types.ReplicaSet{}, fmt.Errorf("replication not enabled. Please add replSetName in your mongod.conf file, then create, plan and apply mongodb_replicaset resource first")
		}

		return types.ReplicaSet{}, fmt.Errorf("get replica set config failed with error: %s", err)
	}

	if rsc.Config.Name != name {
		return types.ReplicaSet{}, fmt.Errorf("replica set %s does not exist", name)
	}

	status, err := getReplicaSetStatus(ctx, c)
	if err != nil {
		return types.ReplicaSet{}, fmt.Errorf("get replica set status failed with error: %s", err)
	}

	if !isReplicaSetReady(status, name) {
		return types.ReplicaSet{}, fmt.Errorf("replica set %s not ready or corrupted", name)
	}

	rsc.Config.RemoveDefaults()

	return rsc.Config, nil
}

func (r *ResourceReplicaSet) connect() (*mongo.Client, error) {
	opts := options.Client().ApplyURI(r.Uri)
	return mongo.Connect(opts)
}

func (r *ResourceReplicaSet) directConnect() (*mongo.Client, error) {
	opts := options.Client().ApplyURI(r.Uri)
	opts.ReplicaSet = nil
	opts.Hosts = []string{opts.Hosts[0]}
	opts.SetDirect(true)

	return mongo.Connect(opts)
}
