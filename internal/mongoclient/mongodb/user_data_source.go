package mongodb

import (
	"context"
	"fmt"

	"terraform-provider-mongodb/internal/mongoclient/types"

	"github.com/avast/retry-go/v4"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (d *DataSourceUser) Read(ctx context.Context) (types.Users, error) {
	us := types.Users{}

	err := retry.Do(
		func() error {
			c, err := d.connect(ctx)
			if err != nil {
				return fmt.Errorf("connection to MongoDB failed with error: %s", err)
			}

			defer func() {
				_ = c.Disconnect(ctx)
			}()

			list, err := listUsers(ctx, c)
			if err != nil {
				return fmt.Errorf("list users failed with error: %s", err)
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
				return retry.Unrecoverable(fmt.Errorf("users not found. You can create a user using resource mongodb_user"))
			}

			return nil
		},
		retry.Attempts(d.RetryAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(d.RetryDelay),
		retry.Context(ctx),
	)

	return us, err
}

func (d *DataSourceUser) connect(ctx context.Context) (*mongo.Client, error) {
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
