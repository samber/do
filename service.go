package do

import (
	"context"
	"time"

	"github.com/samber/do/v2/stacktrace"
	typetostring "github.com/samber/go-type-to-string"
)

var MaxInvokationFrames uint32 = 100

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
	getTypeName() string
	getServiceType() ServiceType
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

// Like Service[T] but without the generic type.
type ServiceAny interface {
	getName() string
	getTypeName() string
	getServiceType() ServiceType
	getEmptyInstance() any
	getInstanceAny(Injector) (any, error)
	// getInstance(Injector) (T, error)
	isHealthchecker() bool
	healthcheck(context.Context) error
	isShutdowner() bool
	shutdown(context.Context) error
	clone() any
	source() (stacktrace.Frame, []stacktrace.Frame)
}

type serviceGetName interface{ getName() string }
type serviceGetTypeName interface{ getTypeName() string }
type serviceGetServiceType interface{ getServiceType() ServiceType }
type serviceGetEmptyInstance interface{ getEmptyInstance() any }
type serviceGetInstanceAny interface{ getInstanceAny(Injector) (any, error) }
type serviceGetInstance[T any] interface{ getInstance(Injector) (T, error) } //nolint:unused
type serviceIsHealthchecker interface{ isHealthchecker() bool }
type serviceHealthcheck interface{ healthcheck(context.Context) error }
type serviceIsShutdowner interface{ isShutdowner() bool }
type serviceShutdown interface{ shutdown(context.Context) error }
type serviceClone interface{ clone() any }
type serviceSource interface {
	source() (stacktrace.Frame, []stacktrace.Frame)
}
type serviceBuildTime interface {
	getBuildTime() (time.Duration, bool)
}

var _ serviceGetName = (Service[int])(nil)
var _ serviceGetTypeName = (Service[int])(nil)
var _ serviceGetServiceType = (Service[int])(nil)
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

func inferServiceProviderStacktrace(service ServiceAny) (stacktrace.Frame, bool) {
	if service.getServiceType() == ServiceTypeTransient {
		return stacktrace.Frame{}, false
	} else {
		providerFrame, _ := service.source()
		return providerFrame, true
	}
}

type serviceInfo struct {
	name             string
	serviceType      ServiceType
	serviceBuildTime time.Duration
	healthchecker    bool
	shutdowner       bool
}

func inferServiceInfo(injector Injector, name string) (serviceInfo, bool) {
	if serviceAny, ok := injector.serviceGet(name); ok {
		var buildTime time.Duration
		if lazy, ok := serviceAny.(serviceBuildTime); ok {
			buildTime, _ = lazy.getBuildTime()
		}

		return serviceInfo{
			name:             name,
			serviceType:      serviceAny.(serviceGetServiceType).getServiceType(),
			serviceBuildTime: buildTime,
			healthchecker:    serviceAny.(serviceIsHealthchecker).isHealthchecker(),
			shutdowner:       serviceAny.(serviceIsShutdowner).isShutdowner(),
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
