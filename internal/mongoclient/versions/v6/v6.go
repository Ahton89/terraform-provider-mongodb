package v6

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
	"terraform-provider-mongodb/internal/mongoclient/interfaces"
)

var (
	defaultDatabases = []string{"admin", "config", "local"} // Default databases to exclude from listing
	defaultUsers     = []string{"admin"}                    // Default users to exclude from listing
	defaultDatabase  = "admin"                              // Default database for making queries
)

type Provider struct {
	Client *mongo.Client
}

/* DATA SOURCE */

type DataSource struct {
	Client *mongo.Client
}

type DataSourceDatabase struct {
	Client *mongo.Client
}

type DataSourceUser struct {
	Client *mongo.Client
}

type DataSourceReplicaSet struct {
	Client *mongo.Client
}

func (p *Provider) DataSource() interfaces.DataSource {
	return &DataSource{Client: p.Client}
}

func (d *DataSource) User() interfaces.DataSourceUser {
	return &DataSourceUser{Client: d.Client}
}

func (d *DataSource) Database() interfaces.DataSourceDatabase {
	return &DataSourceDatabase{Client: d.Client}
}

func (d *DataSource) ReplicaSet() interfaces.DataSourceReplicaSet {
	return &DataSourceReplicaSet{Client: d.Client}
}

/* RESOURCE */

type Resource struct {
	Client *mongo.Client
}

type ResourceDatabase struct {
	Client *mongo.Client
}

type ResourceUser struct {
	Client *mongo.Client
}

type ResourceReplicaSet struct {
	Client *mongo.Client
}

func (p *Provider) Resource() interfaces.Resource {
	return &Resource{Client: p.Client}
}

func (r *Resource) User() interfaces.ResourceUser {
	return &ResourceUser{Client: r.Client}
}

func (r *Resource) Database() interfaces.ResourceDatabase {
	return &ResourceDatabase{Client: r.Client}
}

func (r *Resource) ReplicaSet() interfaces.ResourceReplicaSet {
	return &ResourceReplicaSet{Client: r.Client}
}
