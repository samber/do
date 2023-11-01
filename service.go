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
	getEmptyInstance() any
	getInstanceAny(Injector) (any, error)
	getInstance(Injector) (T, error)
	isHealthchecker() bool
	healthcheck(context.Context) error
	isShutdowner() bool
	shutdown(context.Context) error
	clone() any
	source() (stacktrace.Frame, []stacktrace.Frame)
}

type serviceGetName interface {
	getName() string
}

type serviceGetType interface {
	getType() ServiceType
}

type serviceGetEmptyInstance interface {
	getEmptyInstance() any
}

type serviceGetInstanceAny interface {
	getInstanceAny(Injector) (any, error)
}

type serviceGetInstance[T any] interface {
	getInstance(Injector) (T, error)
}

type serviceIsHealthchecker interface {
	isHealthchecker() bool
}

type serviceHealthcheck interface {
	healthcheck(context.Context) error
}

type serviceIsShutdowner interface {
	isShutdowner() bool
}

type serviceShutdown interface {
	shutdown(context.Context) error
}

type serviceClone interface {
	clone() any
}

type serviceSource interface {
	source() (stacktrace.Frame, []stacktrace.Frame)
}

var _ serviceGetName = (Service[int])(nil)
var _ serviceGetType = (Service[int])(nil)
var _ serviceGetEmptyInstance = (Service[int])(nil)
var _ serviceGetInstanceAny = (Service[int])(nil)
var _ serviceIsHealthchecker = (Service[int])(nil)
var _ serviceHealthcheck = (Service[int])(nil)
var _ serviceIsShutdowner = (Service[int])(nil)
var _ serviceShutdown = (Service[int])(nil)
var _ serviceClone = (Service[int])(nil)
var _ serviceSource = (Service[int])(nil)

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
			serviceType:   serviceAny.(serviceGetType).getType(),
			healthchecker: serviceAny.(serviceIsHealthchecker).isHealthchecker(),
			shutdowner:    serviceAny.(serviceIsShutdowner).isShutdowner(),
		}, true
	}

	return serviceInfo{}, false
}

func serviceIsAssignable[T any](service any) bool {
	if svc, ok := service.(serviceGetEmptyInstance); ok {
		// we need an empty instance here, because we don't want to instantiate the service when not needed
		if _, ok = svc.getEmptyInstance().(T); ok {
			return true
		}
	}
	return false
}
