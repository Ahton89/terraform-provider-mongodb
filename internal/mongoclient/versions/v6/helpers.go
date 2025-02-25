package v6

import (
	"context"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"terraform-provider-mongodb/internal/mongoclient/types"
)

/* DATABASES */

func listDatabases(ctx context.Context, client *mongo.Client) ([]string, error) {
	d, err := client.ListDatabaseNames(
		ctx,
		bson.M{
			"name": bson.M{
				"$nin": defaultDatabases,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func createDatabase(ctx context.Context, client *mongo.Client, name string) error {
	db := client.Database(name)
	collection := db.Collection("created_by_terraform")
	document := bson.D{{"created_at", time.Now().Format(time.RFC850)}}

	_, err := collection.InsertOne(ctx, document)
	return err
}

func databaseExists(ctx context.Context, client *mongo.Client, name string) (bool, error) {
	d, err := client.ListDatabaseNames(
		ctx,
		bson.M{
			"name": name,
		},
	)
	if err != nil {
		return false, err
	}

	return len(d) > 0, nil
}

func deleteDatabase(ctx context.Context, client *mongo.Client, name string) error {
	return client.Database(name).Drop(ctx)
}

func isDefaultDatabase(name string) bool {
	return slices.Contains(defaultDatabases, name)
}

/* USERS */

func listUsers(ctx context.Context, client *mongo.Client) (types.Users, error) {
	r := types.Users{}

	err := client.Database(defaultDatabase).RunCommand(ctx, bson.D{
		{"usersInfo", 1},
	}).Decode(&r)

	return r, err
}

func isDefaultUser(username string) bool {
	return slices.Contains(defaultUsers, username)
}

func userExists(ctx context.Context, client *mongo.Client, username string) (bool, error) {
	u, err := listUsers(ctx, client)
	if err != nil {
		return false, err
	}

	return u.Exist(username), nil
}
