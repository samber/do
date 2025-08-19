package do

import "fmt"

// Provider[T] is a function type that creates and returns a service instance of type T.
// This is the core abstraction for service creation in the dependency injection container.
//
// The provider function receives an Injector instance that can be used to resolve
// dependencies for the service being created.
//
// Example:
//
//	func NewMyService(i do.Injector) (*MyService, error) {
//	    db := do.MustInvoke[*Database](i)
//	    config := do.MustInvoke[*Config](i)
//	    return &MyService{DB: db, Config: config}, nil
//	}
//
//	// Register the provider
//	do.Provide(injector, NewMyService)
type Provider[T any] func(Injector) (T, error)

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

	return
}
