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
	star := 0

	if typeOfT == nil {
		typeOfT = reflect.TypeOf(new(T))
	}

	for {
		if typeOfT.Kind() != reflect.Pointer {
			break
		}
		typeOfT = typeOfT.Elem()
		star++
	}

	pkgPath := typeOfT.PkgPath()
	if pkgPath != "" {
		pkgPath += "."
	}
	return pkgPath + strings.Repeat("*", star) + typeOfT.Name()
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
