package do

// Package creates a function that executes multiple service registration functions.
// This function is used to group related service registrations into reusable packages
// that can be applied to any injector instance.
//
// Parameters:
//   - services: Variable number of service registration functions to execute
//
// Returns a function that can be passed to New(), NewWithOpts(), or Scope() to register all services.
//
// Example:
//
//		 // pkg/database/package.go
//			// Create a database package
//			DatabasePackage := do.Package(
//			    do.Lazy[*Database](func(i do.Injector) (*Database, error) {
//			        return &Database{}, nil
//			    }),
//			    do.Lazy[*ConnectionPool](func(i do.Injector) (*ConnectionPool, error) {
//			        return &ConnectionPool{}, nil
//			    }),
//	         ...
//			)
//
//		 // main.go
//			// Apply the package to an injector
//			injector := do.New(database.DatabasePackage)
func Package(services ...func(i Injector)) func(Injector) {
	return func(injector Injector) {
		for i := range services {
			services[i](injector)
		}
	}
}

// Lazy creates a function that registers a lazy service using the default service name.
// This function is a convenience wrapper for creating lazy service registration functions
// that can be used in packages.
//
// Parameters:
//   - p: The provider function that creates the service instance
//
// Returns a function that registers the service as lazy when executed.
//
// Example:
//
//	dbService := do.Lazy[*Database](func(i do.Injector) (*Database, error) {
//	    return &Database{}, nil
//	})
//
//	// Use in a package
//	package := do.Package(dbService, ...)
func Lazy[T any](p Provider[T]) func(Injector) {
	return func(injector Injector) {
		Provide(injector, p)
	}
}

// LazyNamed creates a function that registers a lazy service with a custom name.
// This function is a convenience wrapper for creating named lazy service registration functions
// that can be used in packages.
//
// Parameters:
//   - serviceName: The custom name for the service
//   - p: The provider function that creates the service instance
//
// Returns a function that registers the service as lazy with the specified name when executed.
//
// Example:
//
//	dbService := do.LazyNamed[*Database]("main-db", func(i do.Injector) (*Database, error) {
//	    return &Database{}, nil
//	})
//
//	// Use in a package
//	package := do.Package(dbService)
func LazyNamed[T any](serviceName string, p Provider[T]) func(Injector) {
	return func(injector Injector) {
		ProvideNamed(injector, serviceName, p)
	}
}

// Eager creates a function that registers an eager service using the default service name.
// This function is a convenience wrapper for creating eager service registration functions
// that can be used in packages. The service value is provided directly.
//
// Parameters:
//   - value: The service instance to register eagerly
//
// Returns a function that registers the service as eager when executed.
//
// Example:
//
//	configService := do.Eager[*Config](&Config{Port: 8080})
//
//	// Use in a package
//	package := do.Package(configService)
func Eager[T any](value T) func(Injector) {
	return func(injector Injector) {
		ProvideValue(injector, value)
	}
}

// EagerNamed creates a function that registers an eager service with a custom name.
// This function is a convenience wrapper for creating named eager service registration functions
// that can be used in packages. The service value is provided directly.
//
// Parameters:
//   - serviceName: The custom name for the service
//   - value: The service instance to register eagerly
//
// Returns a function that registers the service as eager with the specified name when executed.
//
// Example:
//
//	configService := do.EagerNamed[*Config]("app-config", &Config{Port: 8080})
//
//	// Use in a package
//	package := do.Package(configService, ...)
func EagerNamed[T any](serviceName string, value T) func(Injector) {
	return func(injector Injector) {
		ProvideNamedValue(injector, serviceName, value)
	}
}

// Transient creates a function that registers a transient service using the default service name.
// This function is a convenience wrapper for creating transient service registration functions
// that can be used in packages.
//
// Parameters:
//   - p: The provider function that creates the service instance
//
// Returns a function that registers the service as transient when executed.
//
// Example:
//
//	loggerService := do.Transient[*Logger](func(i do.Injector) (*Logger, error) {
//	    return &Logger{}, nil
//	})
//
//	// Use in a package
//	package := do.Package(loggerService, ...)
func Transient[T any](p Provider[T]) func(Injector) {
	return func(injector Injector) {
		ProvideTransient(injector, p)
	}
}

// TransientNamed creates a function that registers a transient service with a custom name.
// This function is a convenience wrapper for creating named transient service registration functions
// that can be used in packages.
//
// Parameters:
//   - serviceName: The custom name for the service
//   - p: The provider function that creates the service instance
//
// Returns a function that registers the service as transient with the specified name when executed.
//
// Example:
//
//	loggerService := do.TransientNamed[*Logger]("request-logger", func(i do.Injector) (*Logger, error) {
//	    return &Logger{}, nil
//	})
//
//	// Use in a package
//	package := do.Package(loggerService, ...)
func TransientNamed[T any](serviceName string, p Provider[T]) func(Injector) {
	return func(injector Injector) {
		ProvideNamedTransient(injector, serviceName, p)
	}
}

// Bind creates a function that creates a type alias between two types.
// This function is a convenience wrapper for creating service binding functions
// that can be used in packages.
//
// Parameters:
//   - Initial: The original type to bind from
//   - Alias: The type to bind to (must be implemented by Initial)
//
// Returns a function that creates the type alias when executed.
// Panics if the binding cannot be created.
//
// Example:
//
//	dbBinding := do.Bind[*Database, DatabaseInterface]()
//
//	// Use in a package
//	package := do.Package(dbBinding, ...)
func Bind[Initial any, Alias any]() func(Injector) {
	return func(injector Injector) {
		MustAs[Initial, Alias](injector)
	}
}

// BindNamed creates a function that creates a named type alias between two types.
// This function is a convenience wrapper for creating named service binding functions
// that can be used in packages.
//
// Parameters:
//   - initial: The name of the original service
//   - alias: The name for the alias service
//   - Initial: The original type to bind from
//   - Alias: The type to bind to (must be implemented by Initial)
//
// Returns a function that creates the named type alias when executed.
// Panics if the binding cannot be created.
//
// Example:
//
//	dbBinding := do.BindNamed[*Database, DatabaseInterface]("main-db", "db-interface")
//
//	// Use in a package
//	package := do.Package(dbBinding, ...)
func BindNamed[Initial any, Alias any](initial string, alias string) func(Injector) {
	return func(injector Injector) {
		MustAsNamed[Initial, Alias](injector, initial, alias)
	}
}
