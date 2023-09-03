package do

import (
	"fmt"
)

func Provide[T any](i Injector, provider Provider[T]) {
	name := inferServiceName[T]()
	ProvideNamed[T](i, name, provider)
}

func ProvideNamed[T any](i Injector, name string, provider Provider[T]) {
	provide(i, name, provider, newServiceLazy[T])
}

func ProvideValue[T any](i Injector, value T) {
	name := inferServiceName[T]()
	ProvideNamedValue[T](i, name, value)
}

func ProvideNamedValue[T any](i Injector, name string, value T) {
	provide(i, name, value, newServiceEager[T])
}

func ProvideTransiant[T any](i Injector, provider Provider[T]) {
	name := inferServiceName[T]()
	ProvideNamedTransiant[T](i, name, provider)
}

func ProvideNamedTransiant[T any](i Injector, name string, provider Provider[T]) {
	provide(i, name, provider, newServiceTransiant[T])
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

func Override[T any](i Injector, provider Provider[T]) {
	name := inferServiceName[T]()
	OverrideNamed[T](i, name, provider)
}

func OverrideNamed[T any](i Injector, name string, provider Provider[T]) {
	override(i, name, provider, newServiceLazy[T])
}

func OverrideValue[T any](i Injector, value T) {
	name := inferServiceName[T]()
	OverrideNamedValue[T](i, name, value)
}

func OverrideNamedValue[T any](i Injector, name string, value T) {
	override(i, name, value, newServiceEager[T])
}

func OverrideTransiant[T any](i Injector, provider Provider[T]) {
	name := inferServiceName[T]()
	OverrideNamed[T](i, name, provider)
}

func OverrideNamedTransiant[T any](i Injector, name string, provider Provider[T]) {
	override(i, name, provider, newServiceTransiant[T])
}

func override[T any, A any](i Injector, name string, valueOrProvider A, serviceCtor func(string, A) Service[T]) {
	_i := getInjectorOrDefault(i)

	service := serviceCtor(name, valueOrProvider)
	_i.serviceSet(name, service) // @TODO: should we unload/shutdown ?

	_i.RootScope().opts.Logf("DI: service %s overridden", name)
}

func Invoke[T any](i Injector) (T, error) {
	name := inferServiceName[T]()
	return InvokeNamed[T](i, name)
}

func MustInvoke[T any](i Injector) T {
	s, err := Invoke[T](i)
	must0(err)
	return s
}

func InvokeNamed[T any](i Injector, name string) (T, error) {
	return invoke[T](i, name)
}

func MustInvokeNamed[T any](i Injector, name string) T {
	s, err := InvokeNamed[T](i, name)
	must0(err)
	return s
}
