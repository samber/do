package do

import (
	"reflect"
	"strings"
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
	typeOfT := reflect.TypeOf(t)
	prefix := ""

	if typeOfT == nil {
		typeOfT = reflect.TypeOf(&t)
	}

	for {
		if typeOfT.Kind() == reflect.Pointer {
			typeOfT = typeOfT.Elem()
			prefix += "*"
		} else if typeOfT.Kind() == reflect.Slice || typeOfT.Kind() == reflect.Array {
			typeOfT = typeOfT.Elem()
			prefix += "[]"
		} else {
			break
		}
	}

	if typeOfT.Name() == "" {
		// @TODO: handle "any" and "interface{}" types
	}

	pkgPath := typeOfT.PkgPath()
	if pkgPath != "" {
		pkgPath += "."
	}
	return pkgPath + prefix + typeOfT.Name()
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
