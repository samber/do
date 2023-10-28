package do

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/samber/do/v2/stacktrace"
)

var _ Service[int] = (*ServiceLazy[int])(nil)
var _ healthcheckerService = (*ServiceLazy[int])(nil)
var _ shutdownerService = (*ServiceLazy[int])(nil)
var _ clonerService = (*ServiceLazy[int])(nil)

type ServiceLazy[T any] struct {
	mu       sync.RWMutex
	name     string
	instance T

	// lazy loading
	built     bool
	buildTime time.Duration // @TODO: shoud be exported ?
	provider  Provider[T]

	providerFrame    stacktrace.Frame
	invokationFrames []stacktrace.Frame
}

func newServiceLazy[T any](name string, provider Provider[T]) Service[T] {
	providerFrame, _ := stacktrace.NewFrameFromPtr(reflect.ValueOf(provider).Pointer())

	return &ServiceLazy[T]{
		mu:   sync.RWMutex{},
		name: name,

		built:     false,
		buildTime: 0,
		provider:  provider,

		providerFrame:    providerFrame,
		invokationFrames: []stacktrace.Frame{},
	}
}

func (s *ServiceLazy[T]) getName() string {
	return s.name
}

func (s *ServiceLazy[T]) getType() ServiceType {
	return ServiceTypeLazy
}

func (s *ServiceLazy[T]) getInstance(i Injector) (T, error) {
	frame, ok := stacktrace.NewFrameFromCaller()

	s.mu.Lock()
	defer s.mu.Unlock()

	if ok {
		s.invokationFrames = append(s.invokationFrames, frame)
	}

	if !s.built {
		err := s.build(i)
		if err != nil {
			return empty[T](), err
		}
	}

	return s.instance, nil
}

func (s *ServiceLazy[T]) build(i Injector) (err error) {
	start := time.Now()

	instance, err := handleProviderPanic(s.provider, i)
	if err != nil {
		return err
	}

	s.instance = instance
	s.built = true
	s.buildTime = time.Since(start)

	return nil
}

func (s *ServiceLazy[T]) isHealthchecker() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.built {
		return false
	}

	_, ok1 := any(s.instance).(HealthcheckerWithContext)
	_, ok2 := any(s.instance).(Healthchecker)
	return ok1 || ok2
}

func (s *ServiceLazy[T]) healthcheck(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.built {
		return nil
	}

	if instance, ok := any(s.instance).(HealthcheckerWithContext); ok {
		return instance.HealthCheckWithContext(ctx)
	} else if instance, ok := any(s.instance).(Healthchecker); ok {
		return instance.HealthCheck()
	}

	return nil
}

func (s *ServiceLazy[T]) isShutdowner() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.built {
		return false
	}

	_, ok1 := any(s.instance).(ShutdownerWithContextAndError)
	_, ok2 := any(s.instance).(ShutdownerWithError)
	_, ok3 := any(s.instance).(ShutdownerWithContext)
	_, ok4 := any(s.instance).(Shutdowner)
	return ok1 || ok2 || ok3 || ok4
}

func (s *ServiceLazy[T]) shutdown(ctx context.Context) error {
	s.mu.Lock()

	defer func() {
		// whatever the outcome, reset `build` flag and instance
		s.built = false
		s.instance = empty[T]()
		s.mu.Unlock()
	}()

	if !s.built {
		return nil
	}

	if instance, ok := any(s.instance).(ShutdownerWithContextAndError); ok {
		return instance.Shutdown(ctx)
	} else if instance, ok := any(s.instance).(ShutdownerWithError); ok {
		return instance.Shutdown()
	} else if instance, ok := any(s.instance).(ShutdownerWithContext); ok {
		instance.Shutdown(ctx)
		return nil
	} else if instance, ok := any(s.instance).(Shutdowner); ok {
		instance.Shutdown()
		return nil
	}

	return nil
}

func (s *ServiceLazy[T]) clone() any {
	// reset `build` flag and instance
	return &ServiceLazy[T]{
		mu:   sync.RWMutex{},
		name: s.name,

		built:    false,
		provider: s.provider,

		providerFrame:    s.providerFrame,
		invokationFrames: []stacktrace.Frame{},
	}
}

//nolint:unused
func (s *ServiceLazy[T]) locate() (stacktrace.Frame, []stacktrace.Frame) {
	return s.providerFrame, s.invokationFrames
}
