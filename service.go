package do

import (
	"context"
	"reflect"
	"time"

	"github.com/samber/do/v2/stacktrace"
	typetostring "github.com/samber/go-type-to-string"
)

// MaxInvocationFrames defines the maximum number of stack frames to capture
// when tracking service invocations for debugging and observability purposes.
var MaxInvocationFrames uint32 = 100

// ServiceType represents the different types of services that can be registered
// in the dependency injection container. Each type has different lifecycle
// and instantiation behavior.
type ServiceType string

const (
	// ServiceTypeLazy represents a service that is instantiated only when first requested.
	// The service instance is cached and reused for subsequent requests.
	ServiceTypeLazy ServiceType = "lazy"

	// ServiceTypeEager represents a service that is instantiated immediately when registered.
	// The service is always available and ready to use.
	ServiceTypeEager ServiceType = "eager"

	// ServiceTypeTransient represents a service that is recreated each time it is requested.
	// No singleton caching is performed, ensuring a fresh instance every time. It is basically a factory.
	ServiceTypeTransient ServiceType = "transient"

	// ServiceTypeAlias represents a service that is an alias to another service.
	// It provides a different interface or name for accessing an existing service.
	ServiceTypeAlias ServiceType = "alias"
)

// serviceTypeToIcon maps each service type to a visual icon for debugging
// and observability purposes in logs and UI displays.
var serviceTypeToIcon = map[ServiceType]string{
	ServiceTypeLazy:      "😴",
	ServiceTypeEager:     "🔁",
	ServiceTypeTransient: "🏭",
	ServiceTypeAlias:     "🔗",
}

// serviceWrapper[T] is the main interface that all services in the DI container must implement.
// It provides methods for serviceWrapper lifecycle management, health checking, and shutdown.
// The generic type T represents the type of the serviceWrapper instance.
type serviceWrapper[T any] interface {
	getName() string
	getTypeName() string
	getServiceType() ServiceType
	getReflectType() reflect.Type
	getInstanceAny(Injector) (any, error)
	getInstance(Injector) (T, error)
	isHealthchecker() bool
	healthcheck(context.Context) error
	isShutdowner() bool
	shutdown(context.Context) error
	clone(Injector) any
	source() (stacktrace.Frame, []stacktrace.Frame)
}

// serviceWrapperAny is a non-generic version of serviceWrapper[T] that provides access to
// service functionality without requiring type information. This is useful
// for internal operations where the specific type is not known.
type serviceWrapperAny interface {
	getName() string
	getTypeName() string
	getServiceType() ServiceType
	getReflectType() reflect.Type
	getInstanceAny(Injector) (any, error)
	// getInstance(Injector) (T, error) - Not available in non-generic interface
	isHealthchecker() bool
	healthcheck(context.Context) error
	isShutdowner() bool
	shutdown(context.Context) error
	clone(Injector) any
	source() (stacktrace.Frame, []stacktrace.Frame)
}

// Interface definitions for specific service capabilities.
// These interfaces allow for type-safe access to specific service methods
// without requiring the full serviceWrapper[T] interface.

type serviceWrapperGetName interface{ getName() string }
type serviceWrapperGetTypeName interface{ getTypeName() string }
type serviceWrapperGetServiceType interface{ getServiceType() ServiceType }
type serviceWrapperGetReflectType interface{ getReflectType() reflect.Type }
type serviceWrapperGetInstanceAny interface{ getInstanceAny(Injector) (any, error) }
type serviceWrapperGetInstance[T any] interface{ getInstance(Injector) (T, error) } //nolint:unused
type serviceWrapperIsHealthchecker interface{ isHealthchecker() bool }
type serviceWrapperHealthcheck interface{ healthcheck(context.Context) error }
type serviceWrapperIsShutdowner interface{ isShutdowner() bool }
type serviceWrapperShutdown interface{ shutdown(context.Context) error }
type serviceWrapperClone interface{ clone(Injector) any }
type serviceWrapperSource interface {
	source() (stacktrace.Frame, []stacktrace.Frame)
}
type serviceWrapperBuildTime interface {
	getBuildTime() (time.Duration, bool)
}

// Interface compliance checks to ensure serviceWrapper[T] implements all required interfaces.
// These compile-time checks help catch interface implementation errors early.
var _ serviceWrapperGetName = (serviceWrapper[int])(nil)
var _ serviceWrapperGetTypeName = (serviceWrapper[int])(nil)
var _ serviceWrapperGetServiceType = (serviceWrapper[int])(nil)
var _ serviceWrapperGetReflectType = (serviceWrapper[int])(nil)
var _ serviceWrapperGetInstanceAny = (serviceWrapper[int])(nil)
var _ serviceWrapperIsHealthchecker = (serviceWrapper[int])(nil)
var _ serviceWrapperHealthcheck = (serviceWrapper[int])(nil)
var _ serviceWrapperIsShutdowner = (serviceWrapper[int])(nil)
var _ serviceWrapperShutdown = (serviceWrapper[int])(nil)
var _ serviceWrapperClone = (serviceWrapper[int])(nil)
var _ serviceWrapperSource = (serviceWrapper[int])(nil)

// inferServiceName uses type inference to determine the service name
// based on the generic type parameter T. This is used internally
// to automatically generate service names from types.
func inferServiceName[T any]() string {
	return typetostring.GetType[T]()
}

// inferServiceProviderStacktrace extracts stacktrace information from a service
// for debugging and observability purposes. Transient services don't have
// provider stacktraces since they are recreated on each request.
func inferServiceProviderStacktrace(service serviceWrapperAny) (stacktrace.Frame, bool) {
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
		if lazy, ok := serviceAny.(serviceWrapperBuildTime); ok {
			buildTime, _ = lazy.getBuildTime()
		}

		return serviceInfo{
			name:             name,
			serviceType:      serviceAny.(serviceWrapperGetServiceType).getServiceType(),
			serviceBuildTime: buildTime,
			healthchecker:    serviceAny.(serviceWrapperIsHealthchecker).isHealthchecker(),
			shutdowner:       serviceAny.(serviceWrapperIsShutdowner).isShutdowner(),
		}, true
	}

	return serviceInfo{}, false
}

func serviceCanCastToGeneric[T any](service any) bool {
	if svc, ok := service.(serviceWrapperGetReflectType); ok {
		// we need type reflection here, because we don't want to invoke the service when not needed
		return typeCanCastToGeneric[T](svc.getReflectType())
	}

	return false
}

func serviceCanCastToType(service any, toType reflect.Type) bool {
	if svc, ok := service.(serviceWrapperGetReflectType); ok {
		// we need type reflection here, because we don't want to invoke the service when not needed
		return typeCanCastToType(svc.getReflectType(), toType)
	}

	return false
}
