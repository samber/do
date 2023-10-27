package do

import "context"

func HealthCheck[T any](i Injector) error {
	name := inferServiceName[T]()
	return getInjectorOrDefault(i).serviceHealthCheck(context.Background(), name)
}

func HealthCheckWithContext[T any](ctx context.Context, i Injector) error {
	name := inferServiceName[T]()
	return getInjectorOrDefault(i).serviceHealthCheck(ctx, name)
}

func HealthCheckNamed(i Injector, name string) error {
	return getInjectorOrDefault(i).serviceHealthCheck(context.Background(), name)
}

func HealthCheckNamedWithContext(ctx context.Context, i Injector, name string) error {
	return getInjectorOrDefault(i).serviceHealthCheck(ctx, name)
}

func Shutdown[T any](i Injector) error {
	name := inferServiceName[T]()
	return getInjectorOrDefault(i).serviceShutdown(context.Background(), name)
}

func ShutdownWithContext[T any](ctx context.Context, i Injector) error {
	name := inferServiceName[T]()
	return getInjectorOrDefault(i).serviceShutdown(ctx, name)
}

func MustShutdown[T any](i Injector) {
	name := inferServiceName[T]()
	must0(getInjectorOrDefault(i).serviceShutdown(context.Background(), name))
}

func MustShutdownWithContext[T any](ctx context.Context, i Injector) {
	name := inferServiceName[T]()
	must0(getInjectorOrDefault(i).serviceShutdown(ctx, name))
}

func ShutdownNamed(i Injector, name string) error {
	return getInjectorOrDefault(i).serviceShutdown(context.Background(), name)
}

func ShutdownNamedWithContext(ctx context.Context, i Injector, name string) error {
	return getInjectorOrDefault(i).serviceShutdown(ctx, name)
}

func MustShutdownNamed(i Injector, name string) {
	must0(getInjectorOrDefault(i).serviceShutdown(context.Background(), name))
}

func MustShutdownNamedWithContext(ctx context.Context, i Injector, name string) {
	must0(getInjectorOrDefault(i).serviceShutdown(ctx, name))
}
