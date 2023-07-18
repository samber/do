package do

type ServiceEager[T any] struct {
	name     string
	instance T

	shutdownFunc shutdownFunc[T]
}

func newServiceEager[T any](name string, instance T, opts ...ServiceOpt[T]) Service[T] {
	s := &ServiceEager[T]{
		name:     name,
		instance: instance,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

//nolint:unused
func (s *ServiceEager[T]) getName() string {
	return s.name
}

//nolint:unused
func (s *ServiceEager[T]) getInstance(i *Injector) (T, error) {
	return s.instance, nil
}

func (s *ServiceEager[T]) healthcheck() error {
	instance, ok := any(s.instance).(Healthcheckable)
	if ok {
		return instance.HealthCheck()
	}

	return nil
}

//nolint:unused
func (s *ServiceEager[T]) setShutdownFunc(shutdownFunc shutdownFunc[T]) {
	s.shutdownFunc = shutdownFunc
}

func (s *ServiceEager[T]) shutdown() error {
	if s.shutdownFunc != nil {
		err := s.shutdownFunc(s.instance)
		if err != nil {
			return err
		}
	} else if instance, ok := any(s.instance).(Shutdownable); ok {
		err := instance.Shutdown()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ServiceEager[T]) clone() any {
	return s
}
