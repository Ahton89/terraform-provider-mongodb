package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"terraform-provider-mongodb/internal/mongoclient/types"
)

func (d *DataSourceReplicaSet) Read(ctx context.Context) (types.ReplicaSet, error) {
	var rsc *types.ReplicaSetConfig

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

	rsc, err = getReplicaSetConfig(ctx, c)
	if err != nil {
		return types.ReplicaSet{}, fmt.Errorf("get replica set config failed with error: %s", err)
	}

	return rsc.Config, nil
}

func (d *DataSourceReplicaSet) connect() (*mongo.Client, error) {
	opts := options.Client().ApplyURI(d.Uri)
	return mongo.Connect(opts)
}
