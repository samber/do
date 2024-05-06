package do

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/samber/do/v2/stacktrace"
)

var _ Service[int] = (*serviceAlias[int, int])(nil)
var _ serviceHealthcheck = (*serviceAlias[int, int])(nil)
var _ serviceShutdown = (*serviceAlias[int, int])(nil)
var _ serviceClone = (*serviceAlias[int, int])(nil)

type serviceAlias[Initial any, Alias any] struct {
	mu         sync.RWMutex
	name       string
	scope      Injector
	targetName string

	providerFrame           stacktrace.Frame
	invokationFrames        map[stacktrace.Frame]struct{} // map garanties uniqueness
	invokationFramesCounter uint32
}

func newServiceAlias[Initial any, Alias any](name string, scope Injector, targetName string) *serviceAlias[Initial, Alias] {
	providerFrame, _ := stacktrace.NewFrameFromCaller()

	return &serviceAlias[Initial, Alias]{
		mu:         sync.RWMutex{},
		name:       name,
		scope:      scope,
		targetName: targetName,

		providerFrame:           providerFrame,
		invokationFrames:        map[stacktrace.Frame]struct{}{},
		invokationFramesCounter: 0,
	}
}

func (s *serviceAlias[Initial, Alias]) getName() string {
	return s.name
}

func (s *serviceAlias[Initial, Alias]) getType() ServiceType {
	return ServiceTypeAlias
}

func (s *serviceAlias[Initial, Alias]) getEmptyInstance() any {
	return empty[Alias]()
}

func (s *serviceAlias[Initial, Alias]) getInstanceAny(i Injector) (any, error) {
	return s.getInstance(i)
}

func (s *serviceAlias[Initial, Alias]) getInstance(i Injector) (Alias, error) {
	// Collect up to 100 invokation frames.
	// In the future, we can implement a LFU list, to evict the oldest
	// frames and keep the most recent ones, but it would be much more costly.
	if atomic.AddUint32(&s.invokationFramesCounter, 1) < MaxInvokationFrames {
		frame, ok := stacktrace.NewFrameFromCaller()
		if ok {
			s.mu.Lock()
			s.invokationFrames[frame] = struct{}{}
			s.mu.Unlock()
		}
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

func (s *serviceAlias[Initial, Alias]) isHealthchecker() bool {
	serviceAny, _, ok := s.scope.serviceGetRec(s.targetName)
	if !ok {
		return false
	}

	service, ok := serviceAny.(Service[Initial])
	if !ok {
		return false
	}

	switch service.getEmptyInstance().(type) {
	case Alias:
		return service.isHealthchecker()
	default:
		return false
	}
}

func (s *serviceAlias[Initial, Alias]) healthcheck(ctx context.Context) error {
	serviceAny, _, ok := s.scope.serviceGetRec(s.targetName)
	if !ok {
		return nil
	}

	service, ok := serviceAny.(Service[Initial])
	if !ok {
		return nil
	}

	switch service.getEmptyInstance().(type) {
	case Alias:
		return service.healthcheck(ctx)
	default:
		return fmt.Errorf("DI: could not cast `%s` as `%s`", s.targetName, s.name)
	}
}

func (s *serviceAlias[Initial, Alias]) isShutdowner() bool {
	serviceAny, _, ok := s.scope.serviceGetRec(s.targetName)
	if !ok {
		return false
	}

	service, ok := serviceAny.(Service[Initial])
	if !ok {
		return false
	}

	switch service.getEmptyInstance().(type) {
	case Alias:
		return service.isShutdowner()
	default:
		return false
	}
}

func (s *serviceAlias[Initial, Alias]) shutdown(ctx context.Context) error {
	serviceAny, _, ok := s.scope.serviceGetRec(s.targetName)
	if !ok {
		return nil
	}

	service, ok := serviceAny.(Service[Initial])
	if !ok {
		return nil
	}

	switch service.getEmptyInstance().(type) {
	case Alias:
		return service.shutdown(ctx)
	default:
		return fmt.Errorf("DI: could not cast `%s` as `%s`", s.targetName, s.name)
	}
}

func (s *serviceAlias[Initial, Alias]) clone() any {
	return &serviceAlias[Initial, Alias]{
		mu:   sync.RWMutex{},
		name: s.name,
		// scope:      s.scope,		<-- @TODO: we should inject here the cloned scope
		targetName: s.targetName,

		providerFrame:           s.providerFrame,
		invokationFrames:        map[stacktrace.Frame]struct{}{},
		invokationFramesCounter: 0,
	}
}

// nolint:unused
func (s *serviceAlias[Initial, Alias]) source() (stacktrace.Frame, []stacktrace.Frame) {
	s.mu.RLock()
	invokationFrames := make([]stacktrace.Frame, 0, len(s.invokationFrames))
	for frame := range s.invokationFrames {
		invokationFrames = append(invokationFrames, frame)
	}
	s.mu.RUnlock()

	return s.providerFrame, invokationFrames
}
