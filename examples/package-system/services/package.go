package services

import (
	"github.com/samber/do/v2"
)

// ServicesPackage is the global package for business logic services.
var ServicesPackage = do.Package(
	do.Lazy(func(i do.Injector) (*UserService, error) {
		return &UserService{
			DB:     do.MustInvokeAs[Database](i),
			Cache:  do.MustInvokeAs[Cache](i),
			Logger: do.MustInvokeAs[Logger](i),
		}, nil
	}),
	do.Lazy(func(i do.Injector) (*OrderService, error) {
		return &OrderService{
			DB:     do.MustInvokeAs[Database](i),
			Cache:  do.MustInvokeAs[Cache](i),
			Logger: do.MustInvokeAs[Logger](i),
		}, nil
	}),
)
