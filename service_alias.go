package do

import (
	"context"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/samber/do/v2/stacktrace"
)

var (
	_ serviceWrapper[int]       = (*serviceAlias[int, int])(nil)
	_ serviceWrapperHealthcheck = (*serviceAlias[int, int])(nil)
	_ serviceWrapperShutdown    = (*serviceAlias[int, int])(nil)
	_ serviceWrapperClone       = (*serviceAlias[int, int])(nil)
)

type serviceAlias[Initial any, Alias any] struct {
	mu         sync.RWMutex
	name       string
	typeName   string // string representation of the Alias type
	scope      Injector
	targetName string // string representation of the Initial type

	providerFrame           stacktrace.Frame
	invokationFrames        map[stacktrace.Frame]struct{} // map garanties uniqueness
	invokationFramesCounter uint32
}

func newServiceAlias[Initial any, Alias any](
	name string,
	scope Injector,
	targetName string,
) *serviceAlias[Initial, Alias] {
	providerFrame, _ := stacktrace.NewFrameFromCaller()

	return &serviceAlias[Initial, Alias]{
		mu:         sync.RWMutex{},
		name:       name,
		typeName:   inferServiceName[Alias](),
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

func (s *serviceAlias[Initial, Alias]) getTypeName() string {
	return s.typeName
}

func (s *serviceAlias[Initial, Alias]) getServiceType() ServiceType {
	return ServiceTypeAlias
}

func (s *serviceAlias[Initial, Alias]) getReflectType() reflect.Type {
	return reflect.TypeOf((*Alias)(nil)).Elem() // if T is a pointer or interface, it will return a typed nil
}

func (s *serviceAlias[Initial, Alias]) getInstanceAny(i Injector) (any, error) {
	return s.getInstance(i)
}

func (s *serviceAlias[Initial, Alias]) getInstance(i Injector) (Alias, error) {
	// Collect up to 100 invokation frames.
	// In the future, we can implement a LFU list, to evict the oldest
	// frames and keep the most recent ones, but it would be much more costly.
	if atomic.AddUint32(&s.invokationFramesCounter, 1) < MaxInvocationFrames {
		frame, ok := stacktrace.NewFrameFromCaller()
		if ok {
			s.mu.Lock()
			s.invokationFrames[frame] = struct{}{}
			s.mu.Unlock()
		}
	}

	// Use the virtual scope received as parameter to ensure proper circular dependency detection.
	// The injector passed here should be a virtual scope that contains the current invocation chain
	instance, err := invokeByName[Initial](i, s.targetName)
	if err != nil {
		return empty[Alias](), err
	}

	switch target := any(instance).(type) {
	case Alias:
		return target, nil
	default:
		// should never happen, since invoke() checks the type
		return empty[Alias](), serviceTypeMismatch(inferServiceName[Alias](), inferServiceName[Initial]())
	}
}

func (s *serviceAlias[Initial, Alias]) isHealthchecker() bool {
	serviceAny, _, ok := s.scope.serviceGetRec(s.targetName)
	if !ok {
		return false
	}

	service, ok := serviceAny.(serviceWrapperIsHealthchecker)
	if !ok {
		return false
	}

	return service.isHealthchecker()
}

func (s *serviceAlias[Initial, Alias]) healthcheck(ctx context.Context) error {
	serviceAny, _, ok := s.scope.serviceGetRec(s.targetName)
	if !ok {
		return nil
	}

	service, ok := serviceAny.(serviceWrapperHealthcheck)
	if !ok {
		return nil
	}

	return service.healthcheck(ctx)
}

func (s *serviceAlias[Initial, Alias]) isShutdowner() bool {
	serviceAny, _, ok := s.scope.serviceGetRec(s.targetName)
	if !ok {
		return false
	}

	service, ok := serviceAny.(serviceWrapperIsShutdowner)
	if !ok {
		return false
	}

	return service.isShutdowner()
}

func (s *serviceAlias[Initial, Alias]) shutdown(ctx context.Context) error {
	serviceAny, _, ok := s.scope.serviceGetRec(s.targetName)
	if !ok {
		return nil
	}

	service, ok := serviceAny.(serviceWrapperShutdown)
	if !ok {
		return nil
	}

	return service.shutdown(ctx)
}

func (s *serviceAlias[Initial, Alias]) clone(newScope Injector) any {
	return &serviceAlias[Initial, Alias]{
		mu:         sync.RWMutex{},
		name:       s.name,
		typeName:   s.typeName,
		scope:      newScope,
		targetName: s.targetName,

		providerFrame:           s.providerFrame,
		invokationFrames:        map[stacktrace.Frame]struct{}{},
		invokationFramesCounter: 0,
	}
}

func (s *serviceAlias[Initial, Alias]) source() (stacktrace.Frame, []stacktrace.Frame) {
	s.mu.RLock()
	invokationFrames := make([]stacktrace.Frame, 0, len(s.invokationFrames))
	for frame := range s.invokationFrames {
		invokationFrames = append(invokationFrames, frame)
	}
	s.mu.RUnlock()

	return s.providerFrame, invokationFrames
}
