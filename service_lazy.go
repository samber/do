package do

import (
	"sync"
)

type Provider[T any] func(*Injector) (T, error)

type ServiceLazy[T any] struct {
	mu       sync.RWMutex
	name     string
	instance T

	// lazy loading
	built    bool
	provider Provider[T]
}

func newServiceLazy[T any](name string, provider Provider[T]) Service[T] {
	return &ServiceLazy[T]{
		name: name,

		built:    false,
		provider: provider,
	}
}

//nolint:unused
func (s *ServiceLazy[T]) getName() string {
	return s.name
}

//nolint:unused
func (s *ServiceLazy[T]) getInstance(i *Injector) (T, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.built {
		err := s.build(i)
		if err != nil {
			return empty[T](), err
		}
	}

	return s.instance, nil
}

//nolint:unused
func (s *ServiceLazy[T]) build(i *Injector) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				panic(r)
			}
		}
	}()

	instance, err := s.provider(i)
	if err != nil {
		return err
	}

	s.instance = instance
	s.built = true

	return nil
}

func (s *ServiceLazy[T]) healthcheck() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.built {
		return nil
	}

	instance, ok := any(s.instance).(Healthcheckable)
	if ok {
		return instance.HealthCheck()
	}

	return nil
}

func (s *ServiceLazy[T]) shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.built {
		return nil
	}

	instance, ok := any(s.instance).(Shutdownable)
	if ok {
		err := instance.Shutdown()
		if err != nil {
			return err
		}
	}

	s.built = false
	s.instance = empty[T]()

	return nil
}

func (s *ServiceLazy[T]) clone() any {
	// reset `build` flag and instance
	return &ServiceLazy[T]{
		name: s.name,

		built:    false,
		provider: s.provider,
	}
}
