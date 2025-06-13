package mongoclient

import (
	"time"

	"terraform-provider-mongodb/internal/mongoclient/interfaces"
	"terraform-provider-mongodb/internal/mongoclient/mongodb"
	"terraform-provider-mongodb/internal/mongoclient/types"
)

type client struct {
	uri           string
	retryAttempts uint
	retryDelay    time.Duration
}

func New(uri string, retryAttempts, retryDelay uint) interfaces.Client {
	// Validate the retry parameters
	if retryAttempts == 0 {
		retryAttempts = types.RetryAttempts
	}

	if retryDelay == 0 {
		retryDelay = types.RetryDelaySec
	}

	return &client{
		uri:           uri,
		retryAttempts: retryAttempts,
		retryDelay:    time.Duration(retryDelay) * time.Second,
	}
}

func (c *client) DataSource() interfaces.DataSource {
	return &mongodb.DataSource{
		Uri:           c.uri,
		RetryAttempts: c.retryAttempts,
		RetryDelay:    c.retryDelay,
	}
}
func (c *client) Resource() interfaces.Resource {
	return &mongodb.Resource{
		Uri:           c.uri,
		RetryAttempts: c.retryAttempts,
		RetryDelay:    c.retryDelay,
	}
}
