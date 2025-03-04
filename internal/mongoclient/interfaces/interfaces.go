package interfaces

import (
	"context"

	"terraform-provider-mongodb/internal/mongoclient/types"
)

/* CLIENT */

type Client interface {
	DataSource() DataSource
	Resource() Resource
}

/* DATA SOURCE */

type DataSource interface {
	Database() DataSourceDatabase
	User() DataSourceUser
	ReplicaSet() DataSourceReplicaSet
}

type DataSourceDatabase interface {
	Read(ctx context.Context) (types.Databases, error)
}

type DataSourceUser interface {
	Read(ctx context.Context) (types.Users, error)
}

type DataSourceReplicaSet interface {
	Read(ctx context.Context) (types.ReplicaSet, error)
}

/* RESOURCE */

type Resource interface {
	Database() ResourceDatabase
	User() ResourceUser
	ReplicaSet() ResourceReplicaSet
}

type ResourceDatabase interface {
	Create(ctx context.Context, plan types.Database) error
	Delete(ctx context.Context, state types.Database) error
	Exists(ctx context.Context, state types.Database) (bool, error)
	ImportState(ctx context.Context, name string) (types.Database, error)
}

type ResourceUser interface {
	Create(ctx context.Context, plan types.User) error
	Delete(ctx context.Context, state types.User) error
	Update(ctx context.Context, plan types.User) error
	Exists(ctx context.Context, state types.User) (bool, error)
	ImportState(ctx context.Context, name string) (types.User, error)
}

type ResourceReplicaSet interface {
	Create(ctx context.Context, plan types.ReplicaSet) error
	Update(ctx context.Context, plan types.ReplicaSet) error
	Exists(ctx context.Context, state types.ReplicaSet) (bool, error)
	ImportState(ctx context.Context, name string) (types.ReplicaSet, error)
}
