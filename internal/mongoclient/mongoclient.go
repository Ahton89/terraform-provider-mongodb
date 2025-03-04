package mongoclient

import (
	"terraform-provider-mongodb/internal/mongoclient/interfaces"
	"terraform-provider-mongodb/internal/mongoclient/mongodb"
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
