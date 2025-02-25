package v6

import (
	"context"
	"fmt"

	"terraform-provider-mongodb/internal/mongoclient/types"
)

func (d *DataSourceUser) Read(ctx context.Context) (types.Users, error) {
	us := types.Users{}

	list, err := listUsers(ctx, d.Client)
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
		return us, fmt.Errorf("no users found")
	}

	return us, nil
}
