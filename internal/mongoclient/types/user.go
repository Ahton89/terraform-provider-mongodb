package types

import "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"

var (
	DefaultUsers = []string{"admin"} // Default users to exclude from listing
)

/* USER */

type Users struct {
	Users []User `tfsdk:"users" bson:"users"`
}

type User struct {
	Username string         `tfsdk:"username" bson:"user"`
	Password string         `tfsdk:"password" bson:"password,omitempty"`
	Roles    []Role         `tfsdk:"roles" bson:"roles"`
	Timeouts timeouts.Value `tfsdk:"timeouts" bson:"-"`
}

type Role struct {
	Role     string `tfsdk:"role" bson:"role"`
	Database string `tfsdk:"database" bson:"db"`
}

func (u *Users) Exist(username string) bool {
	for _, user := range u.Users {
		if user.Username == username {
			return true
		}
	}

	return false
}

func (u *Users) Get(username string) *User {
	for _, user := range u.Users {
		if user.Username == username {
			return &user
		}
	}

	return nil
}
