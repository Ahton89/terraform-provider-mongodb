package mongoclient

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"terraform-provider-mongodb/internal/mongoclient/interfaces"
	"terraform-provider-mongodb/internal/mongoclient/versions/v6"
)

var requiredVersion = "0"

type client struct {
	mongoDB         *mongo.Client
	version         string
	provider        interfaces.Provider
	requiredVersion string
}

func New(ctx context.Context, uri string) (interfaces.Client, error) {
	var c *mongo.Client
	var v string
	var err error

	c, err = connect(uri)
	if err != nil {
		return nil, err
	}

	v, err = version(ctx, c)
	if err != nil {
		return nil, err
	}

	s := &client{
		mongoDB:         c,
		version:         v,
		requiredVersion: requiredVersion,
	}

	vPrefix := fmt.Sprintf("%s.", requiredVersion)

	switch {
	case strings.HasPrefix(v, vPrefix):
		s.provider = &v6.Provider{Client: c}
	default:
		return nil, fmt.Errorf("unsupported MongoDB version: %s, required version: %s", v, s.requiredVersion)
	}

	return s, nil
}

func (c *client) Disconnect(ctx context.Context) error {
	return c.mongoDB.Disconnect(ctx)
}

func (c *client) Ping(ctx context.Context) error {
	return c.mongoDB.Ping(ctx, nil)
}

func (c *client) Provider() interfaces.Provider {
	return c.provider
}

func (c *client) Version() string {
	return c.version
}

func version(ctx context.Context, client *mongo.Client) (string, error) {
	var v struct {
		Version string `bson:"version"`
	}

	err := client.Database("admin").RunCommand(
		ctx,
		bson.D{{"buildInfo", 1}},
	).Decode(&v)
	if err != nil {
		return "", err
	}

	return v.Version, nil
}

func connect(uri string) (*mongo.Client, error) {
	return mongo.Connect(
		options.Client().ApplyURI(uri),
		options.Client().SetDirect(true),
		options.Client().SetReadPreference(readpref.Primary()),
	)
}
