package application

import (
	"github.com/samber/do/v2"
)

// ApplicationPackage is the global package for the main application
var ApplicationPackage = do.Package(
	do.Lazy(func(i do.Injector) (*Application, error) {
		return &Application{
			Config:       do.MustInvokeAs[Configuration](i),
			UserService:  do.MustInvokeAs[UserService](i),
			OrderService: do.MustInvokeAs[OrderService](i),
			Logger:       do.MustInvokeAs[Logger](i),
		}, nil
	}),
)
