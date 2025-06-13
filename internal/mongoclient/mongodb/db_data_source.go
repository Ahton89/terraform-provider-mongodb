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

func (d *DataSourceDatabase) Read(ctx context.Context) (types.Databases, error) {
	ds := types.Databases{}

	err := retry.Do(
		func() error {
			c, err := d.connect(ctx)
			if err != nil {
				return fmt.Errorf("connection to MongoDB failed with error: %s", err)
			}

			defer func() {
				_ = c.Disconnect(ctx)
			}()

			list, err := c.ListDatabaseNames(
				ctx,
				bson.M{
					"name": bson.M{
						"$nin": types.DefaultDatabases,
					},
				},
			)
			if err != nil {
				return fmt.Errorf("list databases failed with error: %s", err)
			}

			for _, i := range list {
				ds.Databases = append(ds.Databases, types.Database{
					Name: i,
				})
			}

			if len(list) == 0 {
				return retry.Unrecoverable(fmt.Errorf("databases not found. You can create a database using resource mongodb_database"))
			}

			return nil
		},
		retry.Attempts(d.RetryAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(d.RetryDelay),
		retry.Context(ctx),
	)
	
	return ds, err
}

func (d *DataSourceDatabase) connect(ctx context.Context) (*mongo.Client, error) {
	opts := options.Client().ApplyURI(d.Uri)

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
