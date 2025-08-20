package database

import (
	"fmt"

	"github.com/samber/do/v2"
)

// DatabasePackage is the global package for database-related services
var DatabasePackage = do.Package(
	do.Lazy(func(i do.Injector) (*Database, error) {
		config := do.MustInvokeAs[Configuration](i)
		return &Database{
			URL:       fmt.Sprintf("postgres://localhost:5432/%s", config.GetAppName()),
			Connected: false,
		}, nil
	}),
	do.Eager(&Cache{
		Data: make(map[string]interface{}),
	}),
)
