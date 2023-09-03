package do

import (
	"reflect"
	"sync"
	"time"

	"github.com/samber/do/stacktrace"
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

	_, ok := any(s.instance).(Healthchecker)
	return ok
}

func (s *ServiceLazy[T]) healthcheck() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.built {
		return nil
	}

	instance, ok := any(s.instance).(Healthchecker)
	if ok {
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

	_, ok := any(s.instance).(Shutdowner)
	return ok
}

func (s *ServiceLazy[T]) shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.built {
		// @TODO: mark s.build as false ?
		return nil
	}

	instance, ok := any(s.instance).(Shutdowner)
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
