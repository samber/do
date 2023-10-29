package do

import "context"

// HealthCheck returns a service status, using type inference to determine the service name.
func HealthCheck[T any](i Injector) error {
	name := inferServiceName[T]()
	return getInjectorOrDefault(i).serviceHealthCheck(context.Background(), name)
}

// HealthCheckWithContext returns a service status, using type inference to determine the service name.
func HealthCheckWithContext[T any](ctx context.Context, i Injector) error {
	name := inferServiceName[T]()
	return getInjectorOrDefault(i).serviceHealthCheck(ctx, name)
}

// HealthCheckNamed returns a service status.
func HealthCheckNamed(i Injector, name string) error {
	return getInjectorOrDefault(i).serviceHealthCheck(context.Background(), name)
}

// HealthCheckNamedWithContext returns a service status.
func HealthCheckNamedWithContext(ctx context.Context, i Injector, name string) error {
	return getInjectorOrDefault(i).serviceHealthCheck(ctx, name)
}

// Shutdown stops a service, using type inference to determine the service name.
func Shutdown[T any](i Injector) error {
	name := inferServiceName[T]()
	return getInjectorOrDefault(i).serviceShutdown(context.Background(), name)
}

// ShutdownWithContext stops a service, using type inference to determine the service name.
func ShutdownWithContext[T any](ctx context.Context, i Injector) error {
	name := inferServiceName[T]()
	return getInjectorOrDefault(i).serviceShutdown(ctx, name)
}

// MustShutdown stops a service, using type inference to determine the service name. It panics on error.
func MustShutdown[T any](i Injector) {
	name := inferServiceName[T]()
	must0(getInjectorOrDefault(i).serviceShutdown(context.Background(), name))
}

// MustShutdownWithContext stops a service, using type inference to determine the service name. It panics on error.
func MustShutdownWithContext[T any](ctx context.Context, i Injector) {
	name := inferServiceName[T]()
	must0(getInjectorOrDefault(i).serviceShutdown(ctx, name))
}

// ShutdownNamed stops a named service.
func ShutdownNamed(i Injector, name string) error {
	return getInjectorOrDefault(i).serviceShutdown(context.Background(), name)
}

// ShutdownNamedWithContext stops a named service.
func ShutdownNamedWithContext(ctx context.Context, i Injector, name string) error {
	return getInjectorOrDefault(i).serviceShutdown(ctx, name)
}

// MustShutdownNamed stops a named service. It panics on error.
func MustShutdownNamed(i Injector, name string) {
	must0(getInjectorOrDefault(i).serviceShutdown(context.Background(), name))
}

// MustShutdownNamedWithContext stops a named service. It panics on error.
func MustShutdownNamedWithContext(ctx context.Context, i Injector, name string) {
	must0(getInjectorOrDefault(i).serviceShutdown(ctx, name))
}
