package do

import "context"

type Healthchecker interface {
	HealthCheck() error
}

type HealthcheckerWithContext interface {
	HealthCheck(context.Context) error
}

type Shutdowner interface {
	Shutdown()
}

type ShutdownerWithError interface {
	Shutdown() error
}

type ShutdownerWithContext interface {
	Shutdown(context.Context)
}

type ShutdownerWithContextAndError interface {
	Shutdown(context.Context) error
}

// HealthCheck returns a service status, using type inference to determine the service name.
func HealthCheck[T any](i Injector) error {
	name := inferServiceName[T]()
	return HealthCheckNamedWithContext(context.Background(), i, name)
}

// HealthCheckWithContext returns a service status, using type inference to determine the service name.
func HealthCheckWithContext[T any](ctx context.Context, i Injector) error {
	name := inferServiceName[T]()
	return HealthCheckNamedWithContext(ctx, i, name)
}

// HealthCheckNamed returns a service status.
func HealthCheckNamed(i Injector, name string) error {
	return HealthCheckNamedWithContext(context.Background(), i, name)
}

// HealthCheckNamedWithContext returns a service status.
func HealthCheckNamedWithContext(ctx context.Context, i Injector, name string) error {
	// @TODO: should we queue the health check into the healthcheck pool ?
	return getInjectorOrDefault(i).serviceHealthCheck(ctx, name)
}

// Shutdown stops a service, using type inference to determine the service name.
func Shutdown[T any](i Injector) error {
	name := inferServiceName[T]()
	return ShutdownNamedWithContext(context.Background(), i, name)
}

// ShutdownWithContext stops a service, using type inference to determine the service name.
func ShutdownWithContext[T any](ctx context.Context, i Injector) error {
	name := inferServiceName[T]()
	return ShutdownNamedWithContext(ctx, i, name)
}

// ShutdownNamed stops a named service.
func ShutdownNamed(i Injector, name string) error {
	return ShutdownNamedWithContext(context.Background(), i, name)
}

// ShutdownNamedWithContext stops a named service.
func ShutdownNamedWithContext(ctx context.Context, i Injector, name string) error {
	return getInjectorOrDefault(i).serviceShutdown(ctx, name)
}

// MustShutdown stops a service, using type inference to determine the service name. It panics on error.
func MustShutdown[T any](i Injector) {
	must0(Shutdown[T](i))
}

// MustShutdownWithContext stops a service, using type inference to determine the service name. It panics on error.
func MustShutdownWithContext[T any](ctx context.Context, i Injector) {
	must0(ShutdownWithContext[T](ctx, i))
}

// MustShutdownNamed stops a named service. It panics on error.
func MustShutdownNamed(i Injector, name string) {
	must0(ShutdownNamed(i, name))
}

// MustShutdownNamedWithContext stops a named service. It panics on error.
func MustShutdownNamedWithContext(ctx context.Context, i Injector, name string) {
	must0(ShutdownNamedWithContext(ctx, i, name))
}
