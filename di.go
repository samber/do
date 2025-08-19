package do

import (
	"fmt"
	"reflect"
)

// NameOf returns the name of the service in the DI container.
// This is highly discouraged to use this function, as your code
// should not declare any dependency explicitly.
//
// The function uses type inference to determine the service name
// based on the generic type parameter T.
func NameOf[T any]() string {
	return inferServiceName[T]()
}

// Provide registers a service in the DI container, using type inference.
// The service will be lazily instantiated when first requested.
//
// Example:
//
//	do.Provide(injector, func(i do.Injector) (*MyService, error) {
//	    return &MyService{...}, nil
//	})
func Provide[T any](i Injector, provider Provider[T]) {
	name := inferServiceName[T]()
	ProvideNamed(i, name, provider)
}

// ProvideNamed registers a named service in the DI container.
// This allows you to register multiple services of the same type
// with different names for disambiguation.
//
// The service will be lazily instantiated when first requested.
func ProvideNamed[T any](i Injector, name string, provider Provider[T]) {
	provide(i, name, provider, func(s string, a Provider[T]) Service[T] {
		return newServiceLazy(s, a)
	})
}

// ProvideValue registers a value in the DI container, using type inference to determine the service name.
// The value is immediately available and will not be recreated on each request.
//
// Example:
//
//	ProvideValue(injector, &MyService{})
func ProvideValue[T any](i Injector, value T) {
	name := inferServiceName[T]()
	ProvideNamedValue(i, name, value)
}

// ProvideNamedValue registers a named value in the DI container.
// This allows you to register multiple values of the same type
// with different names for disambiguation.
//
// The value is immediately available and will not be recreated on each request.
func ProvideNamedValue[T any](i Injector, name string, value T) {
	provide(i, name, value, func(s string, a T) Service[T] {
		return newServiceEager(s, a)
	})
}

// ProvideTransient registers a factory in the DI container, using type inference to determine the service name.
// The service will be recreated each time it is requested, providing a fresh instance.
//
// Example:
//
//	do.ProvideTransient(injector, func(i do.Injector) (*MyService, error) {
//	    return &MyService{...}, nil
//	})
func ProvideTransient[T any](i Injector, provider Provider[T]) {
	name := inferServiceName[T]()
	ProvideNamedTransient(i, name, provider)
}

// ProvideNamedTransient registers a named factory in the DI container.
// This allows you to register multiple transient services of the same type
// with different names for disambiguation.
//
// The service will be recreated each time it is requested, providing a fresh instance.
func ProvideNamedTransient[T any](i Injector, name string, provider Provider[T]) {
	provide(i, name, provider, func(s string, a Provider[T]) Service[T] {
		return newServiceTransient(s, a)
	})
}

// provide is an internal helper function that handles the common logic
// for registering services in the DI container. It ensures that:
// - The injector is properly initialized
// - No duplicate service names are registered
// - The service is properly created and stored
// - Logging is performed for successful registration
func provide[T any, A any](i Injector, name string, valueOrProvider A, serviceCtor func(string, A) Service[T]) {
	_i := getInjectorOrDefault(i)
	if _i.serviceExist(name) {
		panic(fmt.Errorf("DI: service `%s` has already been declared", name))
	}

	service := serviceCtor(name, valueOrProvider)
	_i.serviceSet(name, service)

	_i.RootScope().opts.Logf("DI: service %s injected", name)
}

// Override replaces the service in the DI container, using type inference to determine the service name.
// Warning: this will not unload/shutdown the previously invoked service.
//
// This function is useful for testing or when you need to replace a service
// that has already been registered. However, be cautious as it may lead to
// resource leaks if the original service was already instantiated.
func Override[T any](i Injector, provider Provider[T]) {
	name := inferServiceName[T]()
	OverrideNamed(i, name, provider)
}

// OverrideNamed replaces the named service in the DI container.
// Warning: this will not unload/shutdown the previously invoked service.
//
// This function allows you to replace a specific named service that has
// already been registered. Use with caution to avoid resource leaks.
func OverrideNamed[T any](i Injector, name string, provider Provider[T]) {
	override(i, name, provider, func(s string, a Provider[T]) Service[T] {
		return newServiceLazy(s, a)
	})
}

// OverrideValue replaces the value in the DI container, using type inference to determine the service name.
// Warning: this will not unload/shutdown the previously invoked service.
//
// This function replaces an existing value service with a new one.
// The old value will not be properly cleaned up if it was already instantiated.
func OverrideValue[T any](i Injector, value T) {
	name := inferServiceName[T]()
	OverrideNamedValue(i, name, value)
}

// OverrideNamedValue replaces the named value in the DI container.
// Warning: this will not unload/shutdown the previously invoked service.
//
// This function allows you to replace a specific named value service.
// Use with caution to avoid resource leaks.
func OverrideNamedValue[T any](i Injector, name string, value T) {
	override(i, name, value, func(s string, a T) Service[T] {
		return newServiceEager(s, a)
	})
}

// OverrideTransient replaces the factory in the DI container, using type inference to determine the service name.
// Warning: this will not unload/shutdown the previously invoked service.
//
// This function replaces an existing transient service factory with a new one.
// Since transient services are recreated on each request, this is generally safer
// than overriding lazy or eager services.
func OverrideTransient[T any](i Injector, provider Provider[T]) {
	name := inferServiceName[T]()
	OverrideNamedTransient(i, name, provider)
}

// OverrideNamedTransient replaces the named factory in the DI container.
// Warning: this will not unload/shutdown the previously invoked service.
//
// This function allows you to replace a specific named transient service factory.
// Since transient services are recreated on each request, this is generally safer
// than overriding lazy or eager services.
func OverrideNamedTransient[T any](i Injector, name string, provider Provider[T]) {
	override(i, name, provider, func(s string, a Provider[T]) Service[T] {
		return newServiceTransient(s, a)
	})
}

// override is an internal helper function that handles the common logic
// for overriding services in the DI container. Unlike provide, it allows
// replacing existing services without throwing an error.
func override[T any, A any](i Injector, name string, valueOrProvider A, serviceCtor func(string, A) Service[T]) {
	_i := getInjectorOrDefault(i)

	// Note: We don't check if the service exists here, allowing override
	service := serviceCtor(name, valueOrProvider)
	_i.serviceSet(name, service) // @TODO: should we unload/shutdown the previous service ?

	_i.RootScope().opts.Logf("DI: service %s overridden", name)
}

// Invoke retrieves and instantiates a service from the DI container using type inference.
// The service will be created if it hasn't been instantiated yet (for lazy services).
//
// Example:
//
//	service, err := do.Invoke[*MyService](injector)
func Invoke[T any](i Injector) (T, error) {
	name := inferServiceName[T]()
	return invokeByName[T](i, name)
}

// InvokeNamed retrieves and instantiates a named service from the DI container.
// This allows you to retrieve specific named services when multiple services
// of the same type are registered.
//
// Example:
//
//	service, err := do.InvokeNamed[*MyService](injector, "my-service")
func InvokeNamed[T any](i Injector, name string) (T, error) {
	if typeIsAssignable[T, any]() {
		v, err := invokeAnyByName(i, name)
		t, _ := v.(T) // just skip if v == nil
		return t, err
	}

	return invokeByName[T](i, name)
}

// MustInvoke retrieves and instantiates a service from the DI container using type inference.
// If the service cannot be retrieved or instantiated, it panics.
//
// This function is useful when you're certain the service exists and want
// to avoid error handling in your code.
//
// Example:
//
//	service := do.MustInvoke[*MyService](injector)
func MustInvoke[T any](i Injector) T {
	return must1(Invoke[T](i))
}

// MustInvokeNamed retrieves and instantiates a named service from the DI container.
// If the service cannot be retrieved or instantiated, it panics.
//
// This function is useful when you're certain the named service exists and want
// to avoid error handling in your code.
//
// Example:
//
//	service := do.MustInvokeNamed[*MyService](injector, "my-service")
func MustInvokeNamed[T any](i Injector, name string) T {
	return must1(InvokeNamed[T](i, name))
}

// InvokeStruct invokes services located in struct properties.
// The struct fields must be tagged with `do:""` or `do:"name"`, where `name` is the service name in the DI container.
// If the service is not found in the DI container, an error is returned.
// If the service is found but not assignable to the struct field, an error is returned.
func InvokeStruct[T any](i Injector) (T, error) {
	structName := inferServiceName[T]()
	output := deepEmpty[T]() // if the struct is hidden behind a pointer, we need to init the struct value deep enough
	value := reflect.ValueOf(&output)

	for value.Elem().Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// Check if the empty value is a struct (before passing a pointer to reflect.ValueOf).
	// It will be checked in invokeByTags, but the error message is different.
	if value.Kind() != reflect.Pointer || value.Elem().Kind() != reflect.Struct {
		return empty[T](), fmt.Errorf("DI: must be a struct or a pointer to a struct, but got `%s`", structName)
	}

	err := invokeByTags(i, structName, value)
	if err != nil {
		return empty[T](), err
	}

	return output, nil
}

// InvokeStruct invokes services located in struct properties.
// The struct fields must be tagged with `do:""` or `do:"name"`, where `name` is the service name in the DI container.
// If the service is not found in the DI container, an error is returned.
// If the service is found but not assignable to the struct field, an error is returned.
// It panics on error.
func MustInvokeStruct[T any](i Injector) T {
	return must1(InvokeStruct[T](i))
}
