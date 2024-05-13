package do

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unsafe"

	typetostring "github.com/samber/go-type-to-string"
)

// invokeAnyByName looks for a service by its tag.
func invokeAnyByName(i Injector, name string) (any, error) {
	var invokerChain []string

	injector := getInjectorOrDefault(i)

	vScope, isVirtualScope := injector.(*virtualScope)
	if isVirtualScope {
		invokerChain = vScope.invokerChain

		if err := vScope.detectCircularDependency(name); err != nil {
			return nil, err
		}
	}

	invokerChain = append(invokerChain, name)

	serviceAny, serviceScope, found := injector.serviceGetRec(name)
	if !found {
		return nil, serviceNotFound(injector, ErrServiceNotFound, invokerChain)
	}

	if isVirtualScope {
		vScope.addDependency(injector, name, serviceScope)
	}

	service, ok := serviceAny.(ServiceAny)
	if !ok {
		return nil, serviceNotFound(injector, ErrServiceNotFound, invokerChain)
	}

	injector.RootScope().opts.onBeforeInvocation(serviceScope, name)
	instance, err := service.getInstanceAny(&virtualScope{invokerChain: invokerChain, self: serviceScope})
	injector.RootScope().opts.onAfterInvocation(serviceScope, name, err)
	if err != nil {
		return nil, err
	}

	serviceScope.onServiceInvoke(name)

	injector.RootScope().opts.Logf("DI: service %s invoked", name)

	return instance, nil
}

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

	invokerChain = append(invokerChain, name)

	serviceAny, serviceScope, found := injector.serviceGetRec(name)
	if !found {
		return empty[T](), serviceNotFound(injector, ErrServiceNotFound, invokerChain)
	}

	if isVirtualScope {
		vScope.addDependency(injector, name, serviceScope)
	}

	service, ok := serviceAny.(Service[T])
	if !ok {
		return empty[T](), serviceTypeMismatch(inferServiceName[T](), serviceAny.(ServiceAny).getTypeName())
	}

	injector.RootScope().opts.onBeforeInvocation(serviceScope, name)
	instance, err := service.getInstance(&virtualScope{invokerChain: invokerChain, self: serviceScope})
	injector.RootScope().opts.onAfterInvocation(serviceScope, name, err)

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
		return empty[T](), serviceNotFound(injector, ErrServiceNotMatch, append(invokerChain, serviceAliasName))
	}

	if isVirtualScope {
		if err := vScope.detectCircularDependency(serviceRealName); err != nil {
			return empty[T](), err
		}
	}

	injector.RootScope().opts.onBeforeInvocation(serviceScope, serviceAliasName)
	instance, err := serviceInstance.(serviceGetInstanceAny).getInstanceAny(
		&virtualScope{
			invokerChain: append(invokerChain, serviceRealName),
			self:         serviceScope,
		},
	)
	injector.RootScope().opts.onAfterInvocation(serviceScope, serviceAliasName, err)

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

// invokeByTag looks for a service by its tag.
func invokeByTags(i Injector, structName string, structValue reflect.Value) error {
	injector := getInjectorOrDefault(i)

	// Ensure that servicePtr is a pointer to a struct
	if structValue.Kind() != reflect.Ptr || structValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("DI: not a pointer")
	}

	structValue = structValue.Elem()

	// Iterate through the fields of the struct
	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Type().Field(i)
		fieldValue := structValue.Field(i)

		serviceName, ok := field.Tag.Lookup(coalesce(injector.RootScope().opts.StructTagKey, DefaultStructTagKey))
		if !ok {
			continue
		}

		if !fieldValue.CanAddr() {
			return fmt.Errorf("DI: field is not addressable `%s.%s`", structName, field.Name)
		}

		if !fieldValue.CanSet() {
			// When a field is not exported, we override it.
			// See https://stackoverflow.com/questions/42664837/how-to-access-unexported-struct-fields/43918797#43918797
			fieldValue = reflect.NewAt(fieldValue.Type(), unsafe.Pointer(fieldValue.UnsafeAddr())).Elem()
		}

		if serviceName == "" {
			serviceName = typetostring.GetReflectValueType(fieldValue)
		}

		dependency, err := invokeAnyByName(injector, serviceName)
		if err != nil {
			return err
		}

		dependencyValue := reflect.ValueOf(dependency)

		// Should be check before invocation, because we just built something that is not assignable to the field.
		if !fieldValue.Type().AssignableTo(dependencyValue.Type()) {
			return fmt.Errorf("DI: field `%s.%s` is not assignable to service %s", structName, field.Name, serviceName)
		}

		// Should not panic, since we checked CanAddr() and CanSet() earlier.
		fieldValue.Set(dependencyValue)
	}

	return nil
}

// serviceNotFound returns an error indicating that the specified service was not found.
func serviceNotFound(injector Injector, err error, chain []string) error {
	name := chain[len(chain)-1]
	services := injector.ListProvidedServices()

	if len(services) == 0 {
		if len(chain) > 1 {
			return fmt.Errorf("%w `%s`, no service available, path: %s", err, name, humanReadableInvokerChain(chain))
		}
		return fmt.Errorf("%w `%s`, no service available", err, name)
	}

	serviceNames := getServiceNames(services)
	sortedServiceNames := sortServiceNames(serviceNames)

	if len(chain) > 1 {
		return fmt.Errorf("%w `%s`, available services: %s, path: %s", err, name, strings.Join(sortedServiceNames, ", "), humanReadableInvokerChain(chain))
	}
	return fmt.Errorf("%w `%s`, available services: %s", err, name, strings.Join(sortedServiceNames, ", "))
}

// serviceTypeMismatch returns an error indicating that the specified service was found, but its type does not match the expected type.
func serviceTypeMismatch(invoking string, registered string) error {
	return fmt.Errorf("DI: service found, but type mismatch: invoking `%s` but registered `%s`", invoking, registered)
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

func humanReadableInvokerChain(invokerChain []string) string {
	invokerChain = mAp(invokerChain, func(item string, _ int) string {
		return fmt.Sprintf("`%s`", item)
	})
	return strings.Join(invokerChain, " -> ")
}
