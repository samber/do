package do

import (
	"context"
	"fmt"
	"sync"

	"github.com/samber/do/v2/stacktrace"
)

var _ Service[int] = (*ServiceAlias[int, int])(nil)
var _ serviceHealthcheck = (*ServiceAlias[int, int])(nil)
var _ serviceShutdown = (*ServiceAlias[int, int])(nil)
var _ serviceClone = (*ServiceAlias[int, int])(nil)

type ServiceAlias[Initial any, Alias any] struct {
	mu         sync.RWMutex
	name       string
	scope      Injector
	targetName string

	providerFrame    stacktrace.Frame
	invokationFrames []stacktrace.Frame
}

func newServiceAlias[Initial any, Alias any](name string, scope Injector, targetName string) *ServiceAlias[Initial, Alias] {
	providerFrame, _ := stacktrace.NewFrameFromCaller()

	return &ServiceAlias[Initial, Alias]{
		mu:         sync.RWMutex{},
		name:       name,
		scope:      scope,
		targetName: targetName,

		providerFrame:    providerFrame,
		invokationFrames: []stacktrace.Frame{},
	}
}

func (s *ServiceAlias[Initial, Alias]) getName() string {
	return s.name
}

func (s *ServiceAlias[Initial, Alias]) getType() ServiceType {
	return ServiceTypeAlias
}

func (s *ServiceAlias[Initial, Alias]) getEmptyInstance() any {
	return empty[Alias]()
}

func (s *ServiceAlias[Initial, Alias]) getInstanceAny(i Injector) (any, error) {
	return s.getInstance(i)
}

func (s *ServiceAlias[Initial, Alias]) getInstance(i Injector) (Alias, error) {
	frame, ok := stacktrace.NewFrameFromCaller()
	if ok {
		s.mu.Lock()
		s.invokationFrames = append(s.invokationFrames, frame) // @TODO: potential memory leak
		s.mu.Unlock()
	}

	instance, err := invokeByName[Initial](s.scope, s.targetName)
	if err != nil {
		return empty[Alias](), err
	}

	switch target := any(instance).(type) {
	case Alias:
		return target, nil
	default:
		// should never happen, since invoke() checks the type
		return empty[Alias](), fmt.Errorf("DI: could not cast `%s` as `%s`", s.targetName, s.name)
	}
}

func (s *ServiceAlias[Initial, Alias]) isHealthchecker() bool {
	serviceAny, _, ok := s.scope.serviceGetRec(s.targetName)
	if !ok {
		return false
	}

	service, ok := serviceAny.(Service[Initial])
	if !ok {
		return false
	}

	// @TODO: check convertible to `Alias`?

	return service.isHealthchecker()
}

func (s *ServiceAlias[Initial, Alias]) healthcheck(ctx context.Context) error {
	serviceAny, _, ok := s.scope.serviceGetRec(s.targetName)
	if !ok {
		return nil
	}

	service, ok := serviceAny.(Service[Initial])
	if !ok {
		return nil
	}

	// @TODO: check convertible to `Alias`?

	return service.healthcheck(ctx)
}

func (s *ServiceAlias[Initial, Alias]) isShutdowner() bool {
	serviceAny, _, ok := s.scope.serviceGetRec(s.targetName)
	if !ok {
		return false
	}

	service, ok := serviceAny.(Service[Initial])
	if !ok {
		return false
	}

	// @TODO: check convertible to `Alias`?

	return service.isShutdowner()
}

func (s *ServiceAlias[Initial, Alias]) shutdown(ctx context.Context) error {
	serviceAny, _, ok := s.scope.serviceGetRec(s.targetName)
	if !ok {
		return nil
	}

	service, ok := serviceAny.(Service[Initial])
	if !ok {
		return nil
	}

	// @TODO: check convertible to `Alias`?

	return service.shutdown(ctx)
}

func (s *ServiceAlias[Initial, Alias]) clone() any {
	return &ServiceAlias[Initial, Alias]{
		mu:   sync.RWMutex{},
		name: s.name,
		// scope:      s.scope,		<-- we should inject here the cloned scope
		targetName: s.targetName,

		providerFrame:    s.providerFrame,
		invokationFrames: []stacktrace.Frame{},
	}
}

// nolint:unused
func (s *ServiceAlias[Initial, Alias]) source() (stacktrace.Frame, []stacktrace.Frame) {
	return s.providerFrame, s.invokationFrames
}
