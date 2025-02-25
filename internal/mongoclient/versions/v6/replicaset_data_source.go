package v6

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"terraform-provider-mongodb/internal/mongoclient/types"
)

func (d *DataSourceReplicaSet) Read(ctx context.Context) (types.ReplicaSet, error) {
	rsc := types.ReplicaSetResponse{}

	err := d.Client.Database(defaultDatabase).RunCommand(ctx, bson.D{
		{"replSetGetConfig", 1},
	}).Decode(&rsc)

	if err != nil {
		var commandErr mongo.CommandError

		if errors.As(err, &commandErr) && commandErr.Code == 94 {
			return types.ReplicaSet{}, fmt.Errorf("replica set not initialized. Please create, plan and apply mongodb_replicaset resource first")
		}
		
		return types.ReplicaSet{}, err
	}

	return rsc.Config, nil
}
