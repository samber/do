package do

import (
	"fmt"
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

func generateServiceNameFromInjector[T any](i *Injector) string {
	if i != nil && i.useFQSN {
		return generateServiceNameWithFQSN[T]()
	}
	return generateServiceName[T]()
}

func generateServiceName[T any]() string {
	return generateServiceNameWithReflect[T]()
}

//nolint:unused
func generateServiceNameWithSprintf[T any]() string {
	var t T

	// struct
	name := fmt.Sprintf("%T", t)
	if name != "<nil>" {
		return name
	}

	//interface
	return fmt.Sprintf("%T", new(T))
}

func generateServiceNameWithReflect[T any]() string {
	var t T
	// reflect.TypeOf(t) will be nil when T is an interface type.
	typ := reflect.TypeOf(t)
	if typ == nil {
		typ = reflect.TypeOf(new(T)).Elem()
	}

	return typ.String()
}

// generateServiceNameWithFQSN generates a fully qualified service name.
// It uses the package path and type name to create a unique identifier for the service.
// This is useful for services that are defined in different packages but have the same type name.
// Example: "github.com/user/project/service.MyService"
func generateServiceNameWithFQSN[T any]() string {
	var t T
	// reflect.TypeOf(t) will be nil when T is an interface type.
	typ := reflect.TypeOf(t)
	if typ == nil {
		typ = reflect.TypeOf(new(T)).Elem()
	}

	prefix := ""
	typName := typ
	if typ.Kind() == reflect.Ptr {
		prefix = "*"
		typName = typ.Elem()
	}

	name := typName.Name()
	pkg := typName.PkgPath()
	if name != "" && pkg != "" {
		return prefix + pkg + "." + name
	}

	return prefix + typName.String()
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
