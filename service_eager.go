package do

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/samber/do/v2/stacktrace"
)

var _ Service[int] = (*serviceEager[int])(nil)
var _ serviceHealthcheck = (*serviceEager[int])(nil)
var _ serviceShutdown = (*serviceEager[int])(nil)
var _ serviceClone = (*serviceEager[int])(nil)

type serviceEager[T any] struct {
	name     string
	typeName string
	instance T

	providerFrame           stacktrace.Frame
	invokationFrames        map[stacktrace.Frame]struct{} // map garanties uniqueness
	invokationFramesMu      sync.RWMutex
	invokationFramesCounter uint32
}

func newServiceEager[T any](name string, instance T) *serviceEager[T] {
	providerFrame, _ := stacktrace.NewFrameFromCaller()

	return &serviceEager[T]{
		name:     name,
		typeName: inferServiceName[T](),
		instance: instance,

		providerFrame:           providerFrame,
		invokationFrames:        map[stacktrace.Frame]struct{}{},
		invokationFramesMu:      sync.RWMutex{},
		invokationFramesCounter: 0,
	}
}

func (s *serviceEager[T]) getName() string {
	return s.name
}

func (s *serviceEager[T]) getTypeName() string {
	return s.typeName
}

func (s *serviceEager[T]) getServiceType() ServiceType {
	return ServiceTypeEager
}

func (s *serviceEager[T]) getEmptyInstance() any {
	return empty[T]()
}

func (s *serviceEager[T]) getInstanceAny(i Injector) (any, error) {
	return s.getInstance(i)
}

func (s *serviceEager[T]) getInstance(i Injector) (T, error) {
	// Collect up to 100 invokation frames.
	// In the future, we can implement a LFU list, to evict the oldest
	// frames and keep the most recent ones, but it would be much more costly.
	if atomic.AddUint32(&s.invokationFramesCounter, 1) < MaxInvokationFrames {
		s.invokationFramesMu.Lock()
		frame, ok := stacktrace.NewFrameFromCaller()
		if ok {
			s.invokationFrames[frame] = struct{}{}
		}
		s.invokationFramesMu.Unlock()
	}

	return s.instance, nil
}

func (s *serviceEager[T]) isHealthchecker() bool {
	_, ok1 := any(s.instance).(HealthcheckerWithContext)
	_, ok2 := any(s.instance).(Healthchecker)
	return ok1 || ok2
}

func (s *serviceEager[T]) healthcheck(ctx context.Context) error {
	if instance, ok := any(s.instance).(HealthcheckerWithContext); ok {
		return instance.HealthCheck(ctx)
	} else if instance, ok := any(s.instance).(Healthchecker); ok {
		return instance.HealthCheck()
	}

	return nil
}

func (s *serviceEager[T]) isShutdowner() bool {
	_, ok1 := any(s.instance).(ShutdownerWithContextAndError)
	_, ok2 := any(s.instance).(ShutdownerWithError)
	_, ok3 := any(s.instance).(ShutdownerWithContext)
	_, ok4 := any(s.instance).(Shutdowner)
	return ok1 || ok2 || ok3 || ok4
}

func (s *serviceEager[T]) shutdown(ctx context.Context) error {
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

func (s *serviceEager[T]) clone() any {
	return &serviceEager[T]{
		name:     s.name,
		typeName: s.typeName,
		instance: s.instance,

		providerFrame:           s.providerFrame,
		invokationFrames:        map[stacktrace.Frame]struct{}{},
		invokationFramesMu:      sync.RWMutex{},
		invokationFramesCounter: 0,
	}
}

// nolint:unused
func (s *serviceEager[T]) source() (stacktrace.Frame, []stacktrace.Frame) {
	s.invokationFramesMu.RLock()
	invokationFrames := make([]stacktrace.Frame, 0, len(s.invokationFrames))
	for frame := range s.invokationFrames {
		invokationFrames = append(invokationFrames, frame)
	}
	s.invokationFramesMu.RUnlock()

	return s.providerFrame, invokationFrames
}
