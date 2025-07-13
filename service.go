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
	// For non-pointer types, reflect.TypeOf(t) will never be nil.
	// For pointer types, reflect.TypeOf(t) can be nil if t is nil.
	typ := reflect.TypeOf(t)
	if typ == nil {
		return ""
	}

	if name := typ.String(); name != "" {
		return name
	}

	return reflect.TypeOf(new(T)).String()
}

// func generateServiceName[T any]() string {
// 	var t T
// 	// For non-pointer types, reflect.TypeOf(t) will never be nil.
// 	// For pointer types, reflect.TypeOf(t) can be nil if t is nil.
// 	typ := reflect.TypeOf(t)
// 	if typ == nil {
// 		return ""
// 	}
//
// 	name := typ.Name()
// 	if name != "" {
// 		return name
// 	}
//
// 	return reflect.TypeOf(new(T)).Name()
// }


type Healthcheckable interface {
	HealthCheck() error
}

type Shutdownable interface {
	Shutdown() error
}

type cloneableService interface {
	clone() any
}
