package do

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unsafe"

	typetostring "github.com/samber/go-type-to-string"
)

// invokeAnyByName retrieves and instantiates a service by name, returning it as interface{}.
// This function handles circular dependency detection, service resolution, and invocation hooks.
//
// Parameters:
//   - i: The injector to search for the service
//   - name: The name of the service to invoke
//
// Returns the service instance as interface{} and any error that occurred during invocation.
//
// This function is used internally by the DI container for service resolution and
// supports virtual scopes for dependency tracking and circular dependency detection.
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

	service, ok := serviceAny.(serviceWrapperAny)
	if !ok {
		return nil, serviceNotFound(injector, ErrServiceNotFound, invokerChain)
	}

	injector.RootScope().opts.onBeforeInvocation(serviceScope, name)
	instance, err := service.getInstanceAny(newVirtualScope(serviceScope, invokerChain))
	injector.RootScope().opts.onAfterInvocation(serviceScope, name, err)
	if err != nil {
		return nil, err
	}

	serviceScope.onServiceInvoke(name)

	injector.RootScope().opts.Logf("DI: service %s invoked", name)

	return instance, nil
}

// invokeByName retrieves and instantiates a service by name with type safety.
// This function handles circular dependency detection, service resolution, type checking,
// and invocation hooks.
//
// Parameters:
//   - i: The injector to search for the service
//   - name: The name of the service to invoke
//
// Returns the service instance with the correct type and any error that occurred during invocation.
//
// This function is used internally by the DI container for type-safe service resolution.
// It ensures that the returned service matches the expected type T.
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

	service, ok := serviceAny.(serviceWrapper[T])
	if !ok {
		return empty[T](), serviceTypeMismatch(inferServiceName[T](), serviceAny.(serviceWrapperAny).getTypeName()) //nolint:errcheck,forcetypeassert
	}

	injector.RootScope().opts.onBeforeInvocation(serviceScope, name)
	instance, err := service.getInstance(newVirtualScope(serviceScope, invokerChain))
	injector.RootScope().opts.onAfterInvocation(serviceScope, name, err)

	if err != nil {
		return empty[T](), err
	}

	serviceScope.onServiceInvoke(name)

	injector.RootScope().opts.Logf("DI: service %s invoked", name)

	return instance, nil
}

// invokeByGenericType looks for a service by its type and invokes the first matching service.
// When multiple services match the provided type or interface, the first service found
// will be invoked. This function is useful for interface-based dependency injection.
//
// Parameters:
//   - i: The injector to search for the service
//
// Returns the service instance with the correct type and any error that occurred during invocation.
//
// @TODO: Selection is nondeterministic when multiple services satisfy T; consider deterministic ordering
// or explicit disambiguation to avoid surprising picks.
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
		if serviceCanCastToGeneric[T](s) {
			serviceInstance = s
			serviceScope = scope
			serviceRealName = s.(serviceWrapperGetName).getName() //nolint:errcheck,forcetypeassert
			ok = true

			// Stop or not stop, that's the question -> https://github.com/samber/do/issues/114
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
	instance, err := serviceInstance.(serviceWrapperGetInstanceAny).getInstanceAny( //nolint:errcheck,forcetypeassert
		newVirtualScope(serviceScope, append(invokerChain, serviceRealName)),
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

	return instance.(T), nil //nolint:errcheck,forcetypeassert
}

// invokeAsAllByGenericType finds and invokes all services matching type T.
// This function performs a two-phase operation:
// 1. Discovery phase: Find all services that can be cast to T
// 2. Invocation phase: Invoke each matching service and collect results
//
// The function returns services in deterministic order (sorted by service name)
// and handles partial failures by returning successfully invoked services
// along with detailed error information.
func invokeAsAllByGenericType[T any](i Injector) ([]T, error) {
	injector := getInjectorOrDefault(i)
	results := make([]T, 0)

	var invokerChain []string
	vScope, isVirtualScope := injector.(*virtualScope)
	if isVirtualScope {
		invokerChain = vScope.invokerChain
	}

	// Discovery phase: collect all matching services
	type serviceMatch struct {
		name     string
		instance any
		scope    *Scope
	}
	var matches []serviceMatch

	injector.serviceForEachRec(func(name string, scope *Scope, s any) bool {
		if serviceCanCastToGeneric[T](s) {
			serviceWrapper, ok := s.(serviceWrapperGetName) //nolint:forcetypeassert
			if !ok {
				return true
			}
			matches = append(matches, serviceMatch{
				name:     serviceWrapper.getName(),
				instance: s,
				scope:    scope,
			})
		}
		return true
	})

	if len(matches) == 0 {
		// For InvokeAsAll, returning no services is not an error
		return nil, nil
	}

	// Sort matches for deterministic ordering
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].name < matches[j].name
	})

	// Invocation phase: invoke each matching service
	for _, match := range matches {
		if isVirtualScope {
			if err := vScope.detectCircularDependency(match.name); err != nil {
				// Return partial results with error on circular dependency
				return results, fmt.Errorf("circular dependency detected involving service %s: %w", match.name, err)
			}
		}

		// Use the interface name for hooks to maintain consistency with InvokeAs
		interfaceName := inferServiceName[T]()
		injector.RootScope().opts.onBeforeInvocation(match.scope, interfaceName)

		serviceInstanceWrapper, ok := match.instance.(serviceWrapperGetInstanceAny) //nolint:forcetypeassert
		if !ok {
			return results, fmt.Errorf("failed to invoke service %s: service does not support invocation", match.name)
		}
		instance, err := serviceInstanceWrapper.getInstanceAny( //nolint:errcheck
			newVirtualScope(match.scope, append(invokerChain, match.name)))

		injector.RootScope().opts.onAfterInvocation(match.scope, interfaceName, err)

		if err != nil {
			// Return partial results with specific error details
			return results, fmt.Errorf("failed to invoke service %s: %w", match.name, err)
		}

		if isVirtualScope {
			vScope.addDependency(injector, match.name, match.scope)
		}

		match.scope.onServiceInvoke(match.name)
		injector.RootScope().opts.Logf("DI: service %s invoked", match.name)

		results = append(results, instance.(T)) //nolint:errcheck,forcetypeassert
	}

	return results, nil
}

// invokeByTag injects services into struct fields based on struct tags.
// This function supports automatic dependency injection into struct fields
// using the `do` tag or a custom tag key specified in the injector options.
// If `implicitAliasing` is true and a service is not found by tag, the injector
// will fall back to searching for a service by the field's generic type, like `do.InvokeAs[T]`.
//
// Parameters:
//   - i: The injector to search for services
//   - structName: The name of the struct for error reporting
//   - structValue: A reflect.Value pointing to the struct to inject into
//   - implicitAliasing: Whether to fall back to generic type if service is not found by name
//
// Returns an error if injection fails for any reason.
//
// The function does not manipulate virtual scope because it is done by invokeAnyByName or invokeByGenericType.
//
// @TODO: When implicitAliasing is enabled and the tag name is empty, fallback by type may select an arbitrary
// matching service depending on iteration order; consider stable ordering or explicit disambiguation.
func invokeByTags(i Injector, structName string, structValue reflect.Value, implicitAliasing bool) error { //nolint:gocyclo
	injector := getInjectorOrDefault(i)

	// Ensure that servicePtr is a pointer to a struct
	if structValue.Kind() != reflect.Ptr || structValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("DI: must be a pointer to a struct")
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

		// Keep track if tag was provided without an explicit name (eg: `do:""`)
		wasTagNameEmpty := serviceName == ""

		if !fieldValue.CanAddr() {
			return fmt.Errorf("DI: field is not addressable `%s.%s`", structName, field.Name)
		}

		if !fieldValue.CanSet() {
			// When a field is not exported, we override it.
			// See https://stackoverflow.com/questions/42664837/how-to-access-unexported-struct-fields/43918797#43918797
			// bearer:disable go_gosec_unsafe_unsafe
			fieldValue = reflect.NewAt(fieldValue.Type(), unsafe.Pointer(fieldValue.UnsafeAddr())).Elem()
		}

		if serviceName == "" {
			serviceName = typetostring.GetReflectValueType(fieldValue)
		}

		dependency, err := invokeAnyByName(injector, serviceName)
		// @TODO: This fallback may pick an arbitrary matching service; selection order is not stable.
		if err != nil && implicitAliasing && wasTagNameEmpty && errors.Is(err, ErrServiceNotFound) {
			// Fallback: try to resolve by generic type of the field
			toType := fieldValue.Type()

			var resolvedName string
			var found bool
			injector.serviceForEachRec(func(name string, _ *Scope, s any) bool {
				if serviceCanCastToType(s, toType) {
					resolvedName = s.(serviceWrapperGetName).getName() //nolint:errcheck,forcetypeassert
					found = true

					// Stop or not stop, that's the question -> https://github.com/samber/do/issues/114
					return false
				}

				return true
			})

			if found {
				dependency, err = invokeAnyByName(injector, resolvedName)
			}
		}
		if err != nil {
			return err
		}

		dependencyValue := reflect.ValueOf(dependency)

		// Should be checked before invocation, because we just built something that is not assignable to the field.
		if !dependencyValue.Type().AssignableTo(fieldValue.Type()) {
			return fmt.Errorf("DI: `%s` is not assignable to field `%s.%s`", serviceName, structName, field.Name)
		}

		// Should not panic, since we checked CanAddr() and CanSet() earlier.
		fieldValue.Set(dependencyValue)
	}

	return nil
}

// serviceNotFound returns a detailed error indicating that the specified service was not found.
// This function provides helpful error messages that include available services and
// the invocation chain for debugging purposes.
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
		return fmt.Errorf(
			"%w `%s`, available services: %s, path: %s",
			err,
			name,
			strings.Join(sortedServiceNames, ", "),
			humanReadableInvokerChain(chain),
		)
	}
	return fmt.Errorf("%w `%s`, available services: %s", err, name, strings.Join(sortedServiceNames, ", "))
}

// serviceTypeMismatch returns an error indicating that the specified service was found,
// but its type does not match the expected type. This typically occurs when a service
// is registered with one type but invoked with a different type.
func serviceTypeMismatch(invoking string, registered string) error {
	return fmt.Errorf("DI: service found, but type mismatch: invoking `%s` but registered `%s`", invoking, registered)
}

// getServiceNames formats a list of ServiceDescription names for error reporting.
// This function converts ServiceDescription objects to formatted string names
// that can be displayed in error messages.
func getServiceNames(services []ServiceDescription) []string {
	return mAp(services, func(desc ServiceDescription, _ int) string {
		return fmt.Sprintf("`%s`", desc.Service)
	})
}

// sortServiceNames sorts a list of service names alphabetically.
// This function ensures consistent ordering of service names in error messages
// and other output, making them easier to read and compare.
func sortServiceNames(names []string) []string {
	sort.Strings(names)
	return names
}

// humanReadableInvokerChain formats an invocation chain into a human-readable string.
// This function converts a slice of service names into a formatted string
// that shows the dependency chain for debugging purposes.
//
// This is useful for debugging circular dependencies and understanding
// the service resolution path that led to an error.
func humanReadableInvokerChain(invokerChain []string) string {
	invokerChain = mAp(invokerChain, func(item string, _ int) string {
		return fmt.Sprintf("`%s`", item)
	})
	return strings.Join(invokerChain, " -> ")
}

func handleProviderPanic[T any](provider Provider[T], i Injector) (svc T, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("DI: %v", r)
			}
		}
	}()

	_svc, _err := provider(i)

	// do not return svc when err != nil
	if _err != nil {
		err = _err
	} else {
		svc = _svc
	}

	return svc, err
}
