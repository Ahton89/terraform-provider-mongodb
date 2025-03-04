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

func (d *DataSourceReplicaSet) Read(ctx context.Context) (types.ReplicaSet, error) {
	rsc := types.ReplicaSetResponse{}

	c, err := d.connect()
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

	return rsc.Config, nil
}

func (d *DataSourceReplicaSet) connect() (*mongo.Client, error) {
	opts := options.Client().ApplyURI(d.Uri)
	return mongo.Connect(opts)
}
