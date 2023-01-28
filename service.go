package do

import (
	"fmt"
	"reflect"
)

type Service[T any] interface {
	getName() string
	getInstance(*Injector) (T, error)
	// getInstanceAny(*Injector) (any, error)
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

type anyService interface {
	getInstanceAny(*Injector) (any, error)
}

func getServiceNameFromStructField(field *reflect.StructField) string {
	t := reflect.New(field.Type).Elem().Interface()
	return getServiceNameFromValue(t)
}

func generateServiceName[T any]() string {
	var t T
	return getServiceNameFromValue(t)
}

func getServiceNameFromValue[T any](t T) string {
	// struct
	name := fmt.Sprintf("%T", t)
	if name != "<nil>" {
		return name
	}

	// interface
	return fmt.Sprintf("%T", new(T))
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
