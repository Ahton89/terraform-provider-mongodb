package mongodb

import (
	"context"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"terraform-provider-mongodb/internal/mongoclient/types"
)

/* DATABASES */

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
	return slices.Contains(types.DefaultDatabases, name)
}

/* USERS */

func listUsers(ctx context.Context, client *mongo.Client) (types.Users, error) {
	r := types.Users{}

	err := client.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{
		{"usersInfo", 1},
	}).Decode(&r)

	return r, err
}

func isDefaultUser(username string) bool {
	return slices.Contains(types.DefaultUsers, username)
}

func userExists(ctx context.Context, client *mongo.Client, username string) (bool, error) {
	u, err := listUsers(ctx, client)
	if err != nil {
		return false, err
	}

	return u.Exist(username), nil
}

/* REPLICASET */

func getReplicaSetStatus(ctx context.Context, client *mongo.Client) (*types.ReplicaSetStatus, error) {
	status := types.ReplicaSetStatus{}

	err := client.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{{"replSetGetStatus", 1}}).Decode(&status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

func isReplicaSetReady(status *types.ReplicaSetStatus, replicaSetName string) bool {
	if status.OK != 1 || status.Set != replicaSetName {
		return false
	}

	hasPrimary := false
	allMembersOk := true

	for _, member := range status.Members {
		if member.Health != 1 {
			allMembersOk = false
		}
		if member.StateStr == "PRIMARY" {
			hasPrimary = true
		}
	}

	return hasPrimary && allMembersOk
}
