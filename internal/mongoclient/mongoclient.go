package mongoclient

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"terraform-provider-mongodb/internal/mongoclient/interfaces"
	"terraform-provider-mongodb/internal/mongoclient/mongodb"
	"terraform-provider-mongodb/internal/mongoclient/types"
)

var (
	requiredVersion = "0" // Required MongoDB version
)

type client struct {
	uri string
}

func New(uri string) interfaces.Client {
	return &client{
		uri: uri,
	}
}

func (c *client) DataSource() interfaces.DataSource {
	return &mongodb.DataSource{
		Uri: c.uri,
	}
}
func (c *client) Resource() interfaces.Resource {
	return &mongodb.Resource{
		Uri: c.uri,
	}
}

func (c *client) RequiredVersion(ctx context.Context) error {
	var mc *mongo.Client
	var err error

	mc, err = c.connect()
	if err != nil {
		return err
	}

	defer func() {
		_ = mc.Disconnect(ctx)
	}()

	var v struct {
		Version string `bson:"version"`
	}

	err = mc.Database(types.DefaultDatabase).RunCommand(ctx, bson.D{{"buildInfo", 1}}).Decode(&v)
	if err != nil {
		return fmt.Errorf("failed to get MongoDB version: %s", err)
	}

	vPrefix := fmt.Sprintf("%s.", requiredVersion)

	if !strings.HasPrefix(v.Version, vPrefix) {
		return fmt.Errorf("unsupported MongoDB version. Current version is %s, but provider required only %s version", v.Version, requiredVersion)
	}

	return nil
}

func (c *client) connect() (*mongo.Client, error) {
	// Use direct connection to avoid replica set check
	// This connection needs for version check only
	opts := options.Client().ApplyURI(c.uri)
	opts.ReplicaSet = nil
	opts.Hosts = []string{opts.Hosts[0]}
	opts.SetDirect(true)

	return mongo.Connect(opts)
}
