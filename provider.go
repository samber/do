package do

import "fmt"

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
