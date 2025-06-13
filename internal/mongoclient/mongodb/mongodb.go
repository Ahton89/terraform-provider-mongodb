package mongodb

import (
	"time"

	"terraform-provider-mongodb/internal/mongoclient/interfaces"
)

/* DATA SOURCE */

type DataSource struct {
	Uri           string
	RetryAttempts uint
	RetryDelay    time.Duration
}

type DataSourceDatabase struct {
	Uri           string
	RetryAttempts uint
	RetryDelay    time.Duration
}

type DataSourceUser struct {
	Uri           string
	RetryAttempts uint
	RetryDelay    time.Duration
}

type DataSourceReplicaSet struct {
	Uri           string
	RetryAttempts uint
	RetryDelay    time.Duration
}

func (d *DataSource) DataSource() interfaces.DataSource {
	return &DataSource{
		Uri:           d.Uri,
		RetryAttempts: d.RetryAttempts,
		RetryDelay:    d.RetryDelay,
	}
}

func (d *DataSource) User() interfaces.DataSourceUser {
	return &DataSourceUser{
		Uri:           d.Uri,
		RetryAttempts: d.RetryAttempts,
		RetryDelay:    d.RetryDelay,
	}
}

func (d *DataSource) Database() interfaces.DataSourceDatabase {
	return &DataSourceDatabase{
		Uri:           d.Uri,
		RetryAttempts: d.RetryAttempts,
		RetryDelay:    d.RetryDelay,
	}
}

func (d *DataSource) ReplicaSet() interfaces.DataSourceReplicaSet {
	return &DataSourceReplicaSet{
		Uri:           d.Uri,
		RetryAttempts: d.RetryAttempts,
		RetryDelay:    d.RetryDelay,
	}
}

/* RESOURCE */

type Resource struct {
	Uri           string
	RetryAttempts uint
	RetryDelay    time.Duration
}

type ResourceDatabase struct {
	Uri           string
	RetryAttempts uint
	RetryDelay    time.Duration
}

type ResourceUser struct {
	Uri           string
	RetryAttempts uint
	RetryDelay    time.Duration
}

type ResourceReplicaSet struct {
	Uri           string
	RetryAttempts uint
	RetryDelay    time.Duration
}

func (r *Resource) Resource() interfaces.Resource {
	return &Resource{
		Uri:           r.Uri,
		RetryAttempts: r.RetryAttempts,
		RetryDelay:    r.RetryDelay,
	}
}

func (r *Resource) User() interfaces.ResourceUser {
	return &ResourceUser{
		Uri:           r.Uri,
		RetryAttempts: r.RetryAttempts,
		RetryDelay:    r.RetryDelay,
	}
}

func (r *Resource) Database() interfaces.ResourceDatabase {
	return &ResourceDatabase{
		Uri:           r.Uri,
		RetryAttempts: r.RetryAttempts,
		RetryDelay:    r.RetryDelay,
	}
}

func (r *Resource) ReplicaSet() interfaces.ResourceReplicaSet {
	return &ResourceReplicaSet{
		Uri:           r.Uri,
		RetryAttempts: r.RetryAttempts,
		RetryDelay:    r.RetryDelay,
	}
}
