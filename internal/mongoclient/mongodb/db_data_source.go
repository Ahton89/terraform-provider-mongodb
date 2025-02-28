package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"terraform-provider-mongodb/internal/mongoclient/types"
)

func (d *DataSourceDatabase) Read(ctx context.Context) (types.Databases, error) {
	var c *mongo.Client

	var list []string
	var err error

	ds := types.Databases{}

	c, err = d.connect()
	if err != nil {
		return ds, fmt.Errorf("connection to MongoDB failed with error: %s", err)
	}

	defer func() {
		_ = c.Disconnect(ctx)
	}()

	list, err = c.ListDatabaseNames(
		ctx,
		bson.M{
			"name": bson.M{
				"$nin": types.DefaultDatabases,
			},
		},
	)
	if err != nil {
		return ds, fmt.Errorf("list databases failed with error: %s", err)
	}

	for _, i := range list {
		ds.Databases = append(ds.Databases, types.Database{
			Name: i,
		})
	}

	if len(list) == 0 {
		return ds, fmt.Errorf("databases not found. You can create a database using resource mongodb_database")
	}

	return ds, nil
}

func (d *DataSourceDatabase) connect() (*mongo.Client, error) {
	opts := options.Client().ApplyURI(d.Uri)
	return mongo.Connect(opts)
}
