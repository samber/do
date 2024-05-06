package do

func Package(services ...func(i Injector)) func(Injector) {
	return func(injector Injector) {
		for i := range services {
			services[i](injector)
		}
	}
}

func Lazy[T any](p Provider[T]) func(Injector) {
	return func(injector Injector) {
		Provide(injector, p)
	}
}

func LazyNamed[T any](serviceName string, p Provider[T]) func(Injector) {
	return func(injector Injector) {
		ProvideNamed(injector, serviceName, p)
	}
}

func Eager[T any](value T) func(Injector) {
	return func(injector Injector) {
		ProvideValue(injector, value)
	}
}

func EagerNamed[T any](serviceName string, value T) func(Injector) {
	return func(injector Injector) {
		ProvideNamedValue(injector, serviceName, value)
	}
}

func Transient[T any](p Provider[T]) func(Injector) {
	return func(injector Injector) {
		ProvideTransient(injector, p)
	}
}

func TransientNamed[T any](serviceName string, p Provider[T]) func(Injector) {
	return func(injector Injector) {
		ProvideNamedTransient(injector, serviceName, p)
	}
}

func Bind[Initial any, Alias any]() func(Injector) {
	return func(injector Injector) {
		MustAs[Initial, Alias](injector)
	}
}

func BindNamed[Initial any, Alias any](initial string, alias string) func(Injector) {
	return func(injector Injector) {
		MustAsNamed[Initial, Alias](injector, initial, alias)
	}
}
