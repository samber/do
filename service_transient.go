package do

import (
	"context"

	"github.com/samber/do/v2/stacktrace"
)

var _ Service[int] = (*ServiceTransient[int])(nil)
var _ serviceHealthcheck = (*ServiceTransient[int])(nil)
var _ serviceShutdown = (*ServiceTransient[int])(nil)
var _ serviceClone = (*ServiceTransient[int])(nil)

type ServiceTransient[T any] struct {
	name string

	// lazy loading
	provider Provider[T]
}

func newServiceTransient[T any](name string, provider Provider[T]) *ServiceTransient[T] {
	return &ServiceTransient[T]{
		name: name,

		provider: provider,
	}
}

func (s *ServiceTransient[T]) getName() string {
	return s.name
}

func (s *ServiceTransient[T]) getType() ServiceType {
	return ServiceTypeTransient
}

func (s *ServiceTransient[T]) getEmptyInstance() any {
	return empty[T]()
}

func (s *ServiceTransient[T]) getInstanceAny(i Injector) (any, error) {
	return s.getInstance(i)
}

func (s *ServiceTransient[T]) getInstance(i Injector) (T, error) {
	return handleProviderPanic(s.provider, i)
}

func (s *ServiceTransient[T]) isHealthchecker() bool {
	return false
}

func (s *ServiceTransient[T]) healthcheck(ctx context.Context) error {
	// @TODO: implement healthcheck ?
	// It requires to store each instance of service, which is not good because of memory leaks.
	return nil
}

func (s *ServiceTransient[T]) isShutdowner() bool {
	return false
}

func (s *ServiceTransient[T]) shutdown(ctx context.Context) error {
	// @TODO: implement shutdown ?
	// It requires to store each instance of service, which is not good because of memory leaks.
	return nil
}

func (s *ServiceTransient[T]) clone() any {
	return &ServiceTransient[T]{
		name: s.name,

		provider: s.provider,
	}
}

//nolint:unused
func (s *ServiceTransient[T]) source() (stacktrace.Frame, []stacktrace.Frame) {
	// @TODO: implement stacktrace ?
	// It requires to store each instance of service, which is not good because of memory leaks.
	return stacktrace.Frame{}, []stacktrace.Frame{}
}
