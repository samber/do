package do

import (
	"context"

	"github.com/samber/do/v2/stacktrace"
)

var _ Service[int] = (*serviceTransient[int])(nil)
var _ serviceHealthcheck = (*serviceTransient[int])(nil)
var _ serviceShutdown = (*serviceTransient[int])(nil)
var _ serviceClone = (*serviceTransient[int])(nil)

type serviceTransient[T any] struct {
	name     string
	typeName string

	// lazy loading
	provider Provider[T]
}

func newServiceTransient[T any](name string, provider Provider[T]) *serviceTransient[T] {
	return &serviceTransient[T]{
		name:     name,
		typeName: inferServiceName[T](),

		provider: provider,
	}
}

func (s *serviceTransient[T]) getName() string {
	return s.name
}

func (s *serviceTransient[T]) getTypeName() string {
	return s.typeName
}

func (s *serviceTransient[T]) getServiceType() ServiceType {
	return ServiceTypeTransient
}

func (s *serviceTransient[T]) getEmptyInstance() any {
	return empty[T]()
}

func (s *serviceTransient[T]) getInstanceAny(i Injector) (any, error) {
	return s.getInstance(i)
}

func (s *serviceTransient[T]) getInstance(i Injector) (T, error) {
	return handleProviderPanic(s.provider, i)
}

func (s *serviceTransient[T]) isHealthchecker() bool {
	return false
}

func (s *serviceTransient[T]) healthcheck(ctx context.Context) error {
	// @TODO: implement healthcheck ?
	// It requires to store each instance of service, which is not good because of memory leaks.
	return nil
}

func (s *serviceTransient[T]) isShutdowner() bool {
	return false
}

func (s *serviceTransient[T]) shutdown(ctx context.Context) error {
	// @TODO: implement shutdown ?
	// It requires to store each instance of service, which is not good because of memory leaks.
	return nil
}

func (s *serviceTransient[T]) clone() any {
	return &serviceTransient[T]{
		name:     s.name,
		typeName: s.typeName,

		provider: s.provider,
	}
}

//nolint:unused
func (s *serviceTransient[T]) source() (stacktrace.Frame, []stacktrace.Frame) {
	// @TODO: implement stacktrace ?
	// It requires to store each instance of service, which is not good because of memory leaks.
	return stacktrace.Frame{}, []stacktrace.Frame{}
}
