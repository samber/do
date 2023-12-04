package do

import (
	"fmt"
	"sort"
	"strings"
)

// invokeByName looks for a service by its name.
func invokeByName[T any](i Injector, name string) (T, error) {
	var invokerChain []string

	injector := getInjectorOrDefault(i)

	vScope, isVirtualScope := injector.(*virtualScope)
	if isVirtualScope {
		invokerChain = vScope.invokerChain

		if err := vScope.detectCircularDependency(name); err != nil {
			return empty[T](), err
		}
	}

	serviceAny, serviceScope, found := injector.serviceGetRec(name)
	if !found {
		return empty[T](), serviceNotFound(injector, name)
	}

	if isVirtualScope {
		vScope.addDependency(injector, name, serviceScope)
	}

	service, ok := serviceAny.(Service[T])
	if !ok {
		return empty[T](), serviceNotFound(injector, name)
	}

	instance, err := service.getInstance(&virtualScope{invokerChain: append(invokerChain, name), self: serviceScope})
	if err != nil {
		return empty[T](), err
	}

	serviceScope.onServiceInvoke(name)

	injector.RootScope().opts.Logf("DI: service %s invoked", name)

	return instance, nil
}

// invokeByGenericType look for a service by its type.
// When many services match, the first service matching
// the provided type or interface will be invoked.
func invokeByGenericType[T any](i Injector) (T, error) {
	injector := getInjectorOrDefault(i)
	serviceAliasName := inferServiceName[T]()

	var invokerChain []string

	vScope, isVirtualScope := injector.(*virtualScope)
	if isVirtualScope {
		invokerChain = vScope.invokerChain
	}

	var serviceInstance any
	var serviceScope *Scope
	var serviceRealName string
	var ok bool

	injector.serviceForEachRec(func(name string, scope *Scope, s any) bool {
		if serviceIsAssignable[T](s) {
			// we need an empty instance here, because we don't want to instantiate the service when not needed

			serviceInstance = s
			serviceScope = scope
			serviceRealName = s.(serviceGetName).getName()
			ok = true

			return false
		}

		return true
	})

	if !ok {
		return empty[T](), serviceNotFound(injector, serviceAliasName)
	}

	if isVirtualScope {
		if err := vScope.detectCircularDependency(serviceRealName); err != nil {
			return empty[T](), err
		}
	}

	instance, err := serviceInstance.(serviceGetInstanceAny).getInstanceAny(
		&virtualScope{
			invokerChain: append(invokerChain, serviceRealName),
			self:         serviceScope,
		},
	)
	if err != nil {
		return empty[T](), err
	}

	if isVirtualScope {
		// We chose to register the real service name in invocation chain, because using the
		// interface name would break cycle detection.

		vScope.addDependency(injector, serviceRealName, serviceScope)
	}

	serviceScope.onServiceInvoke(serviceRealName)

	injector.RootScope().opts.Logf("DI: service %s invoked", serviceAliasName)

	return instance.(T), nil
}

// serviceNotFound returns an error indicating that the specified service was not found.
func serviceNotFound(injector Injector, name string) error {
	services := injector.ListProvidedServices()

	if len(services) == 0 {
		return fmt.Errorf("%w `%s`, no service available", ErrServiceNotFound, name)
	}

	serviceNames := getServiceNames(services)
	sortedServiceNames := sortServiceNames(serviceNames)

	return fmt.Errorf("%w `%s`, available services: %s", ErrServiceNotFound, name, strings.Join(sortedServiceNames, ", "))
}

// getServiceNames formats a list of EdgeService names.
func getServiceNames(services []EdgeService) []string {
	return mAp(services, func(edge EdgeService, _ int) string {
		return fmt.Sprintf("`%s`", edge.Service)
	})
}

// sortServiceNames sorts a list of service names.
func sortServiceNames(names []string) []string {
	sort.Strings(names)

	return names
}
