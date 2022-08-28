package do

type ServiceEager[T any] struct {
	name     string
	instance T
}

func newServiceEager[T any](name string, instance T) Service[T] {
	return &ServiceEager[T]{
		name:     name,
		instance: instance,
	}
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

func (s *ServiceEager[T]) shutdown() error {
	instance, ok := any(s.instance).(Shutdownable)
	if ok {
		return instance.Shutdown()
	}

	return nil
}

func (s *ServiceEager[T]) clone() any {
	return s
}
