package do

import (
	"fmt"
	"reflect"
)

// NameOf returns the name of the service in the DI container.
// This is higly discouraged to use this function, as your code
// should not declare any dependency explicitly.
func NameOf[T any]() string {
	return inferServiceName[T]()
}

// Provide registers a service in the DI container, using type inference.
func Provide[T any](i Injector, provider Provider[T]) {
	name := inferServiceName[T]()
	ProvideNamed[T](i, name, provider)
}

// ProvideNamed registers a named service in the DI container.
func ProvideNamed[T any](i Injector, name string, provider Provider[T]) {
	provide(i, name, provider, func(s string, a Provider[T]) Service[T] {
		return newServiceLazy[T](s, a)
	})
}

// ProvideValue registers a value in the DI container, using type inference to determine the service name.
func ProvideValue[T any](i Injector, value T) {
	name := inferServiceName[T]()
	ProvideNamedValue[T](i, name, value)
}

// ProvideNamedValue registers a named value in the DI container.
func ProvideNamedValue[T any](i Injector, name string, value T) {
	provide(i, name, value, func(s string, a T) Service[T] {
		return newServiceEager[T](s, a)
	})
}

// ProvideTransient registers a factory in the DI container, using type inference to determine the service name.
func ProvideTransient[T any](i Injector, provider Provider[T]) {
	name := inferServiceName[T]()
	ProvideNamedTransient[T](i, name, provider)
}

// ProvideNamedTransient registers a named factory in the DI container.
func ProvideNamedTransient[T any](i Injector, name string, provider Provider[T]) {
	provide(i, name, provider, func(s string, a Provider[T]) Service[T] {
		return newServiceTransient[T](s, a)
	})
}

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
func Override[T any](i Injector, provider Provider[T]) {
	name := inferServiceName[T]()
	OverrideNamed[T](i, name, provider)
}

// OverrideNamed replaces the named service in the DI container.
func OverrideNamed[T any](i Injector, name string, provider Provider[T]) {
	override(i, name, provider, func(s string, a Provider[T]) Service[T] {
		return newServiceLazy[T](s, a)
	})
}

// OverrideValue replaces the value in the DI container, using type inference to determine the service name.
func OverrideValue[T any](i Injector, value T) {
	name := inferServiceName[T]()
	OverrideNamedValue[T](i, name, value)
}

// OverrideNamedValue replaces the named value in the DI container.
func OverrideNamedValue[T any](i Injector, name string, value T) {
	override(i, name, value, func(s string, a T) Service[T] {
		return newServiceEager[T](s, a)
	})
}

// OverrideTransient replaces the factory in the DI container, using type inference to determine the service name.
func OverrideTransient[T any](i Injector, provider Provider[T]) {
	name := inferServiceName[T]()
	OverrideNamed[T](i, name, provider)
}

// OverrideNamedTransient replaces the named factory in the DI container.
func OverrideNamedTransient[T any](i Injector, name string, provider Provider[T]) {
	override(i, name, provider, func(s string, a Provider[T]) Service[T] {
		return newServiceTransient[T](s, a)
	})
}

func override[T any, A any](i Injector, name string, valueOrProvider A, serviceCtor func(string, A) Service[T]) {
	_i := getInjectorOrDefault(i)

	service := serviceCtor(name, valueOrProvider)
	_i.serviceSet(name, service) // @TODO: should we unload/shutdown ?

	_i.RootScope().opts.Logf("DI: service %s overridden", name)
}

// Invoke invokes a service in the DI container, using type inference to determine the service name.
func Invoke[T any](i Injector) (T, error) {
	name := inferServiceName[T]()
	return InvokeNamed[T](i, name)
}

// MustInvoke invokes a service in the DI container, using type inference to determine the service name. It panics on error.
func MustInvoke[T any](i Injector) T {
	return must1(Invoke[T](i))
}

// InvokeNamed invokes a named service in the DI container.
func InvokeNamed[T any](i Injector, name string) (T, error) {
	return invokeByName[T](i, name)
}

// MustInvokeNamed invokes a named service in the DI container. It panics on error.
func MustInvokeNamed[T any](i Injector, name string) T {
	return must1(InvokeNamed[T](i, name))
}

// InvokeStruct invokes services located in struct properties.
// The struct fields must be tagged with `do:""` or `do:"name"`, where `name` is the service name in the DI container.
// If the service is not found in the DI container, an error is returned.
// If the service is found but not assignable to the struct field, an error is returned.
func InvokeStruct[T any](i Injector) (*T, error) {
	output := empty[T]()
	value := reflect.ValueOf(&output)

	// Check if the empty value is a struct (before passing a pointer to reflect.ValueOf).
	// It will be checked in invokeByTags, but the error message is different.
	if value.Kind() != reflect.Ptr || value.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("DI: not a struct")
	}

	structName := inferServiceName[T]()

	err := invokeByTags(i, structName, value)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

// InvokeStruct invokes services located in struct properties.
// The struct fields must be tagged with `do:""` or `do:"name"`, where `name` is the service name in the DI container.
// If the service is not found in the DI container, an error is returned.
// If the service is found but not assignable to the struct field, an error is returned.
// It panics on error.
func MustInvokeStruct[T any](i Injector) *T {
	return must1(InvokeStruct[T](i))
}
