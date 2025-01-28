package provider

import (
	"context"
	"strings"

	"github.com/blang/semver/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	defaultDatabases = []string{"admin", "config", "local"}
	defaultUsers     = []string{"admin"}
	minVersion       = "4.4.0"
)

type responseUsers struct {
	Users []responseUser `bson:"users"`
}

type responseUser struct {
	User  string         `bson:"user"`
	Roles []responseRole `bson:"roles"`
}

type responseRole struct {
	Role string `bson:"role"`
	Db   string `bson:"db"`
}

func (r *responseUsers) userExist(username string) bool {
	for _, user := range r.Users {
		if strings.EqualFold(user.User, username) {
			return true
		}
	}

	return false
}

func (r *responseUsers) getUser(username string) *responseUser {
	for _, user := range r.Users {
		if strings.EqualFold(user.User, username) {
			return &user
		}
	}

	return nil
}

// supportedVersion checks if the MongoDB version is supported
func supportedVersion(ctx context.Context, client *mongo.Client) bool {
	buildInfo := bson.M{}
	err := client.Database("admin").RunCommand(ctx, bson.D{{"buildInfo", 1}}).Decode(&buildInfo)
	if err != nil {
		return false
	}

	version, ok := buildInfo["version"].(string)
	if !ok {
		return false
	}

	semverVersion, err := semver.Parse(version)
	if err != nil {
		return false
	}

	semverMinVersion := semver.MustParse(minVersion)

	if semverVersion.LT(semverMinVersion) {
		return false
	}

	return true
}
