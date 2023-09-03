package do

import (
	"github.com/samber/do/stacktrace"
)

var _ Service[int] = (*ServiceTransiant[int])(nil)
var _ healthcheckerService = (*ServiceTransiant[int])(nil)
var _ shutdownerService = (*ServiceTransiant[int])(nil)
var _ clonerService = (*ServiceTransiant[int])(nil)

type ServiceTransiant[T any] struct {
	name string

	// lazy loading
	provider Provider[T]
}

func newServiceTransiant[T any](name string, provider Provider[T]) Service[T] {
	return &ServiceTransiant[T]{
		name: name,

		provider: provider,
	}
}

func (s *ServiceTransiant[T]) getName() string {
	return s.name
}

func (s *ServiceTransiant[T]) getType() ServiceType {
	return ServiceTypeTransiant
}

func (s *ServiceTransiant[T]) getInstance(i Injector) (T, error) {
	return handleProviderPanic(s.provider, i)
}

func (s *ServiceTransiant[T]) isHealthchecker() bool {
	return false
}

func (s *ServiceTransiant[T]) healthcheck() error {
	// @TODO: implement healthcheck ?
	// It requires to store each instance of service, which is not good because of memory leaks.
	return nil
}

func (s *ServiceTransiant[T]) isShutdowner() bool {
	return false
}

func (s *ServiceTransiant[T]) shutdown() error {
	// @TODO: implement shutdown ?
	// It requires to store each instance of service, which is not good because of memory leaks.
	return nil
}

func (s *ServiceTransiant[T]) clone() any {
	return &ServiceTransiant[T]{
		name: s.name,

		provider: s.provider,
	}
}

//nolint:unused
func (s *ServiceTransiant[T]) locate() (stacktrace.Frame, []stacktrace.Frame) {
	// @TODO: implement stacktrace ?
	// It requires to store each instance of service, which is not good because of memory leaks.
	return stacktrace.Frame{}, []stacktrace.Frame{}
}
