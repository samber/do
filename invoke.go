package do

import (
	"fmt"
	"sort"
	"strings"
)

// invokeByName look for a service by its name.
func invokeByName[T any](i Injector, name string) (T, error) {
	_i := getInjectorOrDefault(i)

	invokerName := ""
	invokerChain := []string{}
	if vs, ok := _i.(*virtualScope); ok {
		if len(vs.invokerChain) > 0 {
			invokerName = vs.invokerChain[len(vs.invokerChain)-1]
		}

		invokerChain = vs.invokerChain

		if contains(invokerChain, name) {
			return empty[T](), fmt.Errorf("DI: circular dependency detected: %s -> %s", strings.Join(invokerChain, " -> "), name)
		}
	}

	serviceAny, serviceScope, ok := _i.serviceGetRec(name)
	if !ok {
		return empty[T](), serviceNotFound(_i, name)
	}

	service, ok := serviceAny.(Service[T])
	if !ok {
		return empty[T](), serviceNotFound(_i, name)
	}

	instance, err := service.getInstance(&virtualScope{invokerChain: append(invokerChain, name), self: serviceScope})
	if err != nil {
		return empty[T](), err
	}

	if _, ok := _i.(*virtualScope); ok {
		_i.RootScope().dag.addDependency(_i.ID(), _i.Name(), invokerName, serviceScope.ID(), serviceScope.Name(), name)
	}

	serviceScope.onServiceInvoke(name)
	_i.RootScope().opts.Logf("DI: service %s invoked", name)

	return instance, nil
}

// invokeByGenericType look for a service by its type.
// When many services match, the first service matching
// the provided type or interface will be invoked.
func invokeByGenericType[T any](i Injector) (T, error) {
	_i := getInjectorOrDefault(i)
	serviceAliasName := inferServiceName[T]()

	invokerName := ""
	invokerChain := []string{}
	if vs, ok := _i.(*virtualScope); ok {
		if len(vs.invokerChain) > 0 {
			invokerName = vs.invokerChain[len(vs.invokerChain)-1]
		}

		invokerChain = vs.invokerChain

		// compared to invoke(), we check for circular dependencies lazily
	}

	var service any
	var serviceScope *Scope
	var ok bool = false
	_i.serviceForEachRec(func(name string, scope *Scope, s any) bool {
		if serviceIsAssignable[T](s) {
			// we need an empty instance here, because we don't want to instantiate the service when not needed
			service = s
			serviceScope = scope
			ok = true
			return false
		}

		return true
	})

	if !ok {
		return empty[T](), serviceNotFound(_i, serviceAliasName)
	}

	// We chose to register the real service name in invocation chain, because using the
	// interface name would break cycle detection.
	serviceRealName := service.(serviceGetName).getName()

	// Check for circular dependencies.
	for i := range invokerChain {
		if invokerChain[i] == serviceRealName {
			return empty[T](), fmt.Errorf("DI: circular dependency detected: %s -> %s", strings.Join(invokerChain, " -> "), serviceRealName)
		}
	}

	instance, err := service.(serviceGetInstanceAny).getInstanceAny(&virtualScope{invokerChain: append(invokerChain, serviceRealName), self: serviceScope})
	if err != nil {
		return empty[T](), err
	}

	if _, ok := _i.(*virtualScope); ok {
		// Should we use the alias name or the real name?
		_i.RootScope().dag.addDependency(_i.ID(), _i.Name(), invokerName, serviceScope.ID(), serviceScope.Name(), serviceRealName)
	}

	serviceScope.onServiceInvoke(serviceRealName)                        // from the service POV, we use the real name injected into the scope
	_i.RootScope().opts.Logf("DI: service %s invoked", serviceAliasName) // from the invoker POV, we use the alias name

	return instance.(T), nil
}

func serviceNotFound(injector Injector, name string) error {
	// @TODO: use the Keys+Map functions from `golang.org/x/exp/maps` as
	// soon as it is released in stdlib.
	services := injector.ListProvidedServices()
	servicesNames := mAp(services, func(edge EdgeService, _ int) string {
		return fmt.Sprintf("`%s`", edge.Service)
	})

	// cool for unit tests
	sorter := sort.StringSlice(servicesNames)
	sorter.Sort()
	servicesNames = []string(sorter)

	if len(servicesNames) == 0 {
		return fmt.Errorf("DI: could not find service `%s`, no service available", name)
	}

	return fmt.Errorf("DI: could not find service `%s`, available services: %s", name, strings.Join(servicesNames, ", "))
}
