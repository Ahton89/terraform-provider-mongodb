package mongodb

import (
	"terraform-provider-mongodb/internal/mongoclient/interfaces"
)

/* DATA SOURCE */

type DataSource struct {
	Uri string
}

type DataSourceDatabase struct {
	Uri string
}

type DataSourceUser struct {
	Uri string
}

type DataSourceReplicaSet struct {
	Uri string
}

func (d *DataSource) DataSource() interfaces.DataSource {
	return &DataSource{
		Uri: d.Uri,
	}
}

func (d *DataSource) User() interfaces.DataSourceUser {
	return &DataSourceUser{
		Uri: d.Uri,
	}
}

func (d *DataSource) Database() interfaces.DataSourceDatabase {
	return &DataSourceDatabase{
		Uri: d.Uri,
	}
}

func (d *DataSource) ReplicaSet() interfaces.DataSourceReplicaSet {
	return &DataSourceReplicaSet{
		Uri: d.Uri,
	}
}

/* RESOURCE */

type Resource struct {
	Uri string
}

type ResourceDatabase struct {
	Uri string
}

type ResourceUser struct {
	Uri string
}

type ResourceReplicaSet struct {
	Uri string
}

func (r *Resource) Resource() interfaces.Resource {
	return &Resource{
		Uri: r.Uri,
	}
}

func (r *Resource) User() interfaces.ResourceUser {
	return &ResourceUser{
		Uri: r.Uri,
	}
}

func (r *Resource) Database() interfaces.ResourceDatabase {
	return &ResourceDatabase{
		Uri: r.Uri,
	}
}

func (r *Resource) ReplicaSet() interfaces.ResourceReplicaSet {
	return &ResourceReplicaSet{
		Uri: r.Uri,
	}
}
