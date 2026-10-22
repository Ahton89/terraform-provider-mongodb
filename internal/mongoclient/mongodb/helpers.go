package mongodb

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"terraform-provider-mongodb/internal/mongoclient/types"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

/* DATABASES */

// createDatabase creates a new database in the MongoDB.
func createDatabase(ctx context.Context, client *mongo.Client, name string) error {
	db := client.Database(name)
	collection := db.Collection("created_by_terraform")
	document := bson.D{{"created_at", time.Now().Format(time.RFC850)}}

	_, err := collection.InsertOne(ctx, document)
	return err
}

// databaseExists checks if the database already exists in the MongoDB.
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

// deleteDatabase deletes the database.
func deleteDatabase(ctx context.Context, client *mongo.Client, name string) error {
	return client.Database(name).Drop(ctx)
}

// isDefaultDatabase checks if the database is a default database.
func isDefaultDatabase(name string) bool {
	return slices.Contains(types.DefaultDatabases, name)
}

/* USERS */

// listUsers returns a list of users in the database.
func listUsers(ctx context.Context, client *mongo.Client) (types.Users, error) {
	r := types.Users{}

	err := client.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{
		{"usersInfo", 1},
	}).Decode(&r)

	return r, err
}

// isDefaultUser checks if the user is a default user.
func isDefaultUser(username string) bool {
	return slices.Contains(types.DefaultUsers, username)
}

// userExists checks if the user already exists in the database.
func userExists(ctx context.Context, client *mongo.Client, username string) (bool, error) {
	u, err := listUsers(ctx, client)
	if err != nil {
		return false, err
	}

	return u.Exist(username), nil
}

/* REPLICASET */

// getReplicaSetStatus returns the status of the replica set.
// Using the isReplicaSetReady function, we can check if the replica set is ready and has a primary node.
func getReplicaSetStatus(ctx context.Context, client *mongo.Client) (*types.ReplicaSetStatus, error) {
	status := types.ReplicaSetStatus{}

	err := client.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{{"replSetGetStatus", 1}}).Decode(&status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

// getReplicaSetConfig returns the configuration of the replica set.
func getReplicaSetConfig(ctx context.Context, client *mongo.Client) (*types.ReplicaSetConfig, error) {
	rsc := types.ReplicaSetConfig{}

	err := client.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{
		{"replSetGetConfig", 1},
	}).Decode(&rsc)

	if err != nil {
		var commandErr mongo.CommandError

		// NotYetInitialized
		if errors.As(err, &commandErr) && commandErr.Code == 94 {
			return &rsc, fmt.Errorf("replica set not initialized. Please create, plan and apply mongodb_replicaset resource first: %w", err)
		}

		// NoReplicationEnabled
		if errors.As(err, &commandErr) && commandErr.Code == 76 {
			return &rsc, fmt.Errorf("replication not enabled. Please add replSetName in your mongod.conf file, then create, plan and apply mongodb_replicaset resource first: %w", err)
		}

		return &rsc, fmt.Errorf("get replica set config failed with error: %w", err)
	}

	rsc.Config.ClearVersion()

	return &rsc, nil
}

func getReplicaSetConfigVersion(ctx context.Context, client *mongo.Client) (int64, error) {
	type resultType struct {
		Config struct {
			Version int64 `bson:"version"`
		} `bson:"config"`
	}

	var result resultType

	err := client.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{
		{Key: "replSetGetConfig", Value: 1},
	}).Decode(&result)

	if err != nil {
		return 0, err
	}

	if result.Config.Version == 0 {
		return 0, fmt.Errorf("something went wrong while getting replica set version. Either there is no field with key version or it is zero, which we don't expect")
	}

	return result.Config.Version, nil
}

// isReplicaSetReady checks if the replica set is ready and has a primary node.
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

/* DEFAULT RW CONCERN */

// getDefaultRWConcern returns the cluster-wide default read and write concern.
func getDefaultRWConcern(ctx context.Context, client *mongo.Client) (*types.RWConcernDefaults, error) {
	d := types.RWConcernDefaults{}

	err := client.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{{"getDefaultRWConcern", 1}}).Decode(&d)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

// setDefaultRWConcern sets the cluster-wide default read and write concern.
// A nil document leaves the current value unchanged, an empty one unsets it.
func setDefaultRWConcern(ctx context.Context, client *mongo.Client, readConcern, writeConcern bson.D) error {
	cmd := bson.D{{"setDefaultRWConcern", 1}}

	if readConcern != nil {
		cmd = append(cmd, bson.E{Key: "defaultReadConcern", Value: readConcern})
	}

	if writeConcern != nil {
		cmd = append(cmd, bson.E{Key: "defaultWriteConcern", Value: writeConcern})
	}

	cmd = append(cmd, bson.E{Key: "writeConcern", Value: bson.D{{"w", "majority"}}})

	return client.Database(types.DefaultDatabase).RunCommand(ctx, cmd).Err()
}

func readConcernDocument(rc *types.ReadConcern) bson.D {
	if rc == nil {
		return nil
	}

	return bson.D{{"level", rc.Level}}
}

func writeConcernDocument(wc *types.WriteConcern) bson.D {
	if wc == nil {
		return nil
	}

	doc := bson.D{{"w", writeConcernW(wc.W)}}

	if wc.Journal != nil {
		doc = append(doc, bson.E{Key: "j", Value: *wc.Journal})
	}

	if wc.WTimeout != nil {
		doc = append(doc, bson.E{Key: "wtimeout", Value: *wc.WTimeout})
	}

	return doc
}

// writeConcernW converts numeric w to a number, MongoDB treats a string w as a tag set name.
func writeConcernW(w string) any {
	if n, err := strconv.ParseInt(w, 10, 32); err == nil {
		return int32(n)
	}

	return w
}

// requiredVersion checks if the current MongoDB version is supported by the provider.
// Supported versions: 6.x, 7.x, and 8.x
func requiredVersion(ctx context.Context, client *mongo.Client) error {
	var v struct {
		Version string `bson:"version"`
	}

	err := client.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{{"buildInfo", 1}}).Decode(&v)
	if err != nil {
		return fmt.Errorf("failed to get MongoDB version: %w", err)
	}

	// Extract major version
	parts := strings.Split(v.Version, ".")
	if len(parts) == 0 {
		return fmt.Errorf("invalid MongoDB version format: %s", v.Version)
	}

	majorVersion := parts[0]

	// Check if major version is supported
	for _, supported := range types.SupportedMajorVersions {
		if majorVersion == supported {
			return nil
		}
	}

	return fmt.Errorf("unsupported MongoDB version: %s. Supported versions: %s",
		v.Version, types.MongoDBSupportedVersions)
}
