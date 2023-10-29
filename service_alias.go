package do

import (
	"context"
	"fmt"
	"sync"

	"github.com/samber/do/v2/stacktrace"
)

var _ Service[int] = (*ServiceAlias[int])(nil)
var _ healthcheckerService = (*ServiceAlias[int])(nil)
var _ shutdownerService = (*ServiceAlias[int])(nil)
var _ clonerService = (*ServiceAlias[int])(nil)

type ServiceAlias[T any] struct {
	mu         sync.RWMutex
	name       string
	scope      Injector
	targetName string

	providerFrame    stacktrace.Frame
	invokationFrames []stacktrace.Frame
}

func newServiceAlias[T any](name string, scope Injector, targetName string) Service[T] {
	providerFrame, _ := stacktrace.NewFrameFromCaller()

	return &ServiceAlias[T]{
		mu:         sync.RWMutex{},
		name:       name,
		scope:      scope,
		targetName: targetName,

		providerFrame:    providerFrame,
		invokationFrames: []stacktrace.Frame{},
	}
}

func (s *ServiceAlias[T]) getName() string {
	return s.name
}

func (s *ServiceAlias[T]) getType() ServiceType {
	return ServiceTypeAlias
}

func (s *ServiceAlias[T]) getInstance(i Injector) (T, error) {
	frame, ok := stacktrace.NewFrameFromCaller()
	if ok {
		s.mu.Lock()
		s.invokationFrames = append(s.invokationFrames, frame) // @TODO: potential memory leak
		s.mu.Unlock()
	}

	instance, _, ok := s.scope.serviceGetRec(s.targetName)
	if !ok {
		return empty[T](), serviceNotFound(i, s.name)
	}

	target, ok := any(instance).(T)
	if !ok {
		return empty[T](), fmt.Errorf("DI: could not cast service `%s` to type `%s`", s.name, s.targetName)
	}

	return target, nil
}

func (s *ServiceAlias[T]) isHealthchecker() bool {
	instance, _, ok := s.scope.serviceGetRec(s.targetName)
	if !ok {
		return false
	}

	_, ok1 := any(instance).(HealthcheckerWithContext)
	_, ok2 := any(instance).(Healthchecker)
	return ok1 || ok2
}

func (s *ServiceAlias[T]) healthcheck(ctx context.Context) error {
	instance, _, ok := s.scope.serviceGetRec(s.targetName)
	if !ok {
		return nil
	}

	if instance, ok := any(instance).(HealthcheckerWithContext); ok {
		return instance.HealthCheckWithContext(ctx)
	} else if instance, ok := any(instance).(Healthchecker); ok {
		return instance.HealthCheck()
	}

	return nil
}

func (s *ServiceAlias[T]) isShutdowner() bool {
	instance, _, ok := s.scope.serviceGetRec(s.targetName)
	if !ok {
		return false
	}

	_, ok1 := any(instance).(ShutdownerWithContextAndError)
	_, ok2 := any(instance).(ShutdownerWithError)
	_, ok3 := any(instance).(ShutdownerWithContext)
	_, ok4 := any(instance).(Shutdowner)
	return ok1 || ok2 || ok3 || ok4
}

func (s *ServiceAlias[T]) shutdown(ctx context.Context) error {
	instance, _, ok := s.scope.serviceGetRec(s.targetName)
	if !ok {
		return nil
	}

	if instance, ok := any(instance).(ShutdownerWithContextAndError); ok {
		return instance.Shutdown(ctx)
	} else if instance, ok := any(instance).(ShutdownerWithError); ok {
		return instance.Shutdown()
	} else if instance, ok := any(instance).(ShutdownerWithContext); ok {
		instance.Shutdown(ctx)
		return nil
	} else if instance, ok := any(instance).(Shutdowner); ok {
		instance.Shutdown()
		return nil
	}

	return nil
}

func (s *ServiceAlias[T]) clone() any {
	return &ServiceAlias[T]{
		mu:   sync.RWMutex{},
		name: s.name,
		// scope:      s.scope,		<-- we should inject here the cloned scope
		targetName: s.targetName,

		providerFrame:    s.providerFrame,
		invokationFrames: []stacktrace.Frame{},
	}
}

// nolint:unused
func (s *ServiceAlias[T]) source() (stacktrace.Frame, []stacktrace.Frame) {
	return s.providerFrame, s.invokationFrames
}
