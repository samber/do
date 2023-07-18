package do

type ServiceOpt[T any] func(Service[T])

// custom shutdown

func noopShutdownFunc[T any](instance T) error {
	return nil
}

func WithShutdownFunc[T any](shutdownFunc shutdownFunc[T]) ServiceOpt[T] {
	return func(s Service[T]) {
		if shutdownFunc == nil {
			// disable default shutdown if nil specified
			shutdownFunc = noopShutdownFunc[T]
		}
		s.setShutdownFunc(shutdownFunc)
	}
}
