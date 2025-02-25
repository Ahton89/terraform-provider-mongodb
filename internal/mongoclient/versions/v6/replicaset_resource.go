package v6

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"terraform-provider-mongodb/internal/mongoclient/types"
)

func (r *ResourceReplicaSet) Create(ctx context.Context, plan types.ReplicaSet) error {
	return r.Client.Database(defaultDatabase).RunCommand(ctx, bson.D{
		{"replSetInitiate", plan},
	}).Err()
}

func (r *ResourceReplicaSet) Exists(ctx context.Context, state types.ReplicaSet) (bool, error) {
	rsc := types.ReplicaSetResponse{}

	err := r.Client.Database(defaultDatabase).RunCommand(ctx, bson.D{
		{"replSetGetConfig", 1},
	}).Decode(&rsc)

	if err != nil {
		var commandErr mongo.CommandError

		if errors.As(err, &commandErr) && commandErr.Code == 94 {
			return false, fmt.Errorf("replica set not initialized. Please create, plan and apply mongodb_replicaset resource first")
		}

		return false, err
	}

	return rsc.Config.Name == state.Name, nil
}

func (r *ResourceReplicaSet) Update(ctx context.Context, state types.ReplicaSet) error {
	return r.Client.Database(defaultDatabase).RunCommand(ctx, bson.D{
		{"replSetReconfig", state},
	}).Err()
}

func (r *ResourceReplicaSet) ImportState(ctx context.Context, name string) (types.ReplicaSet, error) {
	rsc := types.ReplicaSetResponse{}

	err := r.Client.Database(defaultDatabase).RunCommand(ctx, bson.D{
		{"replSetGetConfig", 1},
	}).Decode(&rsc)

	if err != nil {
		var commandErr mongo.CommandError

		if errors.As(err, &commandErr) && commandErr.Code == 94 {
			return types.ReplicaSet{}, fmt.Errorf("replica set not initialized. Please create, plan and apply mongodb_replicaset resource first")
		}

		return types.ReplicaSet{}, err
	}

	if rsc.Config.Name != name {
		return types.ReplicaSet{}, fmt.Errorf("replica set %s does not exist", name)
	}

	rsc.Config.RemoveDefaults()

	return rsc.Config, nil
}
