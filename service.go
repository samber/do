package do

import (
	"fmt"

	"github.com/samber/do/stacktrace"
	typetostring "github.com/samber/go-type-to-string"
)

type ServiceType string

const (
	ServiceTypeLazy      ServiceType = "lazy"
	ServiceTypeEager     ServiceType = "eager"
	ServiceTypeTransiant ServiceType = "transiant"
)

var serviceTypeToIcon = map[ServiceType]string{
	ServiceTypeLazy:      "😴",
	ServiceTypeEager:     "🔁",
	ServiceTypeTransiant: "🏭",
}

type Service[T any] interface {
	getName() string
	getType() ServiceType
	getInstance(Injector) (T, error)
	isHealthchecker() bool
	healthcheck() error
	isShutdowner() bool
	shutdown() error
	clone() any
	locate() (stacktrace.Frame, []stacktrace.Frame)
}

type Healthchecker interface {
	HealthCheck() error
}

type Shutdowner interface {
	Shutdown() error
}

var _ isHealthcheckerService = (Service[int])(nil)
var _ healthcheckerService = (Service[int])(nil)
var _ isShutdownerService = (Service[int])(nil)
var _ shutdownerService = (Service[int])(nil)
var _ clonerService = (Service[int])(nil)
var _ getTyperService = (Service[int])(nil)

type isHealthcheckerService interface {
	isHealthchecker() bool
}

type healthcheckerService interface {
	healthcheck() error
}

type isShutdownerService interface {
	isShutdowner() bool
}

type shutdownerService interface {
	shutdown() error
}

type clonerService interface {
	clone() any
}

type getTyperService interface {
	getType() ServiceType
}

func inferServiceName[T any]() string {
	return typetostring.GetType[T]()
}

func inferServiceType[T any](service Service[T]) ServiceType {
	switch service.(type) {
	case *ServiceLazy[T]:
		return ServiceTypeLazy
	case *ServiceEager[T]:
		return ServiceTypeEager
	case *ServiceTransiant[T]:
		return ServiceTypeTransiant
	}

	panic(fmt.Errorf("DI: unknown service type"))
}

func inferServiceStacktrace[T any](service Service[T]) (stacktrace.Frame, bool) {
	switch s := service.(type) {
	case *ServiceLazy[T]:
		return s.providerFrame, true
	case *ServiceEager[T]:
		return s.providerFrame, true
	case *ServiceTransiant[T]:
		return stacktrace.Frame{}, false
	}

	panic(fmt.Errorf("DI: unknown service type"))
}

type serviceInfo struct {
	name          string
	serviceType   ServiceType
	healthchecker bool
	shutdowner    bool
}

func inferServiceInfo(injector Injector, name string) (serviceInfo, bool) {
	if serviceAny, ok := injector.serviceGet(name); ok {
		return serviceInfo{
			name:          name,
			serviceType:   serviceAny.(getTyperService).getType(),
			healthchecker: serviceAny.(isHealthcheckerService).isHealthchecker(),
			shutdowner:    serviceAny.(isShutdownerService).isShutdowner(),
		}, true
	}

	return serviceInfo{}, false
}
