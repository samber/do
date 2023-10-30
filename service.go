package do

import (
	"context"

	"github.com/samber/do/v2/stacktrace"
	typetostring "github.com/samber/go-type-to-string"
)

type ServiceType string

const (
	ServiceTypeLazy      ServiceType = "lazy"
	ServiceTypeEager     ServiceType = "eager"
	ServiceTypeTransient ServiceType = "transient"
	ServiceTypeAlias     ServiceType = "alias"
)

var serviceTypeToIcon = map[ServiceType]string{
	ServiceTypeLazy:      "😴",
	ServiceTypeEager:     "🔁",
	ServiceTypeTransient: "🏭",
	ServiceTypeAlias:     "🔗",
}

type Service[T any] interface {
	getName() string
	getType() ServiceType
	getInstance(Injector) (T, error)
	isHealthchecker() bool
	healthcheck(context.Context) error
	isShutdowner() bool
	shutdown(context.Context) error
	clone() any
	source() (stacktrace.Frame, []stacktrace.Frame)
}

type Healthchecker interface {
	HealthCheck() error
}

type HealthcheckerWithContext interface {
	HealthCheckWithContext(context.Context) error
}

type Shutdowner interface {
	Shutdown()
}

type ShutdownerWithError interface {
	Shutdown() error
}

type ShutdownerWithContext interface {
	Shutdown(context.Context)
}

type ShutdownerWithContextAndError interface {
	Shutdown(context.Context) error
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
	healthcheck(context.Context) error
}

type isShutdownerService interface {
	isShutdowner() bool
}

type shutdownerService interface {
	shutdown(context.Context) error
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

func inferServiceProviderStacktrace[T any](service Service[T]) (stacktrace.Frame, bool) {
	if service.getType() == ServiceTypeTransient {
		return stacktrace.Frame{}, false
	} else {
		providerFrame, _ := service.source()
		return providerFrame, true
	}
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
