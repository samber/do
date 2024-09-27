package do

import (
	"reflect"
)

type Service[T any] interface {
	getName() string
	getInstance(*Injector) (T, error)
	healthcheck() error
	shutdown() error
	clone() any
}

type healthcheckableService interface {
	healthcheck() error
}

type shutdownableService interface {
	shutdown() error
}

func generateServiceName[T any]() string {
	var t T

	// struct
	name := reflect.TypeOf(t).String()
	if name != "<nil>" {
		return name
	}

	// interface
	return reflect.TypeOf((new(T))).String()
}

type Healthcheckable interface {
	HealthCheck() error
}

type Shutdownable interface {
	Shutdown() error
}

type cloneableService interface {
	clone() any
}
