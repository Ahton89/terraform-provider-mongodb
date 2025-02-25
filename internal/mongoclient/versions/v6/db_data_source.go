package v6

import (
	"context"
	"fmt"

	"terraform-provider-mongodb/internal/mongoclient/types"
)

func (d *DataSourceDatabase) Read(ctx context.Context) (types.Databases, error) {
	ds := types.Databases{}

	list, err := listDatabases(ctx, d.Client)
	if err != nil {
		return ds, err
	}

	for _, i := range list {
		ds.Databases = append(ds.Databases, types.Database{
			Name: i,
		})
	}

	if len(list) == 0 {
		return ds, fmt.Errorf("no databases found")
	}

	return ds, nil
}
