package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"terraform-provider-mongodb/internal/mongoclient/types"
)

func (d *DataSourceUser) Read(ctx context.Context) (types.Users, error) {
	var c *mongo.Client
	var list types.Users
	var err error

	us := types.Users{}

	c, err = d.connect()
	if err != nil {
		return us, err
	}

	defer func() {
		_ = c.Disconnect(ctx)
	}()

	list, err = listUsers(ctx, c)
	if err != nil {
		return us, err
	}

	for _, i := range list.Users {
		if isDefaultUser(i.Username) {
			continue
		}

		r := make([]types.Role, 0, len(i.Roles))

		for _, j := range i.Roles {
			r = append(r, types.Role{
				Role:     j.Role,
				Database: j.Database,
			})
		}

		us.Users = append(us.Users, types.User{
			Username: i.Username,
			Password: i.Password,
			Roles:    r,
		})
	}

	if len(us.Users) == 0 {
		return us, fmt.Errorf("users not found. You can create a user using resource mongodb_user")
	}

	return us, nil
}

func (d *DataSourceUser) connect() (*mongo.Client, error) {
	opts := options.Client().ApplyURI(d.Uri)
	return mongo.Connect(opts)
}
