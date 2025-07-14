package do

import (
	"fmt"
)

func Provide[T any](i *Injector, provider Provider[T]) {
	name := generateServiceNameFromInjector[T](i)

	ProvideNamed(i, name, provider)
}

func ProvideNamed[T any](i *Injector, name string, provider Provider[T]) {
	_i := getInjectorOrDefault(i)
	if _i.exists(name) {
		panic(fmt.Errorf("DI: service `%s` has already been declared", name))
	}

	service := newServiceLazy(name, provider)
	_i.set(name, service)

	_i.logf("service %s injected", name)
}

func ProvideValue[T any](i *Injector, value T) {
	name := generateServiceNameFromInjector[T](i)

	ProvideNamedValue(i, name, value)
}

func ProvideNamedValue[T any](i *Injector, name string, value T) {
	_i := getInjectorOrDefault(i)
	if _i.exists(name) {
		panic(fmt.Errorf("DI: service `%s` has already been declared", name))
	}

	service := newServiceEager(name, value)
	_i.set(name, service)

	_i.logf("service %s injected", name)
}

func Override[T any](i *Injector, provider Provider[T]) {
	name := generateServiceNameFromInjector[T](i)

	OverrideNamed(i, name, provider)
}

func OverrideNamed[T any](i *Injector, name string, provider Provider[T]) {
	_i := getInjectorOrDefault(i)

	service := newServiceLazy(name, provider)
	_i.set(name, service)

	_i.logf("service %s overridden", name)
}

func OverrideValue[T any](i *Injector, value T) {
	name := generateServiceNameFromInjector[T](i)

	OverrideNamedValue(i, name, value)
}

func OverrideNamedValue[T any](i *Injector, name string, value T) {
	_i := getInjectorOrDefault(i)

	service := newServiceEager(name, value)
	_i.set(name, service)

	_i.logf("service %s overridden", name)
}

func Invoke[T any](i *Injector) (T, error) {
	name := generateServiceNameFromInjector[T](i)
	return InvokeNamed[T](i, name)
}

func MustInvoke[T any](i *Injector) T {
	s, err := Invoke[T](i)
	must(err)
	return s
}

func InvokeNamed[T any](i *Injector, name string) (T, error) {
	return invokeImplem[T](i, name)
}

func MustInvokeNamed[T any](i *Injector, name string) T {
	s, err := InvokeNamed[T](i, name)
	must(err)
	return s
}

func invokeImplem[T any](i *Injector, name string) (T, error) {
	_i := getInjectorOrDefault(i)

	serviceAny, ok := _i.get(name)
	if !ok {
		return empty[T](), _i.serviceNotFound(name)
	}

	service, ok := serviceAny.(Service[T])
	if !ok {
		return empty[T](), _i.serviceNotFound(name)
	}

	instance, err := service.getInstance(_i)
	if err != nil {
		return empty[T](), err
	}

	_i.onServiceInvoke(name)

	_i.logf("service %s invoked", name)

	return instance, nil
}

func HealthCheck[T any](i *Injector) error {
	name := generateServiceNameFromInjector[T](i)
	return getInjectorOrDefault(i).healthcheckImplem(name)
}

func HealthCheckNamed(i *Injector, name string) error {
	return getInjectorOrDefault(i).healthcheckImplem(name)
}

func Shutdown[T any](i *Injector) error {
	name := generateServiceNameFromInjector[T](i)
	return getInjectorOrDefault(i).shutdownImplem(name)
}

func MustShutdown[T any](i *Injector) {
	name := generateServiceNameFromInjector[T](i)
	must(getInjectorOrDefault(i).shutdownImplem(name))
}

func ShutdownNamed(i *Injector, name string) error {
	return getInjectorOrDefault(i).shutdownImplem(name)
}

func MustShutdownNamed(i *Injector, name string) {
	must(getInjectorOrDefault(i).shutdownImplem(name))
}
