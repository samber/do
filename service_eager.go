package di

type ServiceEager struct {
	name     string
	instance any
}

func newServiceEager(name string, instance any) Service {
	return &ServiceEager{
		name:     name,
		instance: instance,
	}
}

//nolint:unused
func (s *ServiceEager) getName() string {
	return s.name
}

//nolint:unused
func (s *ServiceEager) getInstance(i *Injector) (any, error) {
	return s.instance, nil
}

func (s *ServiceEager) healthcheck() error {
	instance, ok := any(s.instance).(Healthcheckable)
	if ok {
		return instance.HealthCheck()
	}

	return nil
}

func (s *ServiceEager) shutdown() error {
	instance, ok := any(s.instance).(Shutdownable)
	if ok {
		return instance.Shutdown()
	}

	return nil
}

func (s *ServiceEager) clone() any {
	return s
}
