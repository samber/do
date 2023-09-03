package do

func HealthCheck[T any](i Injector) error {
	name := inferServiceName[T]()
	return getInjectorOrDefault(i).serviceHealthCheck(name)
}

func HealthCheckNamed(i Injector, name string) error {
	return getInjectorOrDefault(i).serviceHealthCheck(name)
}

func Shutdown[T any](i Injector) error {
	name := inferServiceName[T]()
	return getInjectorOrDefault(i).serviceShutdown(name)
}

func MustShutdown[T any](i Injector) {
	name := inferServiceName[T]()
	must0(getInjectorOrDefault(i).serviceShutdown(name))
}

func ShutdownNamed(i Injector, name string) error {
	return getInjectorOrDefault(i).serviceShutdown(name)
}

func MustShutdownNamed(i Injector, name string) {
	must0(getInjectorOrDefault(i).serviceShutdown(name))
}
