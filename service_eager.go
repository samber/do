package do

import (
	"context"
	"sync"

	"github.com/samber/do/v2/stacktrace"
)

var _ Service[int] = (*ServiceEager[int])(nil)
var _ healthcheckerService = (*ServiceEager[int])(nil)
var _ shutdownerService = (*ServiceEager[int])(nil)
var _ clonerService = (*ServiceEager[int])(nil)

type ServiceEager[T any] struct {
	mu       sync.RWMutex
	name     string
	instance T

	providerFrame    stacktrace.Frame
	invokationFrames []stacktrace.Frame
}

func newServiceEager[T any](name string, instance T) Service[T] {
	providerFrame, _ := stacktrace.NewFrameFromCaller()

	return &ServiceEager[T]{
		mu:       sync.RWMutex{},
		name:     name,
		instance: instance,

		providerFrame:    providerFrame,
		invokationFrames: []stacktrace.Frame{},
	}
}

func (s *ServiceEager[T]) getName() string {
	return s.name
}

func (s *ServiceEager[T]) getType() ServiceType {
	return ServiceTypeEager
}

func (s *ServiceEager[T]) getInstance(i Injector) (T, error) {
	frame, ok := stacktrace.NewFrameFromCaller()
	if ok {
		s.mu.Lock()
		s.invokationFrames = append(s.invokationFrames, frame)
		s.mu.Unlock()
	}

	return s.instance, nil
}

func (s *ServiceEager[T]) isHealthchecker() bool {
	_, ok1 := any(s.instance).(HealthcheckerWithContext)
	_, ok2 := any(s.instance).(Healthchecker)
	return ok1 || ok2
}

func (s *ServiceEager[T]) healthcheck(ctx context.Context) error {
	if instance, ok := any(s.instance).(HealthcheckerWithContext); ok {
		return instance.HealthCheckWithContext(ctx)
	} else if instance, ok := any(s.instance).(Healthchecker); ok {
		return instance.HealthCheck()
	}

	return nil
}

func (s *ServiceEager[T]) isShutdowner() bool {
	_, ok1 := any(s.instance).(ShutdownerWithContext)
	_, ok2 := any(s.instance).(Shutdowner)
	return ok1 || ok2
}

func (s *ServiceEager[T]) shutdown(ctx context.Context) error {
	if instance, ok := any(s.instance).(ShutdownerWithContext); ok {
		return instance.ShutdownWithContext(ctx)
	} else if instance, ok := any(s.instance).(Shutdowner); ok {
		return instance.Shutdown()
	}

	return nil
}

func (s *ServiceEager[T]) clone() any {
	return &ServiceEager[T]{
		mu:       sync.RWMutex{},
		name:     s.name,
		instance: s.instance,

		providerFrame:    s.providerFrame,
		invokationFrames: []stacktrace.Frame{},
	}
}

// nolint:unused
func (s *ServiceEager[T]) locate() (stacktrace.Frame, []stacktrace.Frame) {
	return s.providerFrame, s.invokationFrames
}
